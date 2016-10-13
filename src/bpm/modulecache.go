package main;

import (
    "github.com/blang/semver"
    "fmt"
    "strings"
    "os"
    "path"
    "io/ioutil"
)

type ModuleCacheItem struct {
    Name string
    Version string
    Commit string
    Path string
}

type ModuleCache struct {
    Items map[string]ModuleCacheItem
}

func (r *ModuleCache) Delete(item string) {
    delete(r.Items, item)
    return
}

func (r *ModuleCache) Install() (error) {

    if Options.PackageManager == "npm" {
        return r.NpmInstall()
        return nil;
    } else if Options.PackageManager == "yarn" {
        return r.CopyAndYarnInstall("./node_modules");
    } else {
        fmt.Println("Error: Unrecognized package manager", Options.PackageManager)
        return nil;
    }
}

func (r *ModuleCache) NpmInstall() (error){
    fmt.Println("Npm installing dependencies to node_modules...")
    workingPath,_ := os.Getwd();
    // Go through each item in the bpm memory cache. There is suppose to only be one item per dependency
    for depName := range r.Items {
        fmt.Println("Processing cached dependency", depName)
        depItem := r.Items[depName];
        // Perform the npm install and pass the url of the dependency. npm install ./bpm_modules/mydep
        npm := NpmCommands{Path: workingPath}
        err := npm.InstallUrl(depItem.Path)
        if err != nil {
            fmt.Println("Error: Failed to npm install module", depName)
            return err;
        }

    }
    return nil;
}

func (r *ModuleCache) Trim() (error) {
    for depName := range r.Items {
        depItem := r.Items[depName];
        // Delete previous cached items
        entries, _ := ioutil.ReadDir(path.Join(Options.BpmCachePath, depItem.Name))
        for _, entry := range entries {
            if entry.IsDir() && entry.Name() != Options.LocalModuleName && entry.Name() != depItem.Commit {
                fmt.Println("Removing previous cache item ...", path.Join(depItem.Name, entry.Name()))
                os.RemoveAll(path.Join(Options.BpmCachePath, depItem.Name, entry.Name()))
            }
        }
    }
    return nil
}

func (r *ModuleCache) CopyAndYarnInstall(nodeModulesPath string) (error) {
    fmt.Println("Copying dependencies to the node_modules folder")
    if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
        fmt.Println("node_modules folder not found. Creating it...")
        os.Mkdir(nodeModulesPath, 0777)
    }
    // Go through each item in the bpm memory cache. There is suppose to only be one item per dependency
    for depName := range r.Items {
        fmt.Println("Processing cached dependency", depName)
        depItem := r.Items[depName];
        nodeModulesItemPath := path.Join(nodeModulesPath, depName);
        // Delete existing copy in node_modules
        os.RemoveAll(nodeModulesItemPath);

        // Perform the yarn install
        yarn := YarnCommands{Path: depItem.Path}
        yarn.ParseOptions(os.Args);
        err := yarn.Install()
        if err != nil {
            fmt.Println("Error: Failed to yarn install module", depName)
            return err;
        }

        // Copy the library to the node_modules folder
        copyDir := CopyDir{Exclude:Options.ExcludeFileList}
        err = copyDir.Copy(depItem.Path, nodeModulesItemPath);
        if err != nil {
            fmt.Println("Error: Failed to copy module to node_modules folder")
            fmt.Println(err)
            return err;
        }
    }
    return nil;
}

func (r *ModuleCache) CopyAndNpmInstall(nodeModulesPath string) (error){
    fmt.Println("Copying dependencies to the node_modules folder")
    if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
        fmt.Println("node_modules folder not found. Creating it...")
        os.Mkdir(nodeModulesPath, 0777)
    }
    // Go through each item in the bpm memory cache. There is suppose to only be one item per dependency
    for depName := range r.Items {
        fmt.Println("Processing cached dependency", depName)
        depItem := r.Items[depName];
        nodeModulesItemPath := path.Join(nodeModulesPath, depName);
        // Delete existing copy in node_modules
        os.RemoveAll(nodeModulesItemPath);

        // Copy the library to the node_modules folder
        copyDir := CopyDir{Exclude:Options.ExcludeFileList}
        err := copyDir.Copy(depItem.Path, nodeModulesItemPath);
        if err != nil {
            fmt.Println("Error: Failed to copy module to node_modules folder")
            fmt.Println(err)
            return err;
        }

        // Perform the npm install and pass the url of the dependency. npm install ./node_modules/mydep
        npm := NpmCommands{Path: path.Join(nodeModulesPath, "..")}
        err = npm.InstallUrl(path.Join("./node_modules", depName))
        if err != nil {
            fmt.Println("Error: Failed to npm install module", depName)
            return err;
        }
    }
    return nil;
}

func (r *ModuleCache) Add(item ModuleCacheItem, resolveVersion bool, resolveGitPath string) (bool, error) {
    existingItem, exists := r.Items[item.Name];

    if exists && resolveVersion && Options.ConflictResolutionType == "revisionlist" {
        fmt.Println("Attempting to determine which commit is the ancestor...")
        // If commitB is printed, then commitA is an ancestor of commit B
        //"git rev-list <commitA> | grep $(git rev-parse <commitB>)"
        git := GitCommands{Path: resolveGitPath}
        result := git.DetermineAncestor(item.Commit, existingItem.Commit)
        if result == item.Commit {
            fmt.Println("The commit " + item.Commit + " is an ancestor of the existing cache item. Replacing existing item with new item.")
        } else {
            return false, nil
        }
    } else if exists && resolveVersion && Options.ConflictResolutionType == "versioning" {
        v1, err := semver.Make(existingItem.Version)
        if err != nil {
            fmt.Println("Error: There was a problem reading the version")
            return false, nil;
        }
        v2, err := semver.Make(item.Version)
        if err != nil {
            fmt.Println("Error: There was a problem reading the version")
            return false, nil;
        }
        versionCompareResult := v1.Compare(v2);
        if versionCompareResult == -1 {
            fmt.Println("Ignoring lower version of ", item.Name);
            return false, nil;
        } else if versionCompareResult == 0 && strings.Compare(existingItem.Commit, item.Commit) != 0 {
            fmt.Println("Conflict. The version number is the same, but the commit hash is different. Ignoring")
            return false, nil
        } else if versionCompareResult == 1 {
            fmt.Println("The version number is greater and this version of the module will be used.")
        } else {
            // The version number is the same...
            return false, nil;
        }
    }
    r.Items[item.Name] = item;
    return true, nil;
}
