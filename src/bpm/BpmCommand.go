package main;

import (
    "strings"
    "os"
    "path"
    "github.com/spf13/cobra"
)

type BpmCommand struct {
    Path string
    Args []string
}

func (cmd *BpmCommand) Description() string {
    return `
The Better Package Manager
`
}

func NewBpmCommand() (*cobra.Command) {
    myCmd := &BpmCommand{}
    cmd := &cobra.Command{
        Use:          "bpm",
        Short:        "The better package manager.",
        Long:         myCmd.Description(),
        SilenceUsage: false,
        SilenceErrors: true,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            return Options.Validate();
        },

    }
    cmd.AddCommand(
        NewAddCommand(),
        NewLsCommand(),
        NewUpdateCommand(),
        NewRemoveCommand(),
        NewVersionCommand(),
        NewStatusCommand(),
        NewInstallCommand(),
    )
    Options.WorkingDir, _ = os.Getwd();

    if strings.Index(Options.Local, ".") == 0 || strings.Index(Options.Local, "..") == 0 {
        Options.Local = path.Join(Options.WorkingDir, Options.Local)
    }
    return cmd
}

