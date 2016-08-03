package main;


import (
    "os"
    "io/ioutil"
    "encoding/json"
    "fmt"
    "path"
    "net/url"
    "strings"
    "github.com/blang/semver"
)

const bpmCachePath = "bpm_modules"
const bpmFileName = "bpm.json"
var useRemote = ""
var useLocal = false
var localPath = ""
var moduleCache = ModuleCache{Items:make(map[string]ModuleCacheItem)};
const excludeFileList = ".git|.gitignore|.gitmodules|" + bpmCachePath
const localModuleName = "local"

func GetDependenciesLocal(itemName string, itemPath string) error {
    theUrl := path.Join(localPath, itemName);
    fmt.Println("Processing local dependency in", theUrl)
    os.RemoveAll(itemPath)
    os.MkdirAll(itemPath, 0777)
    copyDir := CopyDir{Exclude:excludeFileList}
    err := copyDir.Copy(theUrl, itemPath);
    if err != nil {
        fmt.Println(err)
        return err;
    }

    cacheItem := ModuleCacheItem{Name:itemName, Path: itemPath}
    moduleCache.Add(cacheItem, false)
    moduleBpm := BpmData{};
    moduleBpmFilePath := path.Join(itemPath, bpmFileName)
    err = moduleBpm.LoadFile(moduleBpmFilePath);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", moduleBpmFilePath, "for dependency", itemName)
        fmt.Println(err);
        return err
    }
    for depName, _ := range moduleBpm.Dependencies {
        depPath := path.Join(bpmCachePath, depName, localModuleName)
        err = GetDependenciesLocal(depName, depPath)
        if err != nil {
            return err;
        }
    }
    return nil;
}

func PathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    fmt.Println("Error:", err)
    return true
}

func GetDependencies(itemName string, item BpmDependency) error {
    if item.Url == "" {
        fmt.Println("Error: No url specified for " + itemName)
    }
    if item.Commit == "" {
        fmt.Println("Error: No commit specified for " + itemName)
    }
    fmt.Println("Processing dependency", itemName)
    workingPath,_ := os.Getwd();
    git := GitCommands{Path:workingPath}

    itemPath := path.Join(bpmCachePath, itemName)
    os.Mkdir(itemPath, 0777)

    itemClonePath := path.Join(workingPath, itemPath, item.Commit)
    localPath := path.Join(bpmCachePath, itemName, localModuleName)
    if PathExists(localPath) {
        fmt.Println("Found local folder in the bpm modules. Using this folder", localPath)
        itemClonePath = localPath;
    } else if !PathExists(itemClonePath) {
        theUrl, err := git.GetRemoteUrl(useRemote)
        if err != nil {
            fmt.Println("Error: There was a problem getting the remote url", useRemote)
            return err
        }
        remoteUrl, err := url.Parse(theUrl)
        if err != nil {
            fmt.Println(err);
            return err
        }
        itemRemoteUrl := item.Url;
        if strings.Index(item.Url, "http") != 0 {
            itemRemoteUrl = remoteUrl.Scheme + "://" + path.Join(remoteUrl.Host, remoteUrl.Path, item.Url)
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
        fmt.Println("Checkout for", itemName, "already exists.")
    }
    fmt.Println("Done with " + itemName)

    // Recursively get dependencies in the current dependency
    moduleBpm := BpmData{};
    moduleBpmFilePath := path.Join(itemClonePath, bpmFileName)
    err := moduleBpm.LoadFile(moduleBpmFilePath);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", moduleBpmFilePath, "for dependency", itemName)
        fmt.Println(err);
        return err
    }

    if strings.TrimSpace(moduleBpm.Name) == "" {
        fmt.Println("There must be a name field in the bpm. File:", moduleBpmFilePath)
        os.Exit(1);
    }

    if strings.TrimSpace(moduleBpm.Version) == "" {
        fmt.Println("There must be a version field in the bpm.", moduleBpmFilePath)
        os.Exit(1);
    }
    cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpm.Version, Commit: item.Commit, Path: itemClonePath}
    fmt.Println("Adding to cache", cacheItem.Name)
    moduleCache.Add(cacheItem, true)

    fmt.Println("Processing all dependencies for", moduleBpm.Name, "version", moduleBpm.Version);
    for name, v := range moduleBpm.Dependencies {
        // Do not process a dependency which appears to be itself.
        if strings.Compare(strings.ToLower(name), strings.ToLower(itemName)) != 0 {
            err := GetDependencies(name, v)
            if err != nil {
                return err;
            }
        } else {
            fmt.Println("Ignoring self dependency for", name)
        }
    }
    return nil;
}

