package main;

import (
    "os"
    "fmt"
    "strings"
    "errors"
    "path"
)

type UninstallCommand struct {
}

func (cmd *UninstallCommand) Name() string {
    return "uninstall"
}

func (cmd *UninstallCommand) Execute() (error) {

    if _, err := os.Stat(Options.BpmFileName); os.IsNotExist(err) {
        fmt.Println("Error: The bpm file does not exist in the current directory.");
        return err;
    }

    fmt.Println("Reading",Options.BpmFileName,"...")
    bpm := BpmData{}
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
        fmt.Println(err);
        return err;
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }

    bpmModuleName := ""
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "uninstall" });
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        bpmModuleName = os.Args[index + 1];
    }

    if bpmModuleName == "" {
        msg := "Error: a module name must be specified"
        fmt.Println(msg)
        return errors.New(msg)
    } else {
        _, exists := bpm.Dependencies[bpmModuleName];
        if exists {
            delete(bpm.Dependencies, bpmModuleName)
            workingPath,_ := os.Getwd();
            npm := NpmCommands{Path: workingPath}
            err = npm.Uninstall(bpmModuleName)
            if err != nil {
                fmt.Println("Error: Failed to npm uninstall module", bpmModuleName)
                return err;
            }
            itemPath := path.Join(Options.BpmCachePath, bpmModuleName);
            os.RemoveAll(itemPath)
            bpm.IncrementVersion();
            bpm.WriteFile(path.Join(workingPath, Options.BpmFileName));
        } else {
            fmt.Println(bpmModuleName, "is not a dependency")
        }
    }
    return nil;
}
