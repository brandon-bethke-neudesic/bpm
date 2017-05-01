package main;

import (
    "os"
    "fmt"
    "strings"
    "bpmerror"
    "path"
    "errors"
)


type InstallCommand struct {
}

func (cmd *InstallCommand) Name() string {
    return "install"
}

func (cmd *InstallCommand) installNew(moduleUrl string) (error) {
    bpm := BpmData{};
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }
    Options.EnsureBpmCacheFolder();

    if Options.UseLocalPath != "" && strings.Index(moduleUrl, "http") == -1 {
        moduleSourceUrl := path.Join(Options.UseLocalPath, moduleUrl);
        fmt.Println("Processing dependency at", moduleSourceUrl)
        moduleBpm, cacheItem, err := ProcessModule(moduleSourceUrl)
        if err != nil {
            return err;
        }
        moduleCache.Add(cacheItem)
        err = ProcessDependencies(moduleBpm, "")
        if err != nil {
            return err;
        }

        newItem := &BpmDependency{Url: moduleUrl, Commit:"local"}
        bpm.Dependencies[moduleBpm.Name] = newItem;

    } else {
        itemRemoteUrl, err := MakeRemoteUrl(moduleUrl)
        if err != nil {
            return err;
        }
        fmt.Println("Processing dependency at", itemRemoteUrl)
        moduleBpm, cacheItem, err := ProcessModule(itemRemoteUrl)
        if err != nil {
            return err;
        }
        moduleCache.Add(cacheItem)
        if Options.UseParentUrl {
            err = ProcessDependencies(moduleBpm, itemRemoteUrl)
        } else {
            err = ProcessDependencies(moduleBpm, "")
        }
        if err != nil {
            return err;
        }
        newItem := &BpmDependency{Url: moduleUrl, Commit:"local"}
        bpm.Dependencies[moduleBpm.Name] = newItem;
    }

    if !Options.SkipNpmInstall{
        err := moduleCache.Install()
        if err != nil {
            return err;
        }
    }

    err = bpm.IncrementVersion();
    if err != nil {
        return err;
    }
    err = bpm.WriteFile(Options.BpmFileName)
    if err != nil {
        return err;
    }
    return nil;
}

func (cmd *InstallCommand) build(installItem string) (error) {
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    bpm := BpmData{}
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }
    err = bpm.Validate();
    if err != nil {
        return bpmerror.New(err, "Error: There is a problem with the bpm.json file")
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }
    Options.EnsureBpmCacheFolder();
    newBpm := bpm.Clone(installItem);
    fmt.Println("Processing dependencies for", bpm.Name, "version", bpm.Version);
    err = ProcessDependencies(newBpm, "")
    if err != nil {
        return err;
    }
    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            return err;
        }
    }
    return nil;
}

func (cmd *InstallCommand) Execute() (error) {
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "install" });
    installItem := "";
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        installItem = os.Args[index + 1];
    }

    if installItem != "" || strings.HasSuffix(installItem, ".git") {
        return cmd.installNew(installItem);
    }
    return cmd.build(installItem);
}
