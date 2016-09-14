package main;

import (
    "os"
)

type CleanCommand struct {
}

func (cmd *CleanCommand) Execute() (error) {
    os.RemoveAll(Options.BpmCachePath);
    return nil;
}
