package main;

import (
    "fmt"
    "os"
    "bpmerror"
    "errors"
    "github.com/spf13/cobra"
)


type UpdateCommand struct {
    Args []string
    Name string
}

func (cmd *UpdateCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    return nil;
}

func (cmd *UpdateCommand) Execute() (error) {
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
        err = bpm.Dependencies[name].Update();
        if err != nil {
            return err;
        }
    } else if name != "" {
        return errors.New("Error: The item " + name + " has not been installed.")
    } else {
        fmt.Println("Updating all dependencies for " + Options.RootComponent);
        for _, item := range bpm.Dependencies {
            err = item.Update();
            if err != nil {
                return err;
            }
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

func NewUpdateCommand() *cobra.Command {
    myCmd := &UpdateCommand{}
    cmd := &cobra.Command{
        Use:   "update [NAME]",
        Short: "update the specified component",
        Long:  "update the specified component",
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

    flags := cmd.Flags();
    flags.StringVar(&Options.UseLocalPath, "root", "", "A relative local path where the dependent repos can be found. Ex: bpm install --root=..")
    flags.BoolVar(&Options.Deep, "deep", false, "Update all dependencies in the bpm modules hierachy");
    flags.BoolVar(&Options.SkipNpmInstall, "skipnpm", false, "Do not perform a npm install")
    flags.StringVar(&Options.UseRemoteName, "remote", "origin", "")

    return cmd
}
