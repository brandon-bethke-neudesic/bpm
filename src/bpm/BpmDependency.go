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
    "github.com/blang/semver"
    "sort"
)

type BpmDependency struct {
    Name string
    Url string
    Path string
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

    updatePackageJson := false;

    if len(files) > 0 {
        updatePackageJson = true;
    }

    for _, file := range files {
        fileSource := path.Join(source, file)
        fileDestination := path.Join(destination, file)

        fileInfo, err := os.Stat(fileSource)
        if err != nil {
	        return bpmerror.New(err, "Error: There was an error trying to get information about the changed file");
        }
        if fileInfo.IsDir() {
            os.MkdirAll(fileDestination, fileInfo.Mode())
            err = copyDir.Copy(fileSource, fileDestination);
            if err != nil {
                return bpmerror.New(err, "Error: There was an error copying the changes from " + fileSource + " to " + fileDestination);
            }
        } else {
            parent, _ := filepath.Split(fileDestination)
            os.MkdirAll(parent, fileInfo.Mode())
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

    if len(files) > 0 || updatePackageJson {
    	err = UpdatePackageJsonVersion(source);
        if err != nil {
            return err;
        }
        err = copyDir.CopyFile(path.Join(source, "package.json"), path.Join(destination, "package.json"));
        if err != nil {
            return err;
        }
    }
    return nil;
}

func (dep *BpmDependency) SwitchBranches(source string, destination string) (string, error) {
    // Get all the changed files in the source repository, exluding deleted files, and copy them to the destination repository.
    gitSource := GitExec{Path: source}
    gitDestination := GitExec{Path: destination}
    branch, err := gitSource.GetCurrentBranch();
    if err != nil {
        return "", bpmerror.New(err, "Error: Could not find the current branch in the source repository");
    }
    if branch != "HEAD" {
        fmt.Println("Source repository is on branch " + branch + ". Switching to branch " + branch);
        err := gitDestination.Checkout(branch)
        if err != nil {
            return "", bpmerror.New(err, "Error: Could not switch to branch " + branch + " in the destination repository")
        }
    }
    return branch, nil;
}

func (dep *BpmDependency) Scan() (error) {
    git := &GitExec{Path: dep.Path}
    itemCommit, err := git.GetLatestCommit();
    if err != nil {
        return bpmerror.New(err, "Error: Could not get the latest commit for " + git.Path);
    }
    cacheItem := &ModuleCacheItem{Name:dep.Name, Path: dep.Path, Commit: itemCommit}
    existingItem, exists := moduleCache.Items[dep.Name];
    if !exists {
        moduleCache.Add(cacheItem);
    } else {
        if existingItem.Commit != itemCommit {
            fmt.Println(dep.Name + " is already cached. Attempting to determine latest commit. " + existingItem.Commit + " or " + itemCommit)
            git := GitExec{Path: dep.Path}
            result := git.DetermineLatest(itemCommit, existingItem.Commit)
            if result != existingItem.Commit {
                moduleCache.Add(cacheItem)
            }
            fmt.Println("The most recent commit for " + dep.Name + " is " + result);
        } else {
            moduleCache.Add(cacheItem);
        }
    }

    bpm := BpmModules{}
    err = bpm.Load(path.Join(dep.Path, Options.BpmCachePath));
    if err != nil {
        return err;
    }

	if len(bpm.Dependencies) > 0 {
        fmt.Println("Scanning " + dep.Path);
   	}
    for _, subdep := range bpm.Dependencies {
    	fmt.Println("Found " + subdep.Path); 
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
        source = path.Join(Options.Local, dep.Name);
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
    fmt.Println("Updating " + dep.Name)

    if !PathExists(dep.Path) {
        err = dep.Add();
        if err != nil {
            return err;
        }
    }

    // If the item is cached then it is already updated and there is no reason to update it again.
    // Unless this is a deep update
    _, exists := moduleCache.Items[itemName];
    if exists && !Options.Deep {
        return nil;
    }
    git := &GitExec{Path: dep.Path}
	// Clean all uncommitted changes
    err = git.Checkout(".");
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue trying to remove all uncommited changes in " + dep.Path);
    }

    err = dep.AddRemotes(source, dep.Path);
    if err != nil {
        return err;
    }

    pullRemote := Options.Remote;
    // The local remote always has precedence.
    if git.RemoteExists("local") {
        pullRemote = "local"
    }
    git.LogOutput = true
    err = git.Fetch(pullRemote)
    if err != nil {
        return err;
    }
	branch := Options.Branch;
    if pullRemote == "local" {
	    git.LogOutput = false;
	    // The local branch always has precedence
        branch, err = dep.SwitchBranches(source, dep.Path);
        if err != nil {
            return err;
        }
        if branch == "HEAD" {
            branch = "master"
        }
        git.LogOutput = true;
    } else {        
    	// Switch back to the master branch if there is no local remote
	    git.Checkout(branch);
    }
    
    git.LogOutput = true;
	err = git.Merge(pullRemote, branch);
    //err = git.Pull(pullRemote, branch);
    if err != nil {
        return err;
    }

    _, err = git.Run("git submodule update --init --recursive")
    if err != nil {
        return err
    }
    git.LogOutput = false;

    if pullRemote == "local" {
        err = dep.CopyChanges(source, dep.Path);
        if err != nil {
            return err;
        }
    }

    itemCommit, err := git.GetLatestCommit();
    if err != nil {
        return bpmerror.New(err, "Error: Could not get the latest commit for " + git.Path);
    }
    cacheItem := &ModuleCacheItem{Name:dep.Name, Path: dep.Path, Commit: itemCommit}
    moduleCache.Add(cacheItem)

    bpm := BpmModules{}
    err = bpm.Load(path.Join(dep.Path, Options.BpmCachePath));
    if err != nil {
        return err;
    }

	if len(bpm.Dependencies) > 0 {
	    fmt.Println("Scanning " + dep.Path)		
	}
    for _, subdep := range bpm.Dependencies {
    	fmt.Println("Found " + subdep.Path);
        if Options.Deep {
            err = subdep.Update();
        } else {
            err = subdep.Scan();
        }
        if err != nil {
            return err;
        }
    }
    UpdatePackageJsonVersion(".")
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
    remotesModule, err := git.GetRemotes();
    if err != nil {
        return bpmerror.New(err, "Error: There was an error attempting to determine the remotes of " + itemPath)
    }
        
    // Delete the remotes in the dependency that no longer are in the root repository
	for _, remote := range remotesModule {
		if remote.Name != "local" && remote.Name != "origin" {
			i := sort.Search(len(remotes), func(i int) bool { return remotes[i].Name == remote.Name })						
			if i >= len(remotes) {
				git.DeleteRemote(remote.Name);
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
        gitDestination.DeleteRemote("origin");
        
        originUrl := source;
	    if !strings.HasSuffix(originUrl, ".git") {
	        originUrl = originUrl + ".git";
	    }        
        fmt.Println("Adding remote " + originUrl + " as origin to " + itemPath)
        err = gitDestination.AddRemote("origin", originUrl)
        if err != nil {
            return bpmerror.New(err, "Error: There was an error adding the remote to the repository at " + itemPath)
        }                
    }
    return nil
}

func (dep *BpmDependency) Install() (error) {
    pj := &PackageJson{Path: dep.Path};
    pj.Load();

    newItem := &ModuleCacheItem{Name:dep.Name, Path: dep.Path, Version: pj.Version.String()}
    existingItem, exists := moduleCache.Items[dep.Name];
    if exists {
    	fmt.Println(dep.Name + " is already cached. Comparing package.json versions...");
        existingVersion, _ := semver.Make(existingItem.Version)
        newVersion, _ := semver.Make(newItem.Version)
        if newVersion.GT(existingVersion) {
            moduleCache.Add(newItem)
        }
    } else {
        moduleCache.Add(newItem)
    }
    fmt.Println(dep.Name + "@" + newItem.Version + " will be installed");
    bpm := BpmModules{}
    err := bpm.Load(path.Join(dep.Path, Options.BpmCachePath));
    if err != nil {
        return err;
    }
	if len(bpm.Dependencies) > 0 {
	    fmt.Println("Scanning " + dep.Path);
	}
    for _, subdep := range bpm.Dependencies {
    	fmt.Println("Found " + subdep.Path);
        err = subdep.Install();
        if err != nil {
            return err;
        }
    }
    return nil;
}

func (dep *BpmDependency) Add() (error) {
    var err error;
    source := "";
    if UseLocal(dep.Url) {
        source = path.Join(Options.Local, dep.Name);
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
    // If the item is cached then it is already installed and there is no reason to install it again.
    _, exists := moduleCache.Items[dep.Name];
    if exists {
        return nil;
    }

    cacheItem := &ModuleCacheItem{Name:dep.Name, Path: dep.Path}
    moduleCache.Add(cacheItem)

    fmt.Println("Adding submodule " + dep.Name)
    cloneUrl := dep.Url;
    if cloneUrl == "" {
        cloneUrl = source;
    }
    if !strings.HasSuffix(cloneUrl, ".git") {
        cloneUrl = cloneUrl + ".git";
    }

    git := GitExec{LogOutput:true};
    err = git.AddSubmodule("--force " + cloneUrl, dep.Path)
    if err != nil {
        return bpmerror.New(err, "Error: Could not add the submodule " + dep.Path);
    }
    git = GitExec{Path: dep.Path, LogOutput: true}

    err = git.Fetch("origin");
    if err != nil {
        return err;
    }

    err = git.Checkout(Options.Branch)
    if err != nil {
        return err;
    }
    git.LogOutput = false;
    err = dep.AddRemotes(source, dep.Path);
    if err != nil {
        return err;
    }

    if UseLocal(source) {
        err := git.Fetch("local")
        if err != nil {
            return err;
        }
	    // The local branch always has precedence
        branch, err := dep.SwitchBranches(source, dep.Path);
        if err != nil {
            return err
        }
        if branch == "HEAD" {
	        branch = "master";
        }
        err = git.Merge("local", branch)
        if err != nil {
            return err;
        }
    }
    git.LogOutput = true;
    _, err = git.Run("git submodule update --init --recursive")
    if err != nil {
        return err
    }
    git.LogOutput = false;

    if UseLocal(source){
        err = dep.CopyChanges(source, dep.Path);
        if err != nil {
            return err;
        }
    }

    git = GitExec{Path: dep.Path, LogOutput: true}
    itemCommit, err := git.GetLatestCommit();
    if err != nil {
        return bpmerror.New(err, "Error: Could not get the latest commit for " + git.Path);
    }
    cacheItem.Commit = itemCommit;

    bpm := BpmModules{}
    err = bpm.Load(path.Join(dep.Path, Options.BpmCachePath));
    if err != nil {
        return err;
    }
    UpdatePackageJsonVersion(".")

	if len(bpm.Dependencies) > 0 {
	    fmt.Println("Scanning " + dep.Path)
	}
    for _, subdep := range bpm.Dependencies {
        fmt.Println("Found " + subdep.Path)
        err = subdep.Scan();
        if err != nil {
            return err;
        }
    }
    return nil;
}
