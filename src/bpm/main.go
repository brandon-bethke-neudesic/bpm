package main;


import (
    "os"
    "fmt"
    "path"
    "net/url"
    "strings"
    "github.com/blang/semver"
    "bpmerror"
    "path/filepath"
    "errors"
)

var moduleCache = ModuleCache{Items:make(map[string]*ModuleCacheItem)};
var Options = BpmOptions {
    BpmCachePath: "bpm_modules",
    BpmFileName: "bpm.json",
    LocalModuleName: "local",
    ExcludeFileList: ".git|.gitignore|.gitmodules|bpm_modules|node_modules",
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

func AddRemotes(source string, itemPath string) error {
    git := GitExec{}
    remotes, err := git.GetRemotes();
    if err != nil {
        return bpmerror.New(err, "Error: There was an error attempting to determine the remotes of this repository.")
    }

    git = GitExec{Path: itemPath}
    origin := git.GetRemote("origin");
    if origin == nil {
        return errors.New("Error: The remote 'origin' is missing in this repository.")
    }

    for _, remote := range remotes {
        // Always re-add the remotes
        if remote.Name != "local" && remote.Name != "origin" {
            git.DeleteRemote(remote.Name);
            if strings.HasPrefix(source, "http") {
                fmt.Println("Adding remote " + origin.Url + " as " + remote.Name + " to " + itemPath)
                git.AddRemote(remote.Name, origin.Url)
                if err != nil {
                    return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
                }
            } else {
                parsedUrl, err := url.Parse(remote.Url)
                if err != nil {
                    return bpmerror.New(err, "Error: There was a problem parsing the remote url " + remote.Url)
                }
                adjustedUrl := parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, parsedUrl.Path, source)
                fmt.Println("Adding remote " + adjustedUrl + " as " + remote.Name + " to " + itemPath)
                err = git.AddRemote(remote.Name, adjustedUrl)
                if err != nil {
                    return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
                }
            }
        }
    }

    if Options.UseLocalPath != "" && !strings.HasPrefix(source, "http"){
        source = strings.Split(source, ".git")[0]
        // Rename the origin to local and then add the original origin as origin
        gitSource := GitExec{Path: path.Join(".", source)}
        gitDestination := GitExec{Path: itemPath};
        if !gitDestination.RemoteExists("local") {
            origin := gitSource.GetRemote("origin")
            if origin == nil {
                return errors.New("Error: The remote 'origin' is missing")
            }
            localRemote := path.Join(Options.WorkingDir, source);
            fmt.Println("Adding remote " + localRemote + " as local to " + itemPath)
            err = gitDestination.AddRemote("local", localRemote)
            if err != nil {
                return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
            }
            gitDestination.DeleteRemote("origin");
            fmt.Println("Adding remote " + origin.Url + " as origin to " + itemPath)
            err = gitDestination.AddRemote("origin", origin.Url)
            if err != nil {
                return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
            }
        }
    } else {
        gitDestination := GitExec{Path: itemPath};
        gitDestination.DeleteRemote("local");
    }
    return nil
}

func CopyChanges(source string, destination string) error {

    // The the local option is used and the source repository url is a relative path, then copy any uncommitted changes.
    if Options.UseLocalPath != "" && strings.Index(source, "http") == -1 {
        fmt.Println("Copying local changes from " + source + " to " + destination)

        // Get all the changed files in the source repository, exluding deleted files, and copy them to the destination repository.
        git := GitExec{Path: source}
        files, err := git.LsFiles();
        if err != nil {
            return bpmerror.New(err, "Error: There was an error listing the changes in the repository at " + source)
        }
        copyDir := &CopyDir{};
        for _, file := range files {
            fileSource := path.Join(source, file)
            fileDestination := path.Join(destination, file)

            parent, _ := filepath.Split(fileDestination)
            os.MkdirAll(parent, 0777)
            err = copyDir.CopyFile(fileSource, fileDestination);
            if err != nil {
                return bpmerror.New(err, "Error: There was an error copying the changes from " + fileSource + " to " + fileDestination);
            }
        }

        // Retreive all the deleted files in the source repository and delete them from the destination repository
        files, err = git.DeletedFiles();
        if err != nil {
            return bpmerror.New(err, "Error: There was an error listing the deleted files in the repository at " + source)
        }
        for _, file := range files {
            fileDestination := path.Join(destination, file);
            os.Remove(fileDestination);
        }
    }
    return nil;
}

