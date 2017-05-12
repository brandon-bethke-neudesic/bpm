package main;

import (
    "os"
    "github.com/spf13/cobra"
    "fmt"
    "errors"
    "bpmerror"
)

type InstallCommand struct {
    Args []string
    Name string
}

func (cmd *InstallCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    return nil;
}

func (cmd *InstallCommand) Execute() (error) {
    bpm := BpmModules{}
    err := bpm.Load(Options.BpmCachePath);
    if err != nil {
        return err;
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }

    name := cmd.Name;

    if name != "" && bpm.HasDependency(name){
        err = bpm.Dependencies[name].Install();
        if err != nil {
            return err;
        }
    } else if name != "" {
        return errors.New("Error: There is no dependency named " + name)
    } else {
        fmt.Println("Installing all dependencies for " + Options.RootComponent);
        for _, item := range bpm.Dependencies {
            err = item.Install();
            if err != nil {
                return err;
            }
        }
    }

    err = moduleCache.Install()
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue performing npm install on the dependencies")
    }
    return nil;
}

func NewInstallCommand() *cobra.Command {
    myCmd := &InstallCommand{}
    cmd := &cobra.Command{
        Use:   "install [NAME]",
        Short: "install the specified component",
        Long:  "install the specified component",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
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
