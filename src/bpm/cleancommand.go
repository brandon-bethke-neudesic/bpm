package main;

import (
    "os"
)

type CleanCommand struct {
}

func (cmd *CleanCommand) Name() string {
    return "clean"
}

func (cmd *CleanCommand) Execute() (error) {
    os.RemoveAll(Options.BpmCachePath);
    return nil;
}
