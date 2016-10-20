package main;


import (
    "os"
    "fmt"
    "path"
    "net/url"
    "strings"
    "github.com/blang/semver"
    "bpmerror"
)

var moduleCache = ModuleCache{Items:make(map[string]*ModuleCacheItem)};
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

func PathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    fmt.Println("Error:", err)
    return true
}

func ProcessLocalModule(source string) (*BpmData, *ModuleCacheItem, error) {
    moduleBpm := &BpmData{};
    moduleBpmFilePath := path.Join(source, Options.BpmFileName);
    err := moduleBpm.LoadFile(moduleBpmFilePath);
    if err != nil {
        return nil, nil, bpmerror.New(err, "Error: Could not load the bpm.json file for dependency " + source)
    }
    itemPath := path.Join(Options.BpmCachePath, moduleBpm.Name, Options.LocalModuleName);
    // Clean out the destination directory and then copy the files from the source directory to the final location in the bpm cache.
    os.RemoveAll(itemPath)
    os.MkdirAll(itemPath, 0777)
    copyDir := CopyDir{Exclude:Options.ExcludeFileList}
    err = copyDir.Copy(source, itemPath);
    if err != nil {
        return nil, nil, bpmerror.New(err, "Error: There was an issue trying to copy the local folder to the bpm_cache for " + source)
    }
    cacheItem := &ModuleCacheItem{Name:moduleBpm.Name, Path: itemPath}
    return moduleBpm, cacheItem, nil
}

func ProcessRemoteModule(itemRemoteUrl string, moduleCommit string) (*BpmData, *ModuleCacheItem, error) {
    itemPathTemp := path.Join(workingPath, Options.BpmCachePath, "xx_temp_xx", "xx_temp_xx")
    defer os.RemoveAll(path.Join(itemPathTemp, ".."))
    os.RemoveAll(path.Join(itemPathTemp));
    os.MkdirAll(itemPathTemp, 0777)
    moduleBpm := &BpmData{};
    git := GitCommands{Path:itemPathTemp}
    err := git.InitAndCheckout(itemRemoteUrl, moduleCommit)
    if err != nil {
        return nil, nil, bpmerror.New(err, "Error: There was an issue initializing the repository for dependency " + moduleBpm.Name + " Url: " + itemRemoteUrl + " Commit: " + moduleCommit)
    }
    moduleBpmFilePath := path.Join(itemPathTemp, Options.BpmFileName);
    err = moduleBpm.LoadFile(moduleBpmFilePath);
    if err != nil {
        return nil, nil, bpmerror.New(err, "Error: Could not load the bpm.json file for dependency at " + moduleBpmFilePath)
    }

    if moduleCommit == "master" {
        moduleCommit, err = git.GetLatestCommit()
        if err != nil {
            return nil, nil, err;
        }
    }

    itemPath := path.Join(Options.BpmCachePath, moduleBpm.Name, moduleCommit);
    // Clean out the destination directory and then copy the files from the temp directory to the final location in the bpm cache.
    os.RemoveAll(itemPath)
    copyDir := CopyDir{/*Exclude:Options.ExcludeFileList*/}
    err = copyDir.Copy(itemPathTemp, itemPath);
    if err != nil {
        return nil, nil, err;
    }

    moduleBpmVersion, err := semver.Make(moduleBpm.Version);
    if err != nil {
        fmt.Println("Could not read the version");
        return nil, nil, err;
    }
    cacheItem := &ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: moduleCommit, Path: itemPath}
    return moduleBpm, cacheItem, nil;
}

func DetermineCommitValue(source string) (string, error) {
    // Put the commit as local, since we don't really know what it should be.
    // It's expected that the bpm install will fail if the commit is 'local' which will indicate to the user that they
    // didn't finalize configuring the dependency
    // However if the repo has no changes, then grab the commit
    commit := "local"
    git := GitCommands{Path:source}
    if Options.Command.Name() == "update" && (Options.Finalize || !git.HasChanges()) {
        var err error;
        commit, err = git.GetLatestCommit()
        if err != nil {
            return "", err;
        }
    }
    return commit, nil
}

