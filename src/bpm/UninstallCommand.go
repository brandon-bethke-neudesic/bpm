package main;

import (
    "os"
    "fmt"
    "path"
    "bpmerror"
    "errors"
    "github.com/spf13/cobra"
)

type UninstallCommand struct {
    Args []string
    Name string
}

func (cmd *UninstallCommand) Execute() (error) {
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    bpm := BpmData{}
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the " + Options.BpmFileName + " file")
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }

    _, exists := bpm.Dependencies[cmd.Name];
    if !exists {
        fmt.Println("Error: " + cmd.Name + " is not a dependency")
        return nil;
    }

    delete(bpm.Dependencies, cmd.Name)
    workingPath,_ := os.Getwd();
    npm := NpmExec{Path: workingPath}
    fmt.Println("Uninstalling npm module " + cmd.Name);
    err = npm.Uninstall(cmd.Name)
    if err != nil {
        return bpmerror.New(err, "Error: Failed to npm uninstall module " + cmd.Name)
    }
    itemPath := path.Join(Options.BpmCachePath, cmd.Name);
    os.RemoveAll(itemPath)
    bpm.IncrementVersion();
    bpm.WriteFile(path.Join(workingPath, Options.BpmFileName));
    return nil;
}

func (cmd *UninstallCommand) Initialize() (error) {
    if len(cmd.Args) == 0 {
        return errors.New("Error: A name is required. Ex: bpm uninstall NAME");
    }
    cmd.Name = cmd.Args[0];
    return nil;
}

func NewUninstallCommand() *cobra.Command {
    myCmd := &UninstallCommand{}
    cmd := &cobra.Command{
        Use:   "uninstall NAME",
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
