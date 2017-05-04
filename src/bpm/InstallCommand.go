package main;

import (
    "os"
    "fmt"
    "strings"
    "bpmerror"
    "errors"
    "github.com/spf13/cobra"
    "path/filepath"
    "net/url"
)


type InstallCommand struct {
    Args []string
    Component string
    Commit string
}

func (cmd *InstallCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Component = cmd.Args[0];
    }

    if strings.HasPrefix(cmd.Component, "http") && Options.UseLocalPath != "" {
        return bpmerror.New(nil, "Error: The root option should not be used with a full url path when installing a component. Ex: bpm install http://www.github.com/neudesic/myproject.git")
    }

    return nil;
}

func (cmd *InstallCommand) Execute() (error) {
    Options.EnsureBpmCacheFolder();

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

    save := false;
    name := cmd.Component
    if !bpm.HasDependency(name) && name != "" {
        newItem := &BpmDependency{}
        if strings.HasPrefix(cmd.Component, "http") {
            parsedUrl, err := url.Parse(cmd.Component);
            if err != nil {
                return bpmerror.New(err, "Error: Could not parse the component url");
            }
            _, name = filepath.Split(parsedUrl.Path)
            name = strings.Split(name, ".git")[0]
            newItem.Url = cmd.Component;
        }
        fmt.Println("Adding new component " + name)
        bpm.Dependencies[name] = newItem;
        save = true;
    }
    newBpm := bpm.Clone(name);
    fmt.Println("Processing dependencies for", bpm.Name, "version", bpm.Version);
    err = ProcessDependencies(newBpm)
    if err != nil {
        return err;
    }
    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            return err;
        }
    }

    if save {
        err = bpm.WriteFile(Options.BpmFileName)
        if err != nil {
            return err;
        }
    }
    return nil;
}


func NewInstallCommand() *cobra.Command {
    myCmd := &InstallCommand{}
    cmd := &cobra.Command{
        Use:   "install [COMPONENT]",
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
    flags.StringVar(&Options.UseLocalPath, "root", "", "")
    flags.StringVar(&myCmd.Commit, "commit", "latest", "The commit value")
    flags.StringVar(&Options.ConflictResolutionType, "resolution", "revisionlist", "")

    return cmd
}