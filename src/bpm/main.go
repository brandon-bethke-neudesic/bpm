package main;


import (
    "os"
    "fmt"
    "path"
    "net/url"
    "strings"
    "errors"
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

func PathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    fmt.Println("Error:", err)
    return true
}

func MakeRemoteUrl(itemUrl string) (string, error) {
    adjustedUrl := itemUrl;
    if adjustedUrl == "" {
        git := GitCommands{Path:workingPath}
        var err error;
        adjustedUrl, err = git.GetRemoteUrl(Options.UseRemote)
        if err != nil {
            fmt.Println("Error: There was a problem getting the remote url", Options.UseRemote)
            return "", err
        }
    } else if strings.Index(adjustedUrl, "http") != 0 {
        git := GitCommands{Path:workingPath}
        remoteUrl, err := git.GetRemoteUrl(Options.UseRemote)
        if err != nil {
            fmt.Println("Error: There was a problem getting the remote url", Options.UseRemote)
            return "", err;
        }
        parsedUrl, err := url.Parse(remoteUrl)
        if err != nil {
            fmt.Println("Error: There was a problem parsing the remote url", remoteUrl)
            fmt.Println(err);
            return "", err;
        }
        adjustedUrl = parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, parsedUrl.Path, itemUrl)
    }
    return adjustedUrl, nil;
}

func ProcessDependencies(bpm BpmData, parentUrl string) (error) {
    for itemName, item := range bpm.Dependencies {

        if Options.UseLocal != "" && strings.Index(item.Url, "http") == -1 {
            bpmUpdated := false
            theUrl := path.Join(Options.UseLocal, itemName);
            fmt.Println("Processing local dependency in", theUrl)
            // Put the commit as local, since we don't really know what it should be.
            // It's expected that the bpm install will fail if the commit is 'local' which will indicate to the user that they
            // didn't finalize configuring the dependency
            // However if the repo has no changes, then grab the commit
            newCommit := "local"
            git := GitCommands{Path:theUrl}
            if Options.Command.Name() == "update" && (Options.Finalize || !git.HasChanges()) {
                var err error;
                newCommit, err = git.GetLatestCommit()
                if err != nil {
                    fmt.Println("Error: There was an issue getting the latest commit for", itemName)
                    fmt.Println(err)
                    return err;
                }
            }
            itemPath := path.Join(Options.BpmCachePath, itemName, Options.LocalModuleName);
            os.RemoveAll(itemPath)
            os.MkdirAll(itemPath, 0777)
            copyDir := CopyDir{Exclude:Options.ExcludeFileList}
            err := copyDir.Copy(theUrl, itemPath);
            if err != nil {
                fmt.Println("Error: There was an issue trying to copy the cloned temp folder to the named folder for", itemName)
                fmt.Println(err)
                return err;
            }
            //os.RemoveAll(itemPath);

            cacheItem := ModuleCacheItem{Name:itemName, Path: itemPath}
            moduleCache.Add(cacheItem, false, "")

            moduleBpm := BpmData{};
            moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName);
            err = moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                fmt.Println("Error: Could not load the bpm.json file for dependency", itemName)
                fmt.Println(err);
                return err
            }

            err = ProcessDependencies(moduleBpm, "")
            if err != nil {
                return err;
            }

            newItem := BpmDependency{Url: item.Url, Commit:newCommit}
            existingItem := bpm.Dependencies[itemName];
            if !existingItem.Equal(newItem) {
                bpmUpdated = true;
            }
            bpm.Dependencies[itemName] = newItem;
            if Options.Command.Name() == "update" && Options.Recursive && bpmUpdated {
                filePath := path.Join(Options.UseLocal, bpm.Name, Options.BpmFileName);
                bpm.IncrementVersion();
                bpm.WriteFile(filePath)
            }
        } else {
            if item.Url == "" {
                fmt.Println("Error: No url specified for " + itemName)
            }
            if item.Commit == "" {
                fmt.Println("Error: No commit specified for " + itemName)
            }
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
                tempUrl, err := url.Parse(parentUrl)
                if err != nil {
                    fmt.Println("Error: There is something wrong with the module url", parentUrl)
                    fmt.Println(err);
                    return err;
                }
                // If the item URL is a relative URL, then make a full URL using the parent url as the root.
                if strings.Index(item.Url, "http") != 0 {
                    itemRemoteUrl = tempUrl.Scheme + "://" + path.Join(tempUrl.Host, tempUrl.Path, item.Url)
                }
                os.Mkdir(itemClonePath, 0777)
                git := GitCommands{Path: itemClonePath}
                err = git.InitAndCheckout(itemRemoteUrl, item.Commit)
                if err != nil {
                    msg := "Error: There was an issue initializing the repository for dependency " + itemName + " Url: " + itemRemoteUrl + " Commit: " + item.Commit
                    os.RemoveAll(itemClonePath)
                    return errors.New(msg)
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
                return err;
            }

            if strings.TrimSpace(moduleBpm.Name) == "" {
                msg := "Error: There must be a name field in the bpm.json for " + itemName
                return errors.New(msg)
            }

            if strings.TrimSpace(moduleBpm.Version) == "" {
                msg := "Error: There must be a version field in the bpm for " + itemName
                return errors.New(msg)
            }
            cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpm.Version, Commit: item.Commit, Path: itemClonePath}
            fmt.Println("Adding to cache", cacheItem.Name)
            moduleCache.Add(cacheItem, true, itemClonePath)

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
        os.Exit(1)
    }
    os.Exit(0)
}
