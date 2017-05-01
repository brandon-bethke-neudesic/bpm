package main;

import (
    "fmt"
    "os"
    "path"
    "strings"
    "bpmerror"
    "errors"
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
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    bpm := BpmData{};
    err := bpm.LoadFile(Options.BpmFileName);
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

    fmt.Println("Processing dependencies for", bpm.Name, "version", bpm.Version);
    // Always process the keys sorted by name so the installation is consistent
    sortedKeys := bpm.GetSortedKeys();
    for _, updateModule := range sortedKeys {
        depItem := bpm.Dependencies[updateModule]
        // If a specific module name was specified then skip the others.
        if bpmModuleName != "" && bpmModuleName != updateModule {
            continue;
        }
        if Options.UseLocalPath != "" && strings.Index(depItem.Url, "http") == -1 {
            moduleSourceUrl := path.Join(Options.UseLocalPath, updateModule);
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

            newItem := &BpmDependency{Url: depItem.Url, Commit:"local"}
            bpm.Dependencies[updateModule] = newItem;

        } else {

            itemRemoteUrl, err := MakeRemoteUrl(depItem.Url)
            if err != nil {
                return err;
            }
            moduleBpm, cacheItem, err := ProcessModule(itemRemoteUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            fmt.Println("Processing dependencies for", cacheItem.Name, "version", moduleBpm.Version);
            if Options.UseParentUrl {
                err = ProcessDependencies(moduleBpm, itemRemoteUrl)
            } else {
                err = ProcessDependencies(moduleBpm, "")
            }
            if err != nil {
                return err;
            }
            newItem := &BpmDependency{Url: depItem.Url, Commit:"local"}
            bpm.Dependencies[updateModule] = newItem;
        }
    }
    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            return bpmerror.New(err, "Error: There was an issue performing npm install on the dependencies")
        }
    }
    return nil;
}
