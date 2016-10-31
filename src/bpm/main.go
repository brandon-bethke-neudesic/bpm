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
    itemPathTemp := path.Join(Options.WorkingDir, Options.BpmCachePath, "xx_temp_xx", "xx_temp_xx")
    defer os.RemoveAll(path.Join(itemPathTemp, ".."))
    os.RemoveAll(path.Join(itemPathTemp));
    os.MkdirAll(itemPathTemp, 0777)
    moduleBpm := &BpmData{};
    git := GitExec{Path:itemPathTemp}
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

func MakeRemoteUrl(itemUrl string) (string, error) {
    adjustedUrl := itemUrl;
    var err error;
    if adjustedUrl == "" {
        // If a remote url is specified then use that one, otherwise determine the url of the specified remote name.
        if Options.UseRemoteUrl != "" {
            adjustedUrl = Options.UseRemoteUrl;
        } else {
            git := GitExec{Path:Options.WorkingDir}
            adjustedUrl, err = git.GetRemoteUrl(Options.UseRemoteName)
            if err != nil {
                return "", bpmerror.New(err, "Error: There was a problem getting the remote url " + Options.UseRemoteName)
            }
        }
    } else if strings.Index(adjustedUrl, "http") != 0 {
        var remoteUrl string;
        if Options.UseRemoteUrl != "" {
            remoteUrl = Options.UseRemoteUrl;
        } else {
            git := GitExec{Path:Options.WorkingDir}
            remoteUrl, err = git.GetRemoteUrl(Options.UseRemoteName)
            if err != nil {
                return "", bpmerror.New(err, "Error: There was a problem getting the remote url " + Options.UseRemoteName)
            }
        }
        parsedUrl, err := url.Parse(remoteUrl)
        if err != nil {
            return "", bpmerror.New(err, "Error: There was a problem parsing the remote url " + remoteUrl)
        }
        adjustedUrl = parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, parsedUrl.Path, itemUrl)
    }
    return adjustedUrl, nil;
}

type ItemProcessedEvent func(item *ItemProcessed) error;

func ProcessDependencies(bpm *BpmData, parentUrl string, itemProcessedEvent ItemProcessedEvent) (error) {
    // Always process the keys sorted by name so the installation is consistent
    sortedKeys := bpm.GetSortedKeys();
    for _, itemName := range sortedKeys {
        item := bpm.Dependencies[itemName]
        fmt.Println("Validating dependency", itemName)
        err := item.Validate();
        if err != nil {
            return err;
        }
        if Options.UseLocalPath != "" && strings.Index(item.Url, "http") == -1 {
            moduleSourceUrl := path.Join(Options.UseLocalPath, itemName);
            fmt.Println("Processing local dependency in", moduleSourceUrl)
            moduleBpm, cacheItem, err := ProcessLocalModule(moduleSourceUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            err = ProcessDependencies(moduleBpm, "", itemProcessedEvent)
            if err != nil {
                return err;
            }

            if itemProcessedEvent != nil {
                err = itemProcessedEvent(&ItemProcessed{Bpm:bpm, Source: moduleSourceUrl, Name: itemName, Item: item, Local: true})
                if err != nil {
                    return err;
                }
            }

        } else {
            fmt.Println("Processing dependency", itemName)

            itemPath := path.Join(Options.BpmCachePath, itemName)
            os.Mkdir(itemPath, 0777)

            itemRemoteUrl := item.Url;
            itemClonePath := path.Join(Options.WorkingDir, itemPath, item.Commit)
            localPath := path.Join(Options.BpmCachePath, itemName, Options.LocalModuleName)
            if PathExists(localPath) {
                fmt.Println("Found local folder in the bpm modules. Using this folder", localPath)
                itemClonePath = localPath;
            } else if !PathExists(itemClonePath) {
                fmt.Println("Could not find module", itemName, "in the bpm cache. Cloning repository...")
                if item.Commit == "local" {
                    return bpmerror.New(nil, "Error: The commit hash is specified as 'local' for dependency " + itemName + ". Please finalize the commit hash for this dependency.")
                }
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
                git := GitExec{Path: itemClonePath}
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
            err = ProcessDependencies(moduleBpm, itemRemoteUrl, nil)
            if err != nil {
                return err;
            }

            if itemProcessedEvent != nil {
                err = itemProcessedEvent(&ItemProcessed{Bpm:bpm, Source: itemClonePath, Name: itemName, Item: item, Local: false})
                if err != nil {
                    return err;
                }
            }
        }
    }
    return nil;
}

func main() {
    Options.Parse(os.Args);
    err := Options.Validate();
    if err == nil {
        err = Options.Command.Execute();
    }
    if err != nil {
        fmt.Println(err)
        fmt.Println("Finished with errors")
        os.Exit(1)
    }
    fmt.Println("Finished")
    os.Exit(0)
}
