package main;

import (
    "errors"
    "fmt"
    "os"
    "path"
    "strings"
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

    bpmModuleName := cmd.getUpdateModuleName()
    if bpmModuleName != "" && !bpm.HasDependency(bpmModuleName) {
        msg := "Error: Could not find module " + bpmModuleName + " in the dependencies";
        fmt.Println(msg)
        return errors.New(msg)
    }

    for updateModule, depItem := range bpm.Dependencies {
        // If a specific module name was specified then skip the others.
        if bpmModuleName != "" && bpmModuleName != updateModule {
            continue;
        }
        if Options.UseLocal != "" && strings.Index(depItem.Url, "http") == -1 {
            moduleSourceUrl := path.Join(Options.UseLocal, updateModule);
            fmt.Println("Processing local dependency in", moduleSourceUrl)
            commit, err := DetermineCommitValue(moduleSourceUrl)
            if err != nil {
                fmt.Println("Error: There was an issue getting the latest commit for", updateModule)
                fmt.Println(err)
                return err;
            }
            moduleBpm, cacheItem, err := ProcessLocalModule(moduleSourceUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            err = ProcessDependencies(moduleBpm, "")
            if err != nil {
                return err;
            }

            newItem := BpmDependency{Url: depItem.Url, Commit:commit}
            bpm.Dependencies[updateModule] = newItem;

        } else {

            itemRemoteUrl, err := MakeRemoteUrl(depItem.Url)
            if err != nil {
                return err;
            }
            moduleBpm, cacheItem, err := ProcessRemoteModule(itemRemoteUrl, "master")
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            err = ProcessDependencies(moduleBpm, itemRemoteUrl)
            if err != nil {
                return err;
            }
            newItem := BpmDependency{Url: depItem.Url, Commit:cacheItem.Commit}
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
