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

func DetermineLocalCommitValue(source string) (string, error) {
    // Put the commit as local, since we don't really know what it should be.
    // It's expected that the bpm install will fail if the commit is 'local' which will indicate to the user that they
    // didn't finalize configuring the dependency
    // However if the repo has no changes, then grab the commit
    commit := "local"
    git := GitExec{Path:source}
    if Options.Finalize || !git.HasChanges() {
        var err error;
        commit, err = git.GetLatestCommit()
        if err != nil {
            return "", err;
        }
    }
    return commit, nil
}

func updateLocalBpm(bpm *BpmData, moduleSourceUrl string, itemName string, item *BpmDependency) error {
    // Only update and save the bpm if the recursive option is set.
    if !Options.Recursive {
        return nil;
    }
    commit, err := DetermineLocalCommitValue(moduleSourceUrl)
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue getting the latest commit for " + itemName)
    }
    newItem := &BpmDependency{Url: item.Url, Commit:commit}
    existingItem := bpm.Dependencies[itemName];
    if !existingItem.Equal(newItem) {
        bpm.Dependencies[itemName] = newItem;
        filePath := path.Join(Options.UseLocalPath, bpm.Name, Options.BpmFileName);
        err = bpm.IncrementVersion();
        if err != nil {
            return err;
        }
        err = bpm.WriteFile(filePath)
        if err != nil {
            return err;
        }
    }
    return nil;
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
            fmt.Println("Processing local dependency in", moduleSourceUrl)
            commit, err := DetermineLocalCommitValue(moduleSourceUrl)
            if err != nil {
                return bpmerror.New(err, "Error: There was an issue getting the latest commit for " + updateModule)
            }
            moduleBpm, cacheItem, err := ProcessLocalModule(moduleSourceUrl)
            if err != nil {
                return err;
            }
            moduleCache.Add(cacheItem)
            err = ProcessDependencies(moduleBpm, "", updateLocalBpm)
            if err != nil {
                return err;
            }

            newItem := &BpmDependency{Url: depItem.Url, Commit:commit}
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
            err = ProcessDependencies(moduleBpm, itemRemoteUrl, nil)
            if err != nil {
                return err;
            }
            newItem := &BpmDependency{Url: depItem.Url, Commit:cacheItem.Commit}
            bpm.Dependencies[updateModule] = newItem;
        }
    }
    moduleCache.Trim();
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
    err = bpm.WriteFile(path.Join(Options.WorkingDir, Options.BpmFileName));
    if err != nil {
        return err;
    }
    return nil;
}
