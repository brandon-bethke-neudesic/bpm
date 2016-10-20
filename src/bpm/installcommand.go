package main;

import (
    "os"
    "fmt"
    "strings"
    "path"
    "bpmerror"
)


type InstallCommand struct {
}

func (cmd *InstallCommand) Name() string {
    return "install"
}

func (cmd *InstallCommand) installNew(moduleUrl string, moduleCommit string) (error) {
    bpm := BpmData{};
    err := Options.DoesBpmFileExist();
    if err != nil {
        return err;
    }
    err = bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }
    Options.EnsureBpmCacheFolder();

    itemRemoteUrl, err := MakeRemoteUrl(moduleUrl);
    if err != nil {
        return err;
    }
    var moduleBpm *BpmData;
    if len(strings.TrimSpace(moduleCommit)) == 0 {
        moduleCommit = "master";
    }
    moduleBpm, cacheItem, err := ProcessRemoteModule(itemRemoteUrl, moduleCommit)
    if err != nil {
        return err;
    }
    moduleCache.AddLatest(cacheItem)
    err = ProcessDependencies(moduleBpm, itemRemoteUrl)
    if err != nil {
        return err;
    }
    moduleCache.Trim();

    if !Options.SkipNpmInstall{
        err := moduleCache.Install()
        if err != nil {
            return err;
        }
    }

    newItem := BpmDependency{Url:moduleUrl, Commit:cacheItem.Commit};
    bpm.Dependencies[moduleBpm.Name] = newItem;

    bpm.IncrementVersion();
    bpm.WriteFile(path.Join(workingPath, Options.BpmFileName))

    return nil;
}

func (cmd *InstallCommand) build(installItem string) (error) {
    err := Options.DoesBpmFileExist();
    if err != nil {
        return err;
    }
    fmt.Println("Reading",Options.BpmFileName,"...")
    bpm := BpmData{}
    err = bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }
    fmt.Println("Validating bpm.json")
    err = bpm.Validate();
    if err != nil {
        return err;
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }
    Options.EnsureBpmCacheFolder();
    newBpm := bpm.Clone(installItem);
    fmt.Println("Processing all dependencies for", bpm.Name, "version", bpm.Version);
    err = ProcessDependencies(newBpm, "")
    if err != nil {
        return err;
    }
    moduleCache.Trim();
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
    newCommit := ""
    installItem := "";
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        installItem = os.Args[index + 1];
        if len(os.Args) > index + 2 && strings.Index(os.Args[index + 2], "--") != 0 {
            newCommit = os.Args[index + 2];
        }
    }

    if installItem != "" && newCommit != "" || strings.HasSuffix(installItem, ".git") {
        return cmd.installNew(installItem, newCommit);
    }
    return cmd.build(installItem);
}
