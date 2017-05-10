package main;

import (
    "os"
    "fmt"
    "path"
    "bpmerror"
    "github.com/spf13/cobra"
)

type UninstallCommand struct {
    Args []string
    Name string
}

func (cmd *UninstallCommand) Remove(name string) (error) {
    workingPath,_ := os.Getwd();
    npm := NpmExec{Path: workingPath}
    fmt.Println("Uninstalling npm module " + name);
    err := npm.Uninstall(name)
    if err != nil {
        return bpmerror.New(err, "Error: Failed to npm uninstall module " + name)
    }
    itemPath := path.Join(Options.BpmCachePath, name);
    git := &GitExec{LogOutput:true}
    err = git.DeleteSubmodule(itemPath);
    if err != nil {
        return bpmerror.New(err, "Error: Could not delete the submodule at " + itemPath)
    }
    return nil;
}

func (cmd *UninstallCommand) Execute() (error) {
    Options.EnsureBpmCacheFolder();

    bpm := BpmModules{}
    err := bpm.Load(Options.BpmCachePath);
    if err != nil {
        return err;
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }

    if cmd.Name == "" {
        for _, item := range bpm.Dependencies {
            err = cmd.Remove(item.Name);
            if err != nil {
                return err;
            }
        }

    } else {
        _, exists := bpm.Dependencies[cmd.Name];
        if !exists {
            fmt.Println("WARNING: " + cmd.Name + " is not a dependency. Listing dependency tree.");

            ls := &LsCommand{TreeOnly: true};
            ls.Execute();

            return nil;
        }

        err = cmd.Remove(cmd.Name);
        if err != nil {
            return err;
        }
    }
    return nil;
}

func (cmd *UninstallCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    return nil;
}

func NewUninstallCommand() *cobra.Command {
    myCmd := &UninstallCommand{}
    cmd := &cobra.Command{
        Use:   "uninstall [NAME]",
        Short: "uninstall the specified component using npm and remove the item from the bpm_modules",
        Long:  "uninstall the specified component using npm and remove the item from the bpm_modules",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "uninstall"
            err := myCmd.Execute();
            if err != nil {
                fmt.Println(err);
                fmt.Println("Finished with errors");
                os.Exit(1)
            }
        },
    }
    return cmd
}
