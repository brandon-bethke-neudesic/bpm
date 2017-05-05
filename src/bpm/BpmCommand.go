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
    Options.Command = "bpm"
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
        NewInstallCommand(),
        NewLsCommand(),
        NewUpdateCommand(),
        NewUninstallCommand(),
        NewVersionCommand(),
    )

    pf := cmd.PersistentFlags();

    pf.StringVar(&Options.UseRemoteName, "remote", "origin", "")
    pf.StringVar(&Options.UseRemoteUrl, "remoteurl", "", "")
    pf.StringVar(&Options.PackageManager, "pkgm", "npm", "")
    pf.BoolVar(&Options.SkipNpmInstall, "skipnpm", false, "")
    Options.WorkingDir, _ = os.Getwd();

    if strings.Index(Options.UseLocalPath, ".") == 0 || strings.Index(Options.UseLocalPath, "..") == 0 {
        Options.UseLocalPath = path.Join(Options.WorkingDir, Options.UseLocalPath)
    }
    return cmd
}

