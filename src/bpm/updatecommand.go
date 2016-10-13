package main;

import (
    "errors"
    "fmt"
    "os"
    "path"
    "strings"
    "github.com/blang/semver"
)


type UpdateCommand struct {
}

func (cmd *UpdateCommand) Name() string {
    return "update"
}

func (cmd *UpdateCommand) getUpdateModuleName() (string){
    // The next parameter after the 'uninstall' must to be the module name
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "update" });
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        return os.Args[index + 1];
    }
    return "";
}

func (cmd *UpdateCommand) Execute() (error) {
    err := Options.DoesBpmFileExist();
    if err != nil {
        fmt.Println(err);
        return err;
    }
    bpm := BpmData{};
    err = bpm.LoadFile(Options.BpmFileName);
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
    bpmModuleName := cmd.getUpdateModuleName()
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
        if Options.UseLocal != "" && strings.Index(depItem.Url, "http") == -1 {
            theUrl := path.Join(Options.UseLocal, updateModule);
            fmt.Println("Processing local dependency in", theUrl)
            // Put the commit as local, since we don't really know what it should be.
            // It's expected that the bpm install will fail if the commit is 'local' which will indicate to the user that they
            // didn't finalize configuring the dependency
            // However if the repo has no changes, then grab the commit
            git := GitCommands{Path:theUrl}
            newCommit := "local"
            if Options.Finalize || !git.HasChanges() {
                var err error;
                newCommit, err = git.GetLatestCommit()
                if err != nil {
                    fmt.Println("Error: There was an issue getting the latest commit for", updateModule)
                    return err;
                }
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
            err = ProcessDependencies(moduleBpm, "")
            if err != nil {
                return err;
            }

            newItem := BpmDependency{Url: depItem.Url, Commit:newCommit}
            bpm.Dependencies[updateModule] = newItem;

        } else {

            itemRemoteUrl, err := MakeRemoteUrl(depItem.Url)
            if err != nil {
                return err;
            }
            itemPath := path.Join(Options.BpmCachePath, updateModule, "xx_temp_xx")
            os.RemoveAll(itemPath)
            os.MkdirAll(itemPath, 0777)
            git := GitCommands{Path:itemPath}
            err = git.InitAndCheckout(itemRemoteUrl, "master")
            if err != nil {
                msg := "Error: There was an issue initializing the repository for dependency " + updateModule + " Url: " + itemRemoteUrl
                os.RemoveAll(itemPath)
                return errors.New(msg)
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
            err = ProcessDependencies(moduleBpm, itemRemoteUrl)
            if err != nil {
                return err;
            }
            newItem := BpmDependency{Url: depItem.Url, Commit:newCommit}
            bpm.Dependencies[updateModule] = newItem;
        }
    }
    moduleCache.Trim();
    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            fmt.Println("Error: There was an issue performing npm install on the dependencies")
            return err;
        }
    }
    bpm.IncrementVersion();
    bpm.WriteFile(path.Join(workingPath, Options.BpmFileName));
    return nil;
}
