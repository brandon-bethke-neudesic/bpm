package main;

import (
    "fmt"
)

type VersionCommand struct {
}

func (cmd *VersionCommand) Name() string {
    return "version"
}

func (cmd *VersionCommand) Execute() (error) {
    fmt.Println("bpm version 1.0.14")
    return nil
}
