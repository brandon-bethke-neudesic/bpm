package main;

import (
    "path/filepath"
    "strings"
    "bpmerror"
    "fmt"
    "path"
    "os"
    "net/url"
    "errors"
)

type BpmDependency struct {
    Name string
    Url string
    Path string
}

func (dep *BpmDependency) Equal(item *BpmDependency) bool {
    // If the address is the same, then it's equal.
    if dep == item {
        return true;
    }
    if item.Url == dep.Url {
        return true;
    }
    return false;
}

func (dep *BpmDependency) CopyChanges(source string, destination string) error {
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
    return nil;
}


func (dep *BpmDependency) SwitchBranches(source string, destination string) error {
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
    return nil;
}

func (dep *BpmDependency) InstallNew(componentUrl string) (error) {
    parsedUrl, err := url.Parse(componentUrl);
    if err != nil {
        return bpmerror.New(err, "Error: Could not parse the component url");
    }

    if dep.Name == "" {
        _, name := filepath.Split(parsedUrl.Path)
        name = strings.Split(name, ".git")[0]
        dep.Name = name;
    }

    if dep.Url == "" {
        dep.Url = componentUrl;
    }

    return dep.Install();
}

func (dep *BpmDependency) Scan() (error) {
    git := &GitExec{Path: dep.Path}
    itemCommit, err := git.GetLatestCommit();
    if err != nil {
        return err;
    }
    cacheItem := &ModuleCacheItem{Name:dep.Name, Path: dep.Path, Commit: itemCommit}
    existingItem, exists := moduleCache.Items[dep.Name];
    if !exists {
        fmt.Println("Adding item to cache", dep.Name)
        moduleCache.Add(cacheItem);
    } else {
        if existingItem.Commit != itemCommit {
            fmt.Println("Attempting to determine latest commit. " + existingItem.Commit + " or " + itemCommit)
            git := GitExec{Path: dep.Path}
            result := git.DetermineLatest(itemCommit, existingItem.Commit)

            //fmt.Println("current", itemCommit)
            //fmt.Println("existing", existingItem.Commit)
            //fmt.Println("result", result)

            if result != existingItem.Commit {
                moduleCache.Add(cacheItem)
            }
            fmt.Println("The most recent commit for " + dep.Name + " is " + result);
        } else {
            moduleCache.Add(cacheItem);
        }
    }

    bpm := BpmModules{}
    fmt.Println("Loading ", path.Join(dep.Path, Options.BpmCachePath))
    err = bpm.Load(path.Join(dep.Path, Options.BpmCachePath));
    if err != nil {
        return err;
    }

    fmt.Println("Scanning dependencies for " + dep.Name)
    fmt.Println(len(bpm.Dependencies))
    fmt.Println(bpm.Dependencies)
    for _, subdep := range bpm.Dependencies {
        err = subdep.Scan();
        if err != nil {
            return err;
        }
    }

    return nil;
}

func (dep *BpmDependency) Update() (error) {
    var err error;
    source := "";
    if UseLocal(dep.Url) {
        source = path.Join(Options.UseLocalPath, dep.Name);
    } else {
        source = dep.Url;
        if source == "" {
            source = dep.Name;
        }
        source, err = MakeRemoteUrl(source);
        if err != nil {
            return err;
        }
    }
    source = strings.Split(source, ".git")[0]
    _, itemName := filepath.Split(source)
    itemPath := path.Join(Options.BpmCachePath, itemName);

    fmt.Println("Processing item " + dep.Name)

    if !PathExists(itemPath) {
        err = dep.Install();
        if err != nil {
            return err;
        }
    }
    _, exists := moduleCache.Items[itemName];
    if exists {
        fmt.Println(dep.Name + " has already been processed.")
        return nil;
    }

    cacheItem := &ModuleCacheItem{Name:dep.Name, Path: itemPath}
    moduleCache.Add(cacheItem)

    git := GitExec{Path: itemPath}
    err = git.Checkout(".");
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue trying to remove all uncommited changes in " + itemPath);
    }

    err = dep.AddRemotes(source, itemPath);
    if err != nil {
        return err;
    }

    branch := "master";
    pullRemote := Options.UseRemoteName;
    if git.RemoteExists("local") {
        pullRemote = "local"
        branch, err := git.GetCurrentBranch();
        if err != nil {
            return bpmerror.New(err, "Error: Could not find the current branch in repository " + itemPath);
        }
        if branch == "HEAD" {
            branch = "master"
        }
    }

    git.LogOutput = true

    err = git.Pull(pullRemote, branch);
    if err != nil {
        return err;
    }
    git.LogOutput = false;

    if UseLocal(source){
        err = dep.SwitchBranches(source, itemPath);
        if err != nil {
            return err;
        }
    }

    _, err = git.Run("git submodule update --init --recursive")
    if err != nil {
        return err
    }

    if UseLocal(source) {
        err = dep.CopyChanges(source, itemPath);
        if err != nil {
            return err;
        }
    }

    bpm := BpmModules{}
    err = bpm.Load(path.Join(itemPath, Options.BpmCachePath));
    if err != nil {
        return err;
    }

    fmt.Println("Scanning dependencies for " + dep.Name)
    for _, subdep := range bpm.Dependencies {
        err = subdep.Scan();
        if err != nil {
            return err;
        }
    }
    return nil;
}

