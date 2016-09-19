package main;


import (
    "os"
    "fmt"
    "path"
    "net/url"
    "strings"
)

var moduleCache = ModuleCache{Items:make(map[string]ModuleCacheItem)};
var workingPath = "";
var Options = BpmOptions {
    BpmCachePath: "bpm_modules",
    BpmFileName: "bpm.json",
    LocalModuleName: "local",
    ExcludeFileList: ".git|.gitignore|.gitmodules|bpm_modules",
}

func SliceIndex(limit int, predicate func(i int) bool) int {
    for i := 0; i < limit; i++ {
        if predicate(i) {
            return i
        }
    }
    return -1
}

func GetDependenciesLocal(bpm BpmData) error {
    depsToProcess := make(map[string]BpmDependency);
    for name, v := range bpm.Dependencies {
        depsToProcess[name] = v
    }

    bpmUpdated := false;
    for updateModule, v := range depsToProcess {
        theUrl := path.Join(Options.UseLocal, updateModule);
        fmt.Println("Processing local dependency in", theUrl)
        git := GitCommands{Path:theUrl}
        newCommit, err := git.GetLatestCommit()
        if err != nil {
            fmt.Println("Error: There was an issue getting the latest commit for", updateModule)
            return err;
        }
        itemPath := path.Join(Options.BpmCachePath, updateModule, Options.LocalModuleName);
        os.RemoveAll(itemPath)
        os.MkdirAll(itemPath, 0777)
        copyDir := CopyDir{Exclude:Options.ExcludeFileList}
        err = copyDir.Copy(theUrl, itemPath);
        if err != nil {
            fmt.Println("Error: There was an issue trying to copy the cloned temp folder to the named folder for", updateModule)
            fmt.Println(err)
            return err;
        }
        //os.RemoveAll(itemPath);

        cacheItem := ModuleCacheItem{Name:updateModule, Path: itemPath}
        moduleCache.Add(cacheItem, false, "")

        moduleBpm := BpmData{};
        moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName);
        err = moduleBpm.LoadFile(moduleBpmFilePath);
        if err != nil {
            fmt.Println("Error: Could not load the bpm.json file for dependency", updateModule)
            fmt.Println(err);
            return err
        }

        err = GetDependenciesLocal(moduleBpm)
        if err != nil {
            return err;
        }

        newItem := BpmDependency{Url: v.Url, Commit:newCommit}
        existingItem := bpm.Dependencies[updateModule];
        if !existingItem.Equal(newItem) {
            bpmUpdated = true;
        }
        bpm.Dependencies[updateModule] = newItem;
    }
    if Options.Recursive && bpmUpdated {
        filePath := path.Join(Options.UseLocal, bpm.Name, Options.BpmFileName);
        bpm.IncrementVersion();
        bpm.WriteFile(filePath)
    }
    return nil
}

func PathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    fmt.Println("Error:", err)
    return true
}

