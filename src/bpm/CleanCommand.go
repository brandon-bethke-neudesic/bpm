package main;

import (
    "os"
    "fmt"
    "github.com/spf13/cobra"
)

type CleanCommand struct {
    Args []string
    Name string
}

func (cmd *CleanCommand) Execute() (error) {
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
            fmt.Println("Unimplemented", item)
        }

    } else {
        _, exists := bpm.Dependencies[cmd.Name];
        if !exists {
            fmt.Println("WARNING: " + cmd.Name + " is not a dependency");

            ls := &LsCommand{TreeOnly: true};
            ls.Execute();

            return nil;
        }

        fmt.Println("Unimplemented", bpm.Dependencies[cmd.Name])
    }
    return nil;
}

func (cmd *CleanCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    return nil;
}

func NewCleanCommand() *cobra.Command {
    myCmd := &CleanCommand{}
    cmd := &cobra.Command{
        Use:   "clean [ITEM]",
        Short: "",
        Long:  "",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "clean"
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