func (dep *BpmDependency) AddRemotes(source string, itemPath string) error {
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


func (dep *BpmDependency) Install() (error) {
    var err error;
    source := "";
    if UseLocal(dep.Url) {
        source = path.Join(Options.UseLocalPath, dep.Name);
    } else {
        source = dep.Url;
        if source == "" {
            source = dep.Name;
        }
        source, err = MakeRemoteUrl(source);
        if err != nil {
            return err;
        }
    }

    source = strings.Split(source, ".git")[0]
    itemPath := path.Join(Options.BpmCachePath, dep.Name);
    dep.Path = itemPath;

    fmt.Println("Installing item " + dep.Name)

    _, exists := moduleCache.Items[dep.Name];
    if exists {
        fmt.Println(dep.Name + " has already been installed.")
        return nil;
    }

    cacheItem := &ModuleCacheItem{Name:dep.Name, Path: itemPath}
    moduleCache.Add(cacheItem)

    if !PathExists(itemPath) {
        cloneUrl := source;
        if !strings.HasSuffix(cloneUrl, ".git") && (Options.UseLocalPath == "" || strings.HasPrefix(cloneUrl, "http")) {
            cloneUrl = cloneUrl + ".git";
        }

        git := GitExec{LogOutput:true};
        err = git.AddSubmodule("--force " + cloneUrl, itemPath)
        if err != nil {
            return bpmerror.New(err, "Error: Could not add the submodule " + itemPath);
        }
        git = GitExec{Path: itemPath, LogOutput: true}

        if UseLocal(cloneUrl) {
            cloneUrl = path.Join(Options.WorkingDir, source)
        }

        err = git.Fetch("origin");
        if err != nil {
            return err;
        }

        err = git.Checkout("master")
        if err != nil {
            return err;
        }
        git.LogOutput = false;
        err = dep.AddRemotes(source, itemPath);
        if err != nil {
            return err;
        }

        if UseLocal(source) {
            err := git.Fetch("local")
            if err != nil {
                return err;
            }
            err = git.Pull("local", "master")
            if err != nil {
                return err;
            }
            git.LogOutput = true;
            err = dep.SwitchBranches(source, itemPath);
            if err != nil {
                return err
            }
            git.LogOutput = false;
        }
        git.LogOutput = true;
        _, err = git.Run("git submodule update --init --recursive")
        if err != nil {
            return err
        }
        git.LogOutput = false;

        if UseLocal(source){
            err = dep.CopyChanges(source, itemPath);
            if err != nil {
                return err;
            }
        }
    } else {
        fmt.Println(dep.Name + " has already been installed.")
    }

    bpm := BpmModules{}
    fmt.Println("Loading bpm modules for " + path.Join(itemPath, Options.BpmCachePath))
    err = bpm.Load(path.Join(itemPath, Options.BpmCachePath));
    if err != nil {
        return err;
    }

    for _, subdep := range bpm.Dependencies {
        fmt.Println("Scanning dependencies for " + subdep.Path)

        fmt.Println(subdep.Name, subdep.Path);

        err = subdep.Scan();
        if err != nil {
            return err;
        }
    }
    return nil;
}
