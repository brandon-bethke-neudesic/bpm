package main;

import (
    "errors"
    "fmt"
    "os"
    "path"
    "net/url"
    "strings"
    "github.com/blang/semver"
)


type UpdateCommand struct {
    GitRemote string
    LocalPath string
}

func (cmd *UpdateCommand) Execute() (error) {
    cmd.GitRemote = GetRemote(os.Args);
    cmd.LocalPath = GetLocal(os.Args);

    bpmUpdated := false;
    bpmModuleName := ""
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "update" });
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        bpmModuleName = os.Args[index + 1];
    }

    if _, err := os.Stat(Options.BpmFileName); os.IsNotExist(err) {
        fmt.Println("Error: The bpm.json file does not exist.")
        return err;
    }

    bpm := BpmData{};
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
        fmt.Println(err);
        return err;
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    depsToProcess := make(map[string]BpmDependency);

    if bpmModuleName == "" {
        for name, v := range bpm.Dependencies {
            depsToProcess[name] = v
        }
    } else {
        depItem, exists := bpm.Dependencies[bpmModuleName];
        if !exists {
            msg := "Error: Could not find module " + bpmModuleName + " in the dependencies";
            fmt.Println(msg)
            return errors.New(msg)
        }
        depsToProcess[bpmModuleName] = depItem;
    }

    for updateModule, depItem := range depsToProcess {
        if cmd.LocalPath != "" {
            theUrl := path.Join(cmd.LocalPath, updateModule);
            fmt.Println("Processing local dependency in", theUrl)
            git := GitCommands{Path:theUrl}
            newCommit, err := git.GetLatestCommit()
            if err != nil {
                fmt.Println("Error: There was an issue getting the latest commit for", updateModule)
                return err;
            }
            itemPath := path.Join(Options.BpmCachePath, updateModule, Options.LocalModuleName);
            os.RemoveAll(itemPath)
            os.MkdirAll(itemPath, 0777)
            copyDir := CopyDir{Exclude:Options.ExcludeFileList}
            err = copyDir.Copy(theUrl, itemPath);
            if err != nil {
                fmt.Println("Error: There was an issue trying to copy the cloned temp folder to the named folder for", updateModule)
                fmt.Println(err)
                return err;
            }
            //os.RemoveAll(itemPath);

            cacheItem := ModuleCacheItem{Name:updateModule, Path: itemPath}
            moduleCache.Add(cacheItem, false, "")

            moduleBpm := BpmData{};
            moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName);
            err = moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                fmt.Println("Error: Could not load the bpm.json file for dependency", updateModule)
                fmt.Println(err);
                return err
            }
            GetDependenciesLocal(moduleBpm)

            newItem := BpmDependency{Url: depItem.Url, Commit:newCommit}
            existingItem := bpm.Dependencies[updateModule];
            if !existingItem.Equal(newItem) {
                bpmUpdated = true;
            }
            bpm.Dependencies[updateModule] = newItem;

        } else {
            git := GitCommands{Path:workingPath}
            theUrl, err := git.GetRemoteUrl(cmd.GitRemote)
            if err != nil {
                fmt.Println("Error: There was a problem getting the remote url", cmd.GitRemote)
                return err;
            }
            remoteUrl, err := url.Parse(theUrl)
            if err != nil {
                fmt.Println("Error: There was a problem parsing the remote url", theUrl)
                fmt.Println(err);
                return err;
            }
            itemRemoteUrl := depItem.Url;
            if strings.Index(depItem.Url, "http") != 0 {
                itemRemoteUrl = remoteUrl.Scheme + "://" + path.Join(remoteUrl.Host, remoteUrl.Path, depItem.Url)
            }

            itemPath := path.Join(Options.BpmCachePath, updateModule, "xx_temp_xx")
            os.RemoveAll(itemPath)
            os.MkdirAll(itemPath, 0777)
            git = GitCommands{Path:itemPath}
            err = git.InitAndFetch(itemRemoteUrl)
            if err != nil {
                fmt.Println(err)
                return err;
            }
            err = git.Checkout("master")
            if err != nil {
                fmt.Println("Error: There was an issue retrieving the repository for", updateModule)
                return err;
            }

            err = git.SubmoduleUpdate(true, true)
            if err != nil {
                fmt.Println("Error: Could not update submodules for", updateModule)
                return err;
            }

            newCommit, err := git.GetLatestCommit()
            if err != nil {
                fmt.Println("Error: There was an issue getting the latest commit for", updateModule)
                return err;
            }

            moduleBpm := BpmData{};
            moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName);
            err = moduleBpm.LoadFile(moduleBpmFilePath);
            if err != nil {
                fmt.Println("Error: Could not load the bpm.json file for dependency", updateModule)
                fmt.Println(err);
                return err;
            }
            itemNewPath := path.Join(Options.BpmCachePath, updateModule, newCommit);
            os.RemoveAll(itemNewPath)
            copyDir := CopyDir{Exclude:Options.ExcludeFileList}
            err = copyDir.Copy(itemPath, itemNewPath);
            if err != nil {
                fmt.Println("Error: There was an issue trying to copy the cloned temp folder to the named folder for", updateModule)
                fmt.Println(err)
                return err;
            }

            moduleBpmVersion, err := semver.Make(moduleBpm.Version);
            if err != nil {
                fmt.Println("Error: Could not read the version field from the bpm file for dependency", updateModule);
                return err;
            }

            cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: newCommit, Path: itemNewPath}
            moduleCache.Add(cacheItem, false, "")
            os.RemoveAll(itemPath);
            for depName, v := range moduleBpm.Dependencies {
                GetDependencies(depName, v)
            }
            newItem := BpmDependency{Url: depItem.Url, Commit:newCommit}
            existingItem := bpm.Dependencies[updateModule];
            if !existingItem.Equal(newItem) {
                bpmUpdated = true;
            }
            bpm.Dependencies[updateModule] = newItem;
        }
    }
    moduleCache.Trim();
    err = moduleCache.NpmInstall()
    if err != nil {
        fmt.Println("Error: There was an issue performing npm install on the dependencies")
        return err;
    }

    if bpmUpdated {
        bpm.IncrementVersion();
        bpm.WriteFile(path.Join(workingPath, Options.BpmFileName));
    }
    return nil;
}
