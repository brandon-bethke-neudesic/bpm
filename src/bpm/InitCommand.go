package main;

import (
    "fmt"
    "os"
    "bpmerror"
    "github.com/spf13/cobra"
)

type InitCommand struct {
    Args []string
    Name string
}

func (cmd *InitCommand) Execute() (error) {
    bpm := BpmData{Name:cmd.Name, Version:"1.0.0", Dependencies:make(map[string]*BpmDependency)};
    err := bpm.WriteFile(Options.BpmFileName)
    if err != nil {
        return err;
    }
    fmt.Println("Created bpm file: ")
    fmt.Println(bpm.String())
    return nil;
}

func (cmd *InitCommand) Initialize() (error) {
    if len(cmd.Args) == 0 {
        return bpmerror.New(nil, "Error: A name is required. Ex: bpm init NAME");
    }
    cmd.Name = cmd.Args[0];
    return nil;
}

func NewInitCommand() *cobra.Command {
    myCmd := &InitCommand{}
    cmd := &cobra.Command{
        Use:   "init NAME",
        Short: "create a bpm.json file",
        Long:  "create a bpm.json file",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "init"
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