func main() {
    action := "build"
    newUrl := "";
    newCommit := "";
    updateModule := "";
    newBpmName := ""

    for i, arg := range os.Args {

        if strings.Index(strings.ToLower(arg), "--new=") == 0 {
            action = "new"
            tempx := strings.Split(arg, "=")
            if len(tempx) == 2 {
                newBpmName = tempx[1];
            }
        }

        if strings.Compare(strings.ToLower(arg), "--help") == 0 {
            fmt.Println("This comment is not helpful.")
            return;
        }

        if strings.Compare(strings.ToLower(arg), "--clean") == 0 {
            os.RemoveAll(bpmCachePath);
            os.Exit(1)
        }

        if strings.Compare(strings.ToLower(arg), "install") == 0 {

            action = "install"

            if len(os.Args) <= i + 1 {
                fmt.Println("The install command requires a url");
                os.Exit(1);
            }
            newUrl = os.Args[i + 1];
            if len(os.Args) > i + 2 && strings.Index(os.Args[i+2], "--") != 0 {
                newCommit = os.Args[ i + 2];
            }
        }
        if strings.Index(strings.ToLower(arg), "--remote=") == 0 {
            if useLocal {
                fmt.Println("cannot use --remote with --root")
                return;
            }

            tempx := strings.Split(arg, "=")
            if len(tempx) == 2 {
                useRemote = tempx[1];
            }
        }

        if strings.Index(strings.ToLower(arg), "--root=") == 0 {
            if len(useRemote) > 0 {
                fmt.Println("cannot use --root with --remote")
                return;
            }

            tempx := strings.Split(arg, "=")
            if len(tempx) == 2 {
                useLocal = true;
                localPath = tempx[1]
            }
        }

        if strings.Compare(strings.ToLower(arg), "update") == 0 {
            action = "update"

            if len(os.Args) >= i + 1 && strings.Index(os.Args[i + 1], "--") != 0 {
                updateModule = os.Args[i + 1];
            }
        }
    }
    workingPath,_ := os.Getwd();

    if strings.Compare(action, "new") == 0 {
        bpm := BpmData{Name:newBpmName, Version:"1.0.0", Dependencies:make(map[string]BpmDependency)};
        bytes, _ := json.MarshalIndent(bpm, "", "   ")
        err := ioutil.WriteFile(path.Join(workingPath, bpmFileName), bytes, 0666);
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        fmt.Println(string(bytes))
        os.Exit(0)
    }

    if len(useRemote) == 0 {
        useRemote = "origin"
    }

    if useLocal {
        if strings.Index(localPath, ".") == 0 || strings.Index(localPath, "..") == 0 {
            localPath = path.Join(workingPath, localPath)
        }
    }

    if strings.Compare(action, "update") == 0 {
        if _, err := os.Stat(bpmFileName); os.IsNotExist(err) {
            fmt.Println("The bpm.json file does not exist.")
            os.Exit(1)
        }

        bpm := BpmData{};
        err := bpm.LoadFile(bpmFileName);
        if err != nil {
            fmt.Println("Error: There was a problem loading the bpm file at", bpmFileName)
            fmt.Println(err);
            os.Exit(1)
        }
        if !bpm.HasDependencies() {
            fmt.Println("There are no dependencies. Please use the install command first.")
            os.Exit(1)
        }

        depsToProcess := make(map[string]BpmDependency);

        if updateModule == "" {
            for name, v := range bpm.Dependencies {
                depsToProcess[name] = v
            }
        } else {
            depItem, exists := bpm.Dependencies[updateModule];
            if !exists {
                fmt.Println("Could not find module", updateModule, "in the dependencies")
                os.Exit(1)
            }
            depsToProcess[updateModule] = depItem;
        }

        for updateModule, depItem := range depsToProcess {
            if useLocal {
                theUrl := path.Join(localPath, updateModule);
                itemPath := path.Join(bpmCachePath, updateModule, "xx_temp_xx")
                os.RemoveAll(itemPath)
                os.MkdirAll(itemPath, 0777)
                copyDir := CopyDir{Exclude:excludeFileList}
                err := copyDir.Copy(theUrl, itemPath);
                if err != nil {
                    fmt.Println(err)
                    return;
                }

                cacheItem := ModuleCacheItem{Name:updateModule, Path: itemPath}
                moduleCache.Add(cacheItem, false)

                moduleBpm := BpmData{};
                moduleBpmFilePath := path.Join(itemPath, bpmFileName);
                err = moduleBpm.LoadFile(moduleBpmFilePath);
                if err != nil {
                    fmt.Println("Error: There was a problem loading the bpm file at", moduleBpmFilePath, "for dependency", updateModule)
                    fmt.Println(err);
                    return
                }
                for depName, _ := range moduleBpm.Dependencies {
                    depPath := path.Join(bpmCachePath, depName, localModuleName)
                    GetDependenciesLocal(depName, depPath)
                }
                err = moduleCache.CopyAndNpmInstall(path.Join(workingPath, "node_modules"))
                if err != nil {
                    os.Exit(1)
                }
                return;
            }

            git := GitCommands{Path:workingPath}
            theUrl, err := git.GetRemoteUrl(useRemote)
            if err != nil {
                fmt.Println("There was a problem getting the remote url", useRemote)
                os.Exit(1)
            }
            remoteUrl, err := url.Parse(theUrl)
            if err != nil {
                fmt.Println("There was a problem parsing the remote url", theUrl)
                fmt.Println(err);
                os.Exit(1)
            }
            itemRemoteUrl := depItem.Url;
            if strings.Index(depItem.Url, "http") != 0 {
                itemRemoteUrl = remoteUrl.Scheme + "://" + path.Join(remoteUrl.Host, remoteUrl.Path, depItem.Url)
            }

            itemPath := path.Join(bpmCachePath, updateModule, "xx_temp_xx")
            os.RemoveAll(itemPath)
            os.MkdirAll(itemPath, 0777)
            git = GitCommands{Path:itemPath}
            err = git.InitAndFetch(itemRemoteUrl)
            if err != nil {
                fmt.Println(err)
                return;
            }
            err = git.Checkout("master")
            if err != nil {
                fmt.Println("Error: There was an issue retrieving the repository.")
                os.Exit(1)
                return;
            }

            err = git.SubmoduleUpdate(true, true)
            if err != nil {
                fmt.Println("Error: Could not update submodules.")
                os.Exit(1)
                return;
            }

            newCommit, err = git.GetLatestCommit()
            if err != nil {
                fmt.Println("Error: There was an issue getting the latest commit")
                os.Exit(1)
                return;
            }

            moduleBpm := BpmData{};
            moduleBpmFilePath := path.Join(itemPath, bpmFileName);
            err = moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                fmt.Println("Error: There was a problem loading the bpm file at", moduleBpmFilePath, "for dependency", updateModule)
                fmt.Println(err);
                os.Exit(1)
                return
            }
            itemNewPath := path.Join(bpmCachePath, updateModule, newCommit);
            os.RemoveAll(itemNewPath)
            copyDir := CopyDir{Exclude:excludeFileList}
            err = copyDir.Copy(itemPath, itemNewPath);
            if err != nil {
                fmt.Println(err)
                return;
            }
            os.RemoveAll(itemPath);

            moduleBpmVersion, err := semver.Make(moduleBpm.Version);
            if err != nil {
                fmt.Println("Could not read the version");
                return;
            }

            cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: newCommit, Path: itemNewPath}
            moduleCache.Add(cacheItem, true)
            for depName, v := range moduleBpm.Dependencies {
                GetDependencies(depName, v)
            }
            newItem := BpmDependency{Url: depItem.Url, Commit:newCommit}
            bpm.Dependencies[updateModule] = newItem;
        }
        err = moduleCache.CopyAndNpmInstall(path.Join(workingPath, "node_modules"))
        if err != nil {
            os.Exit(1)
        }
        //os.RemoveAll(path.Join(itemNewPath, ".."));
        bpmVersion, err := semver.Make(bpm.Version);
        if err != nil {
            fmt.Println("Could not read the version");
            return;
        }
        bpmVersion.Patch++;
        bpm.Version = bpmVersion.String();
        bytes, err := json.MarshalIndent(bpm, "", "   ")
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        err = ioutil.WriteFile(path.Join(workingPath, bpmFileName), bytes, 0666);
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        return;
    }

    if strings.Compare(action, "install") == 0 {
        bpm := BpmData{};
        if _, err := os.Stat(bpmFileName); os.IsNotExist(err) {
            fmt.Println("Error: The bpm file does not exist. Please use the new command to create the file");
            os.Exit(1)
        } else {
            err := bpm.LoadFile(bpmFileName);
            if err != nil {
                fmt.Println("Error: There was a problem loading the bpm file at", bpmFileName)
                fmt.Println(err);
                os.Exit(1)
            }
        }

        git := GitCommands{Path:workingPath}
        tmpUrl, err := git.GetRemoteUrl(useRemote);
        if err != nil {
            fmt.Println("Error: There was a problem getting the remote url", useRemote)
            return;
        }

        remoteUrl, err := url.Parse(tmpUrl)
        if err != nil {
            fmt.Println(err);
            return
        }
        itemRemoteUrl := newUrl;
        if strings.Index(newUrl, "http") != 0 {
            itemRemoteUrl = remoteUrl.Scheme + "://" + path.Join(remoteUrl.Host, remoteUrl.Path, newUrl)
        }

        moduleBpm := BpmData{};

        if len(strings.TrimSpace(newCommit)) == 0 {
            itemPath := path.Join(bpmCachePath, "xx_temp_xx", "xx_temp_xx")
            os.RemoveAll(path.Join(itemPath, ".."));
            os.MkdirAll(itemPath, 0777)
            git = GitCommands{Path:itemPath}
            err = git.InitAndFetch(itemRemoteUrl)
            if err != nil {
                fmt.Println(err)
                return;
            }
            err = git.Checkout("master")
            if err != nil {
                fmt.Println("Error: There was an issue retrieving the repository.")
                os.Exit(1)
                return;
            }
            newCommit, err = git.GetLatestCommit()
            if err != nil {
                fmt.Println(err)
                return;
            }

            moduleBpmFilePath := path.Join(itemPath, bpmFileName);
            err = moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                fmt.Println("Error: There was a problem loading the bpm file at", moduleBpmFilePath, "for dependency", updateModule)
                fmt.Println(err);
                return
            }
            moduleBpmVersion, err := semver.Make(moduleBpm.Version);
            if err != nil {
                fmt.Println("Error: Could not read the version");
                return;
            }

            itemNewPath := path.Join(bpmCachePath, moduleBpm.Name, newCommit);
            os.RemoveAll(itemNewPath)
            copyDir := CopyDir{Exclude:excludeFileList}
            err = copyDir.Copy(itemPath, itemNewPath);
            if err != nil {
                fmt.Println(err)
                return;
            }
            os.RemoveAll(path.Join(itemPath, ".."))

            cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: newCommit, Path: itemNewPath}
            moduleCache.Add(cacheItem, true)
        } else {
            itemPath := path.Join(bpmCachePath, "xx_temp_xx", newCommit)
            itemClonePath := path.Join(workingPath, itemPath, newCommit)
            os.RemoveAll(path.Join(itemClonePath, ".."));
            os.MkdirAll(itemClonePath, 0777)
            git = GitCommands{Path:itemClonePath}
            err = git.InitAndCheckoutCommit(itemRemoteUrl, newCommit)
            if err != nil {
                fmt.Println("Error: There was an issue initializing the repository for dependency", moduleBpm.Name)
                os.RemoveAll(itemClonePath)
                os.Exit(1)
                return;
            }

            moduleBpmFilePath := path.Join(itemPath, bpmFileName);
            err = moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                fmt.Println("Error: There was a problem loading the bpm file at", moduleBpmFilePath, "for dependency", updateModule)
                fmt.Println(err);
                return
            }

            moduleBpmVersion, err := semver.Make(moduleBpm.Version);
            if err != nil {
                fmt.Println("Could not read the version");
                return;
            }

            itemNewPath := path.Join(bpmCachePath, moduleBpm.Name, newCommit);
            copyDir := CopyDir{Exclude:excludeFileList}
            err = copyDir.Copy(itemPath, itemNewPath);
            if err != nil {
                fmt.Println(err)
                return;
            }
            os.RemoveAll(path.Join(itemPath, ".."))

            cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: newCommit, Path: itemNewPath}
            moduleCache.Add(cacheItem, true)
        }
        for depName, v := range moduleBpm.Dependencies {
            GetDependencies(depName, v)
        }
        err = moduleCache.CopyAndNpmInstall(path.Join(workingPath, "node_modules"))
        if err != nil {
            os.Exit(1)
        }
        newItem := BpmDependency{Url:newUrl, Commit:newCommit};
        bpm.Dependencies[moduleBpm.Name] = newItem;

        bpmVersion, err := semver.Make(bpm.Version);
        if err != nil {
            fmt.Println("Error: Could not read the version");
            return;
        }
        bpmVersion.Patch++;

        bpm.Version = bpmVersion.String()
        bytes, err := json.MarshalIndent(bpm, "", "   ")
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        err = ioutil.WriteFile(path.Join(workingPath, bpmFileName), bytes, 0666);
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        return;
    }

    if _, err := os.Stat(bpmFileName); os.IsNotExist(err) {
        fmt.Print(bpmFileName, "Error: The bpm file does not exist in the current directory.");
        os.Exit(1);
    }

    fmt.Println("Reading",bpmFileName,"...")
    bpm := BpmData{}
    err := bpm.LoadFile(bpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", bpmFileName)
        fmt.Println(err);
        os.Exit(1)
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        os.Exit(0)
    }

    if _, err := os.Stat(bpmCachePath); os.IsNotExist(err) {
        os.Mkdir(bpmCachePath, 0777)
    }

    if strings.TrimSpace(bpm.Name) == "" {
        fmt.Println("Error: There must be a name field in the bpm")
        os.Exit(1);
    }

    if strings.TrimSpace(bpm.Version) == "" {
        fmt.Println("Error: There must be a version field in the bpm")
        os.Exit(1);
    }

    fmt.Println("Processing all dependencies for", bpm.Name, "version", bpm.Version);

    for depName, v := range bpm.Dependencies {
        if useLocal {
            depPath := path.Join(bpmCachePath, depName, localModuleName)
            GetDependenciesLocal(depName, depPath)
        } else {
            err := GetDependencies(depName, v)
            if err != nil {
                fmt.Println("Error: There was an issue processing dependency", depName)
                fmt.Println(err)
                os.Exit(1)
            }
        }
    }
    err = moduleCache.CopyAndNpmInstall(path.Join(workingPath, "node_modules"))
    if err != nil {
        os.Exit(1)
    }

    // TODO
    // remove an empty node_modules folder?

    return;
}