func GetDependencies(itemName string, item BpmDependency, parentUrl string) error {
    if item.Url == "" {
        fmt.Println("Error: No url specified for " + itemName)
    }
    if item.Commit == "" {
        fmt.Println("Error: No commit specified for " + itemName)
    }
    fmt.Println("Processing dependency", itemName)
    workingPath,_ := os.Getwd();
    git := GitCommands{Path:workingPath}

    itemPath := path.Join(Options.BpmCachePath, itemName)
    os.Mkdir(itemPath, 0777)

    itemRemoteUrl := item.Url;
    itemClonePath := path.Join(workingPath, itemPath, item.Commit)
    localPath := path.Join(Options.BpmCachePath, itemName, Options.LocalModuleName)
    if PathExists(localPath) {
        fmt.Println("Found local folder in the bpm modules. Using this folder", localPath)
        itemClonePath = localPath;
    } else if !PathExists(itemClonePath) {
        fmt.Println("Could not find module", itemName, "in the bpm cache. Cloning repository...")
        // If the parent URL is unspecified then use the remote URL as parent.
        if parentUrl == "" {
            tempUrl, err := git.GetRemoteUrl(Options.UseRemote)
            if err != nil {
                fmt.Println("Error: There was a problem getting the remote url", Options.UseRemote)
                return err
            }
            parentUrl = tempUrl
        } else if strings.Index(parentUrl, "http") != 0 {
            tempUrl, err := git.GetRemoteUrl(Options.UseRemote)
            if err != nil {
                fmt.Println("Error: There was a problem getting the remote url", Options.UseRemote)
                return err
            }
            parsedTmpUrl, err := url.Parse(tempUrl)
            if err != nil {
                fmt.Println(err);
                return err
            }
            parentUrl = parsedTmpUrl.Scheme + "://" + path.Join(parsedTmpUrl.Host, parsedTmpUrl.Path, parentUrl)
        }

        tempUrl, err := url.Parse(parentUrl)
        if err != nil {
            fmt.Println(err);
            return err
        }
        // If the item URL is a relative URL, then make a full URL using the parent url as the root.
        if strings.Index(item.Url, "http") != 0 {
            itemRemoteUrl = tempUrl.Scheme + "://" + path.Join(tempUrl.Host, tempUrl.Path, item.Url)
        }
        os.Mkdir(itemClonePath, 0777)
        git := GitCommands{Path: itemClonePath}
        err = git.InitAndCheckoutCommit(itemRemoteUrl, item.Commit)
        if err != nil {
            fmt.Println("Error: There was an issue initializing the repository for dependency", itemName)
            os.RemoveAll(itemClonePath)
            os.Exit(1)
            return err;
        }
    } else {
        fmt.Println("Module", itemName, "already exists in the bpm cache.")
    }
    fmt.Println("Done with", itemName)

    // Recursively get dependencies in the current dependency
    moduleBpm := BpmData{};
    moduleBpmFilePath := path.Join(itemClonePath, Options.BpmFileName)
    err := moduleBpm.LoadFile(moduleBpmFilePath);
    if err != nil {
        fmt.Println("Error: Could not load the bpm.json file for dependency", itemName)
        fmt.Println(err);
        return err
    }

    if strings.TrimSpace(moduleBpm.Name) == "" {
        fmt.Println("Error: There must be a name field in the bpm.json for", itemName)
        os.Exit(1);
    }

    if strings.TrimSpace(moduleBpm.Version) == "" {
        fmt.Println("Error: There must be a version field in the bpm for", itemName)
        os.Exit(1);
    }
    cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpm.Version, Commit: item.Commit, Path: itemClonePath}
    fmt.Println("Adding to cache", cacheItem.Name)
    moduleCache.Add(cacheItem, true, itemClonePath)

    fmt.Println("Processing all dependencies for", moduleBpm.Name, "version", moduleBpm.Version);
    for name, v := range moduleBpm.Dependencies {
        // Do not process a dependency which appears to be itself.
        if strings.Compare(strings.ToLower(name), strings.ToLower(itemName)) != 0 {
            err := GetDependencies(name, v, itemRemoteUrl)
            if err != nil {
                return err;
            }
        } else {
            fmt.Println("Ignoring self dependency for", name)
        }
    }
    return nil;
}

func GetSubCommand() (SubCommand){
    if len(os.Args) > 1 {
        command := strings.ToLower(os.Args[1]);
        if command == "init" {
            return &InitCommand{}
        }
        if command == "ls" {
            return &LsCommand{}
        }
        if command == "help" {
            return &HelpCommand{};
        }
        if command == "clean" {
            return &CleanCommand{}
        }
        if command == "update" {
            return &UpdateCommand{}
        }
        if command == "install" {
            return &InstallCommand{}
        }
        if command == "version" {
            return &VersionCommand{}
        }
        fmt.Println("Unrecognized command", command)
    }
    return &HelpCommand{};
}

func main() {
    workingPath,_ = os.Getwd();

    Options.Recursive = GetRecursiveFlag(os.Args);
    Options.UseRemote = GetRemote(os.Args);
    Options.UseLocal = GetLocal(os.Args);
    Options.SkipNpmInstall = GetSkipNpmInstall(os.Args);
    Options.ConflictResolutionType = GetConflictResolutionType(os.Args)
    cmd := GetSubCommand();
    err := cmd.Execute();
    if err != nil {
        os.Exit(1)
    }
    os.Exit(0)
}
