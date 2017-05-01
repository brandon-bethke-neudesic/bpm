package main;

import (
    "fmt"
    "github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "version",
        Short: "Display the bpm version number",
        Long:  "Display the bpm version number",
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "version"
            fmt.Println("2.0.0")
        },
    }
    return cmd
}