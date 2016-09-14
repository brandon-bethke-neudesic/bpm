package main;

import (
    "os"
    "fmt"
    "path"
    "strings"
    "sort"
)

type LsCommand struct {
    BpmModuleName string
    GitRemote string
    LocalPath string
}

func IndentAndPrintTree(indentLevel int, mytext string){
    text := "|"
    for i := 0; i < indentLevel; i++ {
        text = "|  " + text;
	}
    fmt.Println(text + mytext)
}

func IndentAndPrint(indentLevel int, mytext string){
    text := "   "
    for i := 0; i < indentLevel; i++ {
        text = "   " + text;
	}
    fmt.Println(text + mytext)
}

func PrintDependencies(bpm BpmData, indentLevel int) {
    // Sort the dependency keys so the dependencies always print in the same order
    sortedKeys := make([]string, len(bpm.Dependencies))
    i := 0
    for k, _ := range bpm.Dependencies {
        sortedKeys[i] = k
        i++
    }
    sort.Strings(sortedKeys)
    for _, itemName := range sortedKeys {
        item := bpm.Dependencies[itemName]
        if itemName == bpm.Name {
            fmt.Println("Warning: Ignoring self dependency for", itemName)
            continue
        }
        if item.Url == "" {
            fmt.Println("Error: No url specified for " + itemName)
        }
        if item.Commit == "" {
            fmt.Println("Error: No commit specified for " + itemName)
        }
        IndentAndPrintTree(indentLevel, "")
        workingPath,_ := os.Getwd();
        itemPath := path.Join(Options.BpmCachePath, itemName)
        itemClonePath := path.Join(workingPath, itemPath, item.Commit)
        localPath := path.Join(Options.BpmCachePath, itemName, Options.LocalModuleName)
        if PathExists(localPath) {
            itemClonePath = localPath;
            IndentAndPrintTree(indentLevel, "--" + itemName + " @ [Local]")
        } else if !PathExists(itemClonePath) {
            IndentAndPrintTree(indentLevel, "--" + itemName + " @ " + item.Commit + " [MISSING]")
            continue
        } else {
            IndentAndPrintTree(indentLevel, "--" + itemName + " @ " + item.Commit + " [Installed]")
        }

        // Recursively get dependencies in the current dependency
        moduleBpm := BpmData{};
        moduleBpmFilePath := path.Join(itemClonePath, Options.BpmFileName)
        err := moduleBpm.LoadFile(moduleBpmFilePath);
        if err != nil {
            IndentAndPrint(indentLevel, "Error: Could not load the bpm.json file for dependency " + itemName)
            continue
        }

        if strings.TrimSpace(moduleBpm.Name) == "" {
            msg := "Error: There must be a name field in the bpm.json for " + itemName
            IndentAndPrint(indentLevel, msg)
            continue
        }

        if strings.TrimSpace(moduleBpm.Version) == "" {
            msg := "Error: There must be a version field in the bpm for" + itemName
            IndentAndPrint(indentLevel, msg)
            continue
        }
        PrintDependencies(moduleBpm, indentLevel + 1)
    }
    return;
}

func (cmd *LsCommand) Execute() (error) {
    cmd.GitRemote = GetRemote(os.Args);
    cmd.LocalPath = GetLocal(os.Args);
    if _, err := os.Stat(Options.BpmFileName); os.IsNotExist(err) {
        fmt.Println("Error: The bpm.json file does not exist.")
        return err;
    }

    bpm := BpmData{};
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        fmt.Println("Error: There was a problem loading the bpm file at", Options.BpmFileName)
        fmt.Println(err);
        return err;
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    fmt.Println("")
    fmt.Println(bpm.Name)
    PrintDependencies(bpm, 0)
    fmt.Println("")
    return nil;
}
