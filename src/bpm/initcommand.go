package main;

import (
    "fmt"
    "path"
    "errors"
    "os"
)

type InitCommand struct {
}

func (cmd *InitCommand) Name() string {
    return "init"
}

func (cmd *InitCommand) Execute() (error) {
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "init" });
    if len(os.Args) <= index + 1 {
        msg := "Error: incorrect usage. module name must be specified";
        fmt.Println(msg)
        return errors.New(msg)
    }
    bpmModuleName := os.Args[index + 1];
    bpm := BpmData{Name:bpmModuleName, Version:"1.0.0", Dependencies:make(map[string]BpmDependency)};
    err := bpm.WriteFile(path.Join(workingPath,Options.BpmFileName))
    if err == nil {
        fmt.Println("Created bpm file: ")
        fmt.Println(bpm.String())
    }
    return nil;
}
