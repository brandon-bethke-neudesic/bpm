package main;

import (
    "github.com/blang/semver"
    "fmt"
    "strings"
    "os"
    "path"
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
        copyDir := CopyDir{Exclude:excludeFileList}
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

func (r *ModuleCache) Add(item ModuleCacheItem, resolveVersion bool) (error) {
    existingItem, exists := r.Items[item.Name];
    if exists && resolveVersion {
        v1, err := semver.Make(existingItem.Version)
        if err != nil {
            fmt.Println("Error: There was a problem reading the version")
            return nil;
        }
        v2, err := semver.Make(item.Version)
        if err != nil {
            fmt.Println("Error: There was a problem reading the version")
            return nil;
        }
        versionCompareResult := v1.Compare(v2);
        if versionCompareResult == -1 {
            fmt.Println("Ignoring lower version of ", item.Name);
            return nil;
        } else if versionCompareResult == 0 && strings.Compare(existingItem.Commit, item.Commit) != 0 {
            fmt.Println("Conflict. The version number is the same, but the commit hash is different. Ignoring")
            return nil;
        } else {
            fmt.Println("The version number is greater and this version of the module will be used.")
        }
    }
    r.Items[item.Name] = item;
    return nil;
}
