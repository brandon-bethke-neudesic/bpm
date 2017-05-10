package main;

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
    "strings"
    "errors"
)

type StatusCommand struct {
    Args []string
    Name string
}

func (cmd *StatusCommand) Execute() (error) {
    Options.EnsureBpmCacheFolder();

    folders := strings.Split(cmd.Name, "/");

    newFolder := strings.Join(folders, "/bpm_modules/")
    newFolder = "bpm_modules/" + newFolder;

    fmt.Println(newFolder)

    git := &GitExec{Path:newFolder};
    err := git.Status();
    return err;
}

func (cmd *StatusCommand) Initialize() (error) {
    if len(cmd.Args) == 0 {
        return errors.New("Error: A name is required. Ex: bpm status bpmdep3/bpmdep1");
    }
    cmd.Name = cmd.Args[0];

    return nil;
}

func NewStatusCommand() *cobra.Command {
    myCmd := &StatusCommand{}
    cmd := &cobra.Command{
        Use:   "status NAME",
        Short: "perform a git status on an item in bpm_modules",
        Long:  "perform a git status on an item in bpm_modules. Ex: bpm status bpmdep3/bpmdep1",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "status"
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
