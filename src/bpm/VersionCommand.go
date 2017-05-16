package main;

import (
    "fmt"
    "github.com/spf13/cobra"
)

var TpmVersion = "2.0.1"
// The build number should be set from ldsflags 
// go install -ldflags "-X main.TpmBuildNumber=$BUILD_NUMBER" tpm
var TpmBuildNumber string

func NewVersionCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "version",
        Short: "Display the bpm version number",
        Long:  "Display the bpm version number",
        Run: func(cmd *cobra.Command, args []string) {
        	version := TpmVersion;
        	if TpmBuildNumber != "" {
	        	version = version + "-" + TpmBuildNumber
        	}
            fmt.Println(version)
        },
    }
    return cmd
}