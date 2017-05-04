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
                remoteUrl := origin.Url;
                git.AddRemote(remote.Name, remoteUrl)
                if err != nil {
                    return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
                }
            } else {
                parsedUrl, err := url.Parse(remote.Url)
                if err != nil {
                    return bpmerror.New(err, "Error: There was a problem parsing the remote url " + remote.Url)
                }
                adjustedUrl := parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, parsedUrl.Path, source)
                if !strings.HasSuffix(adjustedUrl, ".git") {
                    adjustedUrl = adjustedUrl + ".git"
                }
                fmt.Println("Adding remote " + adjustedUrl + " as " + remote.Name + " to " + itemPath)
                err = git.AddRemote(remote.Name, adjustedUrl)
                if err != nil {
                    return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
                }
            }
        }
    }

    if UseLocal(source) {
        source = strings.Split(source, ".git")[0]
        // Rename the origin to local and then add the original origin as origin
        gitSource := GitExec{Path: path.Join(".", source)}
        gitDestination := GitExec{Path: itemPath};

        // Always make sure the local remote is correct.
        gitDestination.DeleteRemote("local");
        localRemote := path.Join(Options.WorkingDir, source);
        fmt.Println("Adding remote " + localRemote + " as local to " + itemPath)
        err = gitDestination.AddRemote("local", localRemote)
        if err != nil {
            return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
        }

        if !gitDestination.RemoteExists("local") {
            origin := gitSource.GetRemote("origin")
            if origin == nil {
                return errors.New("Error: The remote 'origin' is missing")
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

func UseLocal(url string) bool {
    if Options.UseLocalPath != "" && !strings.HasPrefix(url, "http") {
        return true;
    }
    return false;
}

func SwitchBranches(source string, destination string) error {
    // The the local option is used and the source repository url is a relative path, then copy any uncommitted changes.
    if UseLocal(source) {
        // Get all the changed files in the source repository, exluding deleted files, and copy them to the destination repository.
        gitSource := GitExec{Path: source}
        gitDestination := GitExec{Path: destination}
        branch, err := gitSource.GetCurrentBranch();
        if err != nil {
            return bpmerror.New(err, "Error: Could not find the current branch in the source repository");
        }
        if branch != "HEAD" {
            fmt.Println("Source repository is on branch " + branch + ". Switching to branch " + branch);
            err := gitDestination.Checkout(branch)
            if err != nil {
                return bpmerror.New(err, "Error: Could not switch to branch " + branch + " in the destination repository")
            }
        }
    }
    return nil;
}

func CopyChanges(source string, destination string) error {

    // The the local option is used and the source repository url is a relative path, then copy any uncommitted changes.
    if UseLocal(source) {

        // Get all the changed files in the source repository, exluding deleted files, and copy them to the destination repository.
        gitSource := GitExec{Path: source}
        files, err := gitSource.LsFiles();
        if err != nil {
            return bpmerror.New(err, "Error: There was an error listing the changes in the repository at " + source)
        }
        fmt.Println("Copying local changes from " + source + " to " + destination)
        copyDir := &CopyDir{};
        for _, file := range files {
            fileSource := path.Join(source, file)
            fileDestination := path.Join(destination, file)

            fileInfo, err := os.Stat(fileSource)
            if fileInfo.IsDir() {
                os.MkdirAll(fileDestination, 0777)
                err = copyDir.Copy(fileSource, fileDestination);
                if err != nil {
                    return bpmerror.New(err, "Error: There was an error copying the changes from " + fileSource + " to " + fileDestination);
                }
            } else {
                parent, _ := filepath.Split(fileDestination)
                os.MkdirAll(parent, 0777)
                err = copyDir.CopyFile(fileSource, fileDestination);
                if err != nil {
                    return bpmerror.New(err, "Error: There was an error copying the changes from " + fileSource + " to " + fileDestination);
                }
            }
        }
        // Retreive all the deleted files in the source repository and delete them from the destination repository
        files, err = gitSource.DeletedFiles();
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

func EnsurePath(path string){
    if _, err := os.Stat(path); os.IsNotExist(err) {
        os.MkdirAll(path, 0777)
    }
}

func ProcessModule(source string) (*BpmData, *ModuleCacheItem, error) {
    source = strings.Split(source, ".git")[0]
    _, itemName := filepath.Split(source)
    itemPath := path.Join(Options.BpmCachePath, itemName);
    isLocal := false;
    if PathExists(itemPath) {
        if Options.Command == "update" {

            git := GitExec{Path: itemPath}
            err := git.Checkout(".");
            if err != nil {
                return nil, nil, bpmerror.New(err, "Error: There was an issue trying to remove all uncommited changes in " + itemPath);
            }
            err = AddRemotes(source, itemPath);
            if err != nil {
                return nil, nil, err;
            }

            branch := "master";
            pullRemote := Options.UseRemoteName;
            if git.RemoteExists("local") {
                isLocal = true;
                pullRemote = "local"
                branch, err := git.GetCurrentBranch();
                if err != nil {
                    return nil, nil, bpmerror.New(err, "Error: Could not find the current branch in repository " + itemPath);
                }
                if branch == "HEAD" {
                    branch = "master"
                }
            }

            git.LogOutput = true

            err = git.Pull(pullRemote, branch);
            if err != nil {
                return nil, nil, err;
            }
            git.LogOutput = false;

            err = SwitchBranches(source, itemPath);
            if err != nil {
                return nil, nil, err;
            }
            err = CopyChanges(source, itemPath);
            if err != nil {
                return nil, nil, err;
            }
        } else {
            fmt.Println("Using cached item", itemName)
        }
    } else {

        cloneUrl := source;
        if !strings.HasSuffix(cloneUrl, ".git") && (Options.UseLocalPath == "" || strings.HasPrefix(cloneUrl, "http")) {
            cloneUrl = cloneUrl + ".git";
        }

        git := GitExec{LogOutput:true};
        err := git.AddSubmodule("--force " + cloneUrl, itemPath)
        if err != nil {
            return nil, nil, bpmerror.New(err, "Error: Could not add the submodule " + itemPath);
        }
        git = GitExec{Path: itemPath, LogOutput: true}

        if UseLocal(cloneUrl) {
            cloneUrl = path.Join(Options.WorkingDir, source)
        }

        err = git.Fetch("origin");
        if err != nil {
            return nil, nil, err;
        }

        err = git.Checkout("master")
        if err != nil {
            return nil, nil, err;
        }
        git.LogOutput = false;
        err = AddRemotes(source, itemPath);
        if err != nil {
            return nil, nil, err;
        }

        if UseLocal(source) {
            isLocal = true;
            err := git.Fetch("local")
            if err != nil {
                return nil, nil, err;
            }
            err = git.Pull("local", "master")
            if err != nil {
                return nil, nil, err;
            }
        }

        git.LogOutput = true;
        err = SwitchBranches(source, itemPath);
        if err != nil {
            return nil, nil, err
        }
        _, err = git.Run("git submodule update --init --recursive")
        if err != nil {
            return nil, nil, err
        }
        git.LogOutput = false;

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
    cacheItem := &ModuleCacheItem{Name:itemName, Version: moduleBpmVersion.String(), Path: itemPath, IsLocal: isLocal}
    return moduleBpm, cacheItem, nil;
}

func MakeRemoteUrl(itemUrl string) (string, error) {
    if UseLocal(itemUrl) {
        _, name := filepath.Split(itemUrl)
        return path.Join(Options.UseLocalPath, strings.Split(name, ".git")[0]), nil;
    }
    adjustedUrl := itemUrl;
    if !strings.HasPrefix(adjustedUrl, "http") {
        var remoteUrl string;

        relative := "";
        if Options.UseRemoteUrl != "" {
            remoteUrl = Options.UseRemoteUrl;
        } else {
            git := GitExec{}
            remote := git.GetRemote(Options.UseRemoteName)
            if remote == nil {
                return "", errors.New("Error: There was a problem getting the remote url " + Options.UseRemoteName)
            }
            remoteUrl = remote.Url;
            relative = "..";
        }
        parsedUrl, err := url.Parse(remoteUrl)
        if err != nil {
            return "", bpmerror.New(err, "Error: There was a problem parsing the remote url " + remoteUrl)
        }

        urlPath := strings.Split(parsedUrl.Path, ".git")[0]
        adjustedUrl = parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, urlPath, relative, itemUrl)
        if !strings.HasSuffix(adjustedUrl, ".git") {
            adjustedUrl = adjustedUrl + ".git";
        }
        fmt.Println("Adjusted Url", adjustedUrl)
    }
    return adjustedUrl, nil;
}

func ProcessDependencies(bpm *BpmData) (error) {
    // Always process the keys sorted by name so the installation is consistent
    sortedKeys := bpm.GetSortedKeys();
    for _, itemName := range sortedKeys {

        _, exists := moduleCache.Items[itemName];
        if exists {
            continue;
        }
        item := bpm.Dependencies[itemName]
        err := item.Validate();
        if err != nil {
            return bpmerror.New(err, "Error: The dependency '" + itemName + "' has an issue")
        }
        if UseLocal(item.Url) {
            moduleSourceUrl := path.Join(Options.UseLocalPath, itemName);
            moduleBpm, cacheItem, err := ProcessModule(moduleSourceUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            fmt.Println("Processing dependencies for", cacheItem.Name, "version", moduleBpm.Version);
            err = ProcessDependencies(moduleBpm)
            if err != nil {
                return err;
            }
        } else {
            var moduleBpm *BpmData = nil;
            var cacheItem *ModuleCacheItem = nil;

            itemPath := path.Join(Options.BpmCachePath, itemName)

            if PathExists(itemPath) {
                fmt.Println("Using cached item", itemName)
                moduleBpm, err = LoadBpmData(itemPath)
                if err != nil {
                    return err;
                }
                isLocal := false;
                git := GitExec{Path: itemPath};
                if git.RemoteExists("local") {
                    isLocal = true;
                }
                cacheItem = &ModuleCacheItem{Name:itemName, Version: moduleBpm.Version, Path: itemPath, IsLocal: isLocal}
            } else {
                var err error;
                itemRemoteUrl := item.Url;
                if itemRemoteUrl == "" {
                    itemRemoteUrl = itemName;
                }
                itemRemoteUrl, err = MakeRemoteUrl(itemRemoteUrl);
                moduleBpm, cacheItem, err = ProcessModule(itemRemoteUrl)
                if err != nil {
                    return err;
                }
            }
            moduleCache.AddLatest(cacheItem)
            fmt.Println("Processing dependencies for", cacheItem.Name, "version", moduleBpm.Version);
            err = ProcessDependencies(moduleBpm)
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
