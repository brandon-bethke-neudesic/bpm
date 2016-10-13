package main;

import (
    "os"
    "fmt"
    "strings"
    "path"
    "github.com/blang/semver"
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
        fmt.Println(err);
        return err;
    }
    err = bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
        fmt.Println(err);
        return err;
    }
    Options.EnsureBpmCacheFolder();

    itemRemoteUrl, err := MakeRemoteUrl(moduleUrl);
    if err != nil {
        fmt.Println(err)
        return err;
    }
    moduleBpm := BpmData{};

    if len(strings.TrimSpace(moduleCommit)) == 0 {
        itemPath := path.Join(Options.BpmCachePath, "xx_temp_xx", "xx_temp_xx")
        os.RemoveAll(path.Join(itemPath, ".."));
        os.MkdirAll(itemPath, 0777)
        git := GitCommands{Path:itemPath}
        err := git.InitAndFetch(itemRemoteUrl)
        if err != nil {
            fmt.Println(err)
            return err;
        }
        err = git.Checkout("master")
        if err != nil {
            fmt.Println("Error: There was an issue retrieving the repository.")
            return err;
        }
        moduleCommit, err = git.GetLatestCommit()
        if err != nil {
            fmt.Println(err)
            return err;
        }

        moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName);
        err = moduleBpm.LoadFile(moduleBpmFilePath);
        if err != nil {
            fmt.Println("Error: Could not load the bpm.json file for dependency at", moduleBpmFilePath)
            fmt.Println(err);
            return err;
        }
        moduleBpmVersion, err := semver.Make(moduleBpm.Version);
        if err != nil {
            fmt.Println("Error: Could not read the version");
            return err;
        }

        itemNewPath := path.Join(Options.BpmCachePath, moduleBpm.Name, moduleCommit);
        os.RemoveAll(itemNewPath)
        copyDir := CopyDir{Exclude:Options.ExcludeFileList}
        err = copyDir.Copy(itemPath, itemNewPath);
        if err != nil {
            fmt.Println(err)
            return err;
        }

        cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: moduleCommit, Path: itemNewPath}
        moduleCache.Add(cacheItem, true, itemPath)
        os.RemoveAll(path.Join(itemPath, ".."))
    } else {
        itemPath := path.Join(Options.BpmCachePath, "xx_temp_xx", moduleCommit)
        itemClonePath := path.Join(workingPath, itemPath, moduleCommit)
        os.RemoveAll(path.Join(itemClonePath, ".."));
        os.MkdirAll(itemClonePath, 0777)
        git := GitCommands{Path:itemClonePath}
        err := git.InitAndCheckout(itemRemoteUrl, moduleCommit)
        if err != nil {
            fmt.Println("Error: There was an issue initializing the repository for dependency", moduleBpm.Name, "Url:", itemRemoteUrl, "Commit:", moduleCommit)
            os.RemoveAll(itemClonePath)
            return err;
        }

        moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName);
        err = moduleBpm.LoadFile(moduleBpmFilePath);
        if err != nil {
            fmt.Println("Error: Could not load the bpm.json file for dependency at", moduleBpmFilePath)
            fmt.Println(err);
            return err;
        }

        moduleBpmVersion, err := semver.Make(moduleBpm.Version);
        if err != nil {
            fmt.Println("Could not read the version");
            return err;
        }

        itemNewPath := path.Join(Options.BpmCachePath, moduleBpm.Name, moduleCommit);
        copyDir := CopyDir{Exclude:Options.ExcludeFileList}
        err = copyDir.Copy(itemPath, itemNewPath);
        if err != nil {
            fmt.Println(err)
            return err;
        }

        cacheItem := ModuleCacheItem{Name:moduleBpm.Name, Version: moduleBpmVersion.String(), Commit: moduleCommit, Path: itemNewPath}
        moduleCache.Add(cacheItem, true, itemPath)
        os.RemoveAll(path.Join(itemPath, ".."))
    }
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

    newItem := BpmDependency{Url:moduleUrl, Commit:moduleCommit};
    bpm.Dependencies[moduleBpm.Name] = newItem;

    bpm.IncrementVersion();
    bpm.WriteFile(path.Join(workingPath, Options.BpmFileName))

    return nil;
}

func (cmd *InstallCommand) build(installItem string) (error) {
    err := Options.DoesBpmFileExist();
    if err != nil {
        fmt.Println(err);
        return err;
    }
    fmt.Println("Reading",Options.BpmFileName,"...")
    bpm := BpmData{}
    err = bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
        fmt.Println(err);
        return err;
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }
    Options.EnsureBpmCacheFolder();

    if strings.TrimSpace(bpm.Name) == "" {
        fmt.Println("Error: There must be a name field in the bpm")
        return err;
    }

    if strings.TrimSpace(bpm.Version) == "" {
        fmt.Println("Error: There must be a version field in the bpm")
        return err;
    }

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
