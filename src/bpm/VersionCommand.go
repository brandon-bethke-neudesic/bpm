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
            fmt.Println("2.0.1")
        },
    }
    return cmd
}