func ProcessModule(source string) (*BpmData, *ModuleCacheItem, error) {
    source = strings.Split(source, ".git")[0]
    _, itemName := filepath.Split(source)
    itemPath := path.Join(Options.BpmCachePath, itemName);

    if PathExists(itemPath) {
        if Options.Command == "update" {

            git := GitExec{Path: itemPath}
            //if Options.UseLocalPath != "" && strings.Index(source, "http") == -1 {
                // Remove any uncommited changes
                err := git.Checkout(".");
                if err != nil {
                    return nil, nil, bpmerror.New(err, "Error: There was an issue trying to remove all uncommited changes in " + itemPath);
                }
            //}

            err = AddRemotes(source, itemPath);
            if err != nil {
                return nil, nil, err;
            }

            pullRemote := Options.UseRemoteName;
            if git.RemoteExists("local") {
                pullRemote = "local"
            }

            git.LogOutput = true
            err = git.Pull(pullRemote, "master");
            if err != nil {
                return nil, nil, err;
            }
            git.LogOutput = false;

            err = CopyChanges(source, itemPath);
            if err != nil {
                return nil, nil, err;
            }

        } else {
            fmt.Println("Using cached item", itemName)
        }
    } else {
        git := GitExec{LogOutput: true}

        cloneUrl := source;
        if !strings.HasSuffix(cloneUrl, ".git") && (Options.UseLocalPath == "" || strings.HasPrefix(cloneUrl, "http")) {
            cloneUrl = cloneUrl + ".git";
        }

        err := git.Clone(cloneUrl, itemPath);
        if err != nil {
            return nil, nil, err;
        }

        err = AddRemotes(source, itemPath);
        if err != nil {
            return nil, nil, err;
        }

        err = CopyChanges(source, itemPath);
        if err != nil {
            return nil, nil, err;
        }
    }

    moduleBpm, err := LoadBpmData(itemPath)
    if err != nil {
        return nil, nil, err;
    }
    moduleBpmVersion, err := semver.Make(moduleBpm.Version);
    if err != nil {
        fmt.Println("Error: Could not read the version");
        return nil, nil, err;
    }
    cacheItem := &ModuleCacheItem{Name:itemName, Version: moduleBpmVersion.String(), Path: itemPath}
    return moduleBpm, cacheItem, nil;
}

func MakeRemoteUrl(itemUrl string) (string, error) {
    if Options.UseLocalPath != "" && strings.Index(itemUrl, "http") == -1 {
        _, name := filepath.Split(itemUrl)
        return path.Join(Options.UseLocalPath, strings.Split(name, ".git")[0]), nil;
    }

    adjustedUrl := itemUrl;
    if adjustedUrl == "" {
        // If a remote url is specified then use that one, otherwise determine the url of the specified remote name.
        if Options.UseRemoteUrl != "" {
            adjustedUrl = Options.UseRemoteUrl;
        } else {
            git := GitExec{}
            remote := git.GetRemote(Options.UseRemoteName)
            if remote == nil {
                return "", errors.New("Error: There was a problem getting the remote url " + Options.UseRemoteName)
            }
            adjustedUrl = remote.Url;
        }
    } else if strings.Index(adjustedUrl, "http") != 0 {
        var remoteUrl string;
        if Options.UseRemoteUrl != "" {
            remoteUrl = Options.UseRemoteUrl;
        } else {
            git := GitExec{}
            remote := git.GetRemote(Options.UseRemoteName)
            if remote == nil {
                return "", errors.New("Error: There was a problem getting the remote url " + Options.UseRemoteName)
            }
            remoteUrl = remote.Url;
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
    // Always process the keys sorted by name so the installation is consistent
    sortedKeys := bpm.GetSortedKeys();
    for _, itemName := range sortedKeys {

        //if Options.Command == "install" {
            // Do not process the item if it has already been processed
            _, exists := moduleCache.Items[itemName];
            if exists {
                continue;
            }
        //}

        item := bpm.Dependencies[itemName]
        err := item.Validate();
        if err != nil {
            return bpmerror.New(err, "Error: The dependency '" + itemName + "' has an issue")
        }
        if Options.UseLocalPath != "" && strings.Index(item.Url, "http") == -1 {
            moduleSourceUrl := path.Join(Options.UseLocalPath, itemName);
            moduleBpm, cacheItem, err := ProcessModule(moduleSourceUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            fmt.Println("Processing dependencies for", cacheItem.Name, "version", moduleBpm.Version);
            err = ProcessDependencies(moduleBpm, "")
            if err != nil {
                return err;
            }
        } else {
            itemPath := path.Join(Options.BpmCachePath, itemName)
            os.Mkdir(Options.BpmCachePath, 0777)

            itemRemoteUrl := item.Url;
            if PathExists(itemPath) {
                fmt.Println("Using cached item", itemName)
            } else {
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
                os.Mkdir(Options.BpmCachePath, 0777)
                git := GitExec{Path: Options.BpmCachePath, LogOutput: true}
                git.Clone(itemRemoteUrl, itemName)
                if err != nil {
                    os.RemoveAll(path.Join(Options.BpmCachePath, itemName))
                    return bpmerror.New(nil, "Error: There was an issue initializing the repository for dependency " + itemName + " Url: " + itemRemoteUrl + " Commit: " + item.Commit)
                }
                err = AddRemotes(item.Url, itemPath);
                if err != nil {
                    return err;
                }
            }

            // Recursively get dependencies in the current dependency
            moduleBpm := &BpmData{};
            moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName)
            err := moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                return bpmerror.New(err, "Error: Could not load the bpm.json file for dependency " + itemName)
            }

            err = moduleBpm.Validate();
            if err != nil {
                return bpmerror.New(err, "Error: There is an issue with the bpm.json file for " + itemName)
            }
            cacheItem := &ModuleCacheItem{Name:itemName, Version: moduleBpm.Version, Path: itemPath}
            moduleCache.AddLatest(cacheItem)

            fmt.Println("Processing dependencies for", cacheItem.Name, "version", moduleBpm.Version);
            if Options.UseParentUrl {
                err = ProcessDependencies(moduleBpm, itemRemoteUrl)
            } else {
                err = ProcessDependencies(moduleBpm, "")
            }
            if err != nil {
                return err;
            }
        }
    }
    return nil;
}

func main() {
    Options.WorkingDir, _ = os.Getwd();
    cmd := NewBpmCommand();
    if err := cmd.Execute(); err != nil {
        fmt.Println(err);
        fmt.Println("Finished with errors");
        os.Exit(1)
    }
    os.Exit(0);
}
