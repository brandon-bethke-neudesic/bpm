package main;

import (
    "fmt"
    "os"
    "path"
    "strings"
    "bpmerror"
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
        return err;
    }
    bpm := BpmData{};
    err = bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(nil, "Error: There was a problem loading the bpm.json file")
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    bpmModuleName := cmd.getUpdateModuleName()
    if bpmModuleName != "" && !bpm.HasDependency(bpmModuleName) {
        return bpmerror.New(err, "Error: Could not find module " + bpmModuleName + " in the dependencies")
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
                return bpmerror.New(err, "Error: There was an issue getting the latest commit for " + updateModule)
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
    err = moduleCache.Trim();
    if err != nil {
        return err;
    }
    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            return bpmerror.New(err, "Error: There was an issue performing npm install on the dependencies")
        }
    }
    err = bpm.IncrementVersion();
    if err != nil {
        return err;
    }
    err = bpm.WriteFile(path.Join(workingPath, Options.BpmFileName));
    if err != nil {
        return err;
    }
    return nil;
}