func MakeRemoteUrl(itemUrl string) (string, error) {
    adjustedUrl := itemUrl;
    if adjustedUrl == "" {
        git := GitCommands{Path:workingPath}
        var err error;
        adjustedUrl, err = git.GetRemoteUrl(Options.UseRemote)
        if err != nil {
            return "", bpmerror.New(err, "Error: There was a problem getting the remote url " + Options.UseRemote)
        }
    } else if strings.Index(adjustedUrl, "http") != 0 {
        git := GitCommands{Path:workingPath}
        remoteUrl, err := git.GetRemoteUrl(Options.UseRemote)
        if err != nil {
            return "", bpmerror.New(err, "Error: There was a problem getting the remote url " + Options.UseRemote)
        }
        parsedUrl, err := url.Parse(remoteUrl)
        if err != nil {
            return "", bpmerror.New(err, "Error: There was a problem parsing the remote url " + remoteUrl)
        }
        adjustedUrl = parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, parsedUrl.Path, itemUrl)
    }
    return adjustedUrl, nil;
}

func ProcessDependencies(bpm *BpmData, parentUrl string) (error) {
    for itemName, item := range bpm.Dependencies {
        fmt.Println("Validating dependency", itemName)
        err := item.Validate();
        if err != nil {
            return err;
        }
        if Options.UseLocal != "" && strings.Index(item.Url, "http") == -1 {
            bpmUpdated := false
            moduleSourceUrl := path.Join(Options.UseLocal, itemName);
            fmt.Println("Processing local dependency in", moduleSourceUrl)
            commit, err := DetermineCommitValue(moduleSourceUrl)
            if err != nil {
                return bpmerror.New(err, "Error: There was an issue getting the latest commit for " + itemName)
            }
            moduleBpm, cacheItem, err := ProcessLocalModule(moduleSourceUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            err = ProcessDependencies(moduleBpm, "")
            if err != nil {
                return err;
            }

            newItem := BpmDependency{Url: item.Url, Commit:commit}
            existingItem := bpm.Dependencies[itemName];
            if !existingItem.Equal(newItem) {
                bpmUpdated = true;
            }
            bpm.Dependencies[itemName] = newItem;
            if Options.Command.Name() == "update" && Options.Recursive && bpmUpdated {
                filePath := path.Join(Options.UseLocal, bpm.Name, Options.BpmFileName);
                err = bpm.IncrementVersion();
                if err != nil {
                    return err;
                }
                err = bpm.WriteFile(filePath)
                if err != nil {
                    return err;
                }
            }
        } else {
            fmt.Println("Processing dependency", itemName)

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
                var err error;
                parentUrl, err = MakeRemoteUrl(parentUrl);
                if err != nil {
                    return err;
                }
                tempUrl, err := url.Parse(parentUrl)
                if err != nil {
                    return bpmerror.New(err, "Error: There is something wrong with the module url " + parentUrl)
                }
                // If the item URL is a relative URL, then make a full URL using the parent url as the root.
                if strings.Index(item.Url, "http") != 0 {
                    itemRemoteUrl = tempUrl.Scheme + "://" + path.Join(tempUrl.Host, tempUrl.Path, item.Url)
                }
                os.Mkdir(itemClonePath, 0777)
                git := GitCommands{Path: itemClonePath}
                err = git.InitAndCheckout(itemRemoteUrl, item.Commit)
                if err != nil {
                    os.RemoveAll(itemClonePath)
                    return bpmerror.New(nil, "Error: There was an issue initializing the repository for dependency " + itemName + " Url: " + itemRemoteUrl + " Commit: " + item.Commit)
                }
            } else {
                fmt.Println("Module", itemName, "already exists in the bpm cache.")
            }
            // Recursively get dependencies in the current dependency
            moduleBpm := &BpmData{};
            moduleBpmFilePath := path.Join(itemClonePath, Options.BpmFileName)
            err := moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                return bpmerror.New(err, "Error: Could not load the bpm.json file for dependency " + itemName)
            }

            fmt.Println("Validating bpm.json for", itemName)
            err = moduleBpm.Validate();
            if err != nil {
                return err;
            }
            cacheItem := &ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpm.Version, Commit: item.Commit, Path: itemClonePath}
            fmt.Println("Adding to cache", cacheItem.Name)
            moduleCache.AddLatest(cacheItem)

            fmt.Println("Processing all dependencies for", moduleBpm.Name, "version", moduleBpm.Version);
            err = ProcessDependencies(moduleBpm, itemRemoteUrl)
            if err != nil {
                return err;
            }
        }
    }
    return nil;
}

func main() {
    workingPath,_ = os.Getwd();
    Options.ParseOptions(os.Args);
    err := Options.Command.Execute();
    if err != nil {
        fmt.Println(err)
        fmt.Println("Finished with errors")
        os.Exit(1)
    }
    fmt.Println("Finished")
    os.Exit(0)
}
