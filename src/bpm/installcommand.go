package main;

import (
    "os"
    "fmt"
    "strings"
    "path"
    "net/url"
    "github.com/blang/semver"
)


type InstallCommand struct {
    BpmModuleName string
    GitRemote string
    LocalPath string
}


func (cmd *InstallCommand) installNew(moduleUrl string, moduleCommit string) (error) {
    bpm := BpmData{};
    if _, err := os.Stat(Options.BpmFileName); os.IsNotExist(err) {
        fmt.Println("Error: The bpm file does not exist. Please use the new command to create the file");
        return err;
    } else {
        err := bpm.LoadFile(Options.BpmFileName);
        if err != nil {
            fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
            fmt.Println(err);
            return err;
        }
    }

    git := GitCommands{Path:workingPath}
    tmpUrl, err := git.GetRemoteUrl(cmd.GitRemote);
    if err != nil {
        fmt.Println("Error: There was a problem getting the remote url", cmd.GitRemote)
        return err;
    }

    remoteUrl, err := url.Parse(tmpUrl)
    if err != nil {
        fmt.Println(err);
        return err;
    }
    itemRemoteUrl := moduleUrl;
    if strings.Index(moduleUrl, "http") != 0 {
        itemRemoteUrl = remoteUrl.Scheme + "://" + path.Join(remoteUrl.Host, remoteUrl.Path, moduleUrl)
    }
    moduleBpm := BpmData{};

    if len(strings.TrimSpace(moduleCommit)) == 0 {
        itemPath := path.Join(Options.BpmCachePath, "xx_temp_xx", "xx_temp_xx")
        os.RemoveAll(path.Join(itemPath, ".."));
        os.MkdirAll(itemPath, 0777)
        git = GitCommands{Path:itemPath}
        err = git.InitAndFetch(itemRemoteUrl)
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
        git = GitCommands{Path:itemClonePath}
        err = git.InitAndCheckoutCommit(itemRemoteUrl, moduleCommit)
        if err != nil {
            fmt.Println("Error: There was an issue initializing the repository for dependency", moduleBpm.Name)
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
    for depName, v := range moduleBpm.Dependencies {
        GetDependencies(depName, v, itemRemoteUrl)
    }
    moduleCache.Trim();

    if !Options.SkipNpmInstall{
        err = moduleCache.NpmInstall()
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

func (cmd *InstallCommand) build() (error) {
    if _, err := os.Stat(Options.BpmFileName); os.IsNotExist(err) {
        fmt.Println(Options.BpmFileName, "Error: The bpm file does not exist in the current directory.");
        return err;
    }

    fmt.Println("Reading",Options.BpmFileName,"...")
    bpm := BpmData{}
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
        fmt.Println(err);
        return err;
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }

    if _, err := os.Stat(Options.BpmCachePath); os.IsNotExist(err) {
        os.Mkdir(Options.BpmCachePath, 0777)
    }

    if strings.TrimSpace(bpm.Name) == "" {
        fmt.Println("Error: There must be a name field in the bpm")
        return err;
    }

    if strings.TrimSpace(bpm.Version) == "" {
        fmt.Println("Error: There must be a version field in the bpm")
        return err;
    }

    fmt.Println("Processing all dependencies for", bpm.Name, "version", bpm.Version);
    if cmd.LocalPath != "" {
        GetDependenciesLocal(bpm)
    } else {
        for depName, v := range bpm.Dependencies {
            err := GetDependencies(depName, v, "")
            if err != nil {
                fmt.Println("Error: There was an issue processing dependency", depName)
                fmt.Println(err)
                return err;
            }
        }
    }
    moduleCache.Trim();
    if !Options.SkipNpmInstall {
        err = moduleCache.NpmInstall()
        if err != nil {
            return err;
        }
    }
    return nil;
}

func (cmd *InstallCommand) Execute() (error) {
    cmd.GitRemote = GetRemote(os.Args);
    cmd.LocalPath = GetLocal(os.Args);
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "install" });
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        newUrl := os.Args[index + 1];
        newCommit := ""
        if len(os.Args) > index + 2 && strings.Index(os.Args[index + 2], "--") != 0 {
            newCommit = os.Args[index + 2];
        }
        return cmd.installNew(newUrl, newCommit);
    } else {
        return cmd.build();
    }
    return nil;
}
