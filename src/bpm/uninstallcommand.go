package main;

import (
    "os"
    "fmt"
    "strings"
    "path"
    "bpmerror"
)

type UninstallCommand struct {
}

func (cmd *UninstallCommand) Name() string {
    return "uninstall"
}

func (cmd *UninstallCommand) getUninstallModuleName() (string){
    // The next parameter after the 'uninstall' must to be the module name
    index := SliceIndex(len(os.Args), func(i int) bool { return os.Args[i] == "uninstall" });
    if len(os.Args) > index + 1 && strings.Index(os.Args[index + 1], "--") != 0 {
        return os.Args[index + 1];
    }
    return "";
}

func (cmd *UninstallCommand) Execute() (error) {
    err := Options.DoesBpmFileExist();
    if err != nil {
        return err;
    }
    fmt.Println("Reading",Options.BpmFileName,"...")
    bpm := BpmData{}
    err = bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }

    uninstallModuleName := cmd.getUninstallModuleName();
    if uninstallModuleName == "" {
        return bpmerror.New(nil, "Error: a module name must be specified")
    }

    _, exists := bpm.Dependencies[uninstallModuleName];
    if !exists {
        fmt.Println(uninstallModuleName, "is not a dependency")
        return nil;
    }

    delete(bpm.Dependencies, uninstallModuleName)
    workingPath,_ := os.Getwd();
    npm := NpmCommands{Path: workingPath}
    err = npm.Uninstall(uninstallModuleName)
    if err != nil {
        return bpmerror.New(err, "Error: Failed to npm uninstall module " + uninstallModuleName)
    }
    itemPath := path.Join(Options.BpmCachePath, uninstallModuleName);
    os.RemoveAll(itemPath)
    bpm.IncrementVersion();
    bpm.WriteFile(path.Join(workingPath, Options.BpmFileName));
    return nil;
}
