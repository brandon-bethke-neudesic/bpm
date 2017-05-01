package main;

import (
    "fmt"
    "os"
    "bpmerror"
)

type InitCommand struct {
}

func (cmd *InitCommand) Name() string {
    return "init"
}

func (cmd *InitCommand) Execute() (error) {
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "init" });
    if len(os.Args) <= index + 1 {
        return bpmerror.New(nil, "Error: incorrect usage. module name must be specified")
    }
    bpmModuleName := os.Args[index + 1];
    bpm := BpmData{Name:bpmModuleName, Version:"1.0.0", Dependencies:make(map[string]*BpmDependency)};
    err := bpm.WriteFile(Options.BpmFileName)
    if err != nil {
        return err;
    }
    fmt.Println("Created bpm file: ")
    fmt.Println(bpm.String())
    return nil;
}
