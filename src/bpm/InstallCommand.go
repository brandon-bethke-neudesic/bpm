package main;

import (
    "os"
    "fmt"
    "strings"
    "bpmerror"
    "errors"
    "github.com/spf13/cobra"
)


type InstallCommand struct {
    Args []string
    InstallItem string
    Commit string
}


func (cmd *InstallCommand) installNew(moduleUrl string, commit string) (error) {
    bpm := BpmData{};
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }
    Options.EnsureBpmCacheFolder();

    moduleSourceUrl, err := MakeRemoteUrl(moduleUrl);
    if err != nil {
        return err;
    }
    parentUrl := "";
    fmt.Println("Processing dependency at", moduleSourceUrl)
    moduleBpm, cacheItem, err := ProcessModule(moduleSourceUrl, commit)
    if err != nil {
        return err;
    }
    moduleCache.Add(cacheItem)
    if Options.UseParentUrl {
        parentUrl = moduleSourceUrl;
    }
    err = ProcessDependencies(moduleBpm, parentUrl)
    if err != nil {
        return err;
    }
    newItem := &BpmDependency{Url: moduleUrl}
    bpm.Dependencies[moduleBpm.Name] = newItem;

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

func (cmd *InstallCommand) Initialize() (error) {
    if len(cmd.Args) > 1 {
        cmd.InstallItem = cmd.Args[1];
    }
    return nil;
}

func (cmd *InstallCommand) Execute() (error) {
    if cmd.InstallItem != "" || strings.HasSuffix(cmd.InstallItem, ".git") {
        return cmd.installNew(cmd.InstallItem, cmd.Commit);
    }
    return cmd.build(cmd.InstallItem);
}


func NewInstallCommand() *cobra.Command {
    myCmd := &InstallCommand{}
    cmd := &cobra.Command{
        Use:   "install [URL]",
        Short: "install the specified item or all the dependencies",
        Long:  "install the specified item or all the dependencies",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "install"
            err := myCmd.Execute();
            if err != nil {
                fmt.Println(err);
                fmt.Println("Finished with errors");
                os.Exit(1)
            }
        },
    }

    flags := cmd.Flags();
    flags.StringVar(&myCmd.Commit, "commit", "latest", "The commit value")

    return cmd
}