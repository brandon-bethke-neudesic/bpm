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
    fmt.Println("2.0.0")
    return nil
}
