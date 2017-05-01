package main;

import (
    "fmt"
    "path"
    "strings"
    "bpmerror"
    "errors"
    "github.com/spf13/cobra"
    "os"
)

type LsCommand struct {
    Args []string
}


func (cmd *LsCommand) IndentAndPrintTree(indentLevel int, mytext string){
    text := "|"
    for i := 0; i < indentLevel; i++ {
        text = "|  " + text;
	}
    fmt.Println(text + mytext)
}

func (cmd *LsCommand) IndentAndPrint(indentLevel int, mytext string){
    text := "   "
    for i := 0; i < indentLevel; i++ {
        text = "   " + text;
	}
    fmt.Println(text + mytext)
}

func (cmd *LsCommand) PrintDependencies(bpm BpmData, indentLevel int) {
    // Sort the dependency keys so the dependencies always print in the same order
    sortedKeys := bpm.GetSortedKeys();
    for _, itemName := range sortedKeys {
        item := bpm.Dependencies[itemName]
        if itemName == bpm.Name {
            fmt.Println("Warning: Ignoring self dependency for", itemName)
            continue
        }
        if item.Url == "" {
            fmt.Println("Error: No url specified for " + itemName)
        }
        cmd.IndentAndPrintTree(indentLevel, "")
        itemPath := path.Join(Options.BpmCachePath, itemName)
        if PathExists(itemPath) {
            cmd.IndentAndPrintTree(indentLevel, "--" + itemName + " [Installed]")
        } else {
            cmd.IndentAndPrintTree(indentLevel, "--" + itemName + " [MISSING]")
        }

        // Recursively get dependencies in the current dependency
        moduleBpm := BpmData{};
        moduleBpmFilePath := path.Join(itemPath, Options.BpmFileName)
        err := moduleBpm.LoadFile(moduleBpmFilePath);
        if err != nil {
            cmd.IndentAndPrint(indentLevel, "Error: Could not load the bpm.json file for dependency " + itemName)
            continue
        }

        if strings.TrimSpace(moduleBpm.Name) == "" {
            msg := "Error: There must be a name field in the bpm.json for " + itemName
            cmd.IndentAndPrint(indentLevel, msg)
            continue
        }

        if strings.TrimSpace(moduleBpm.Version) == "" {
            msg := "Error: There must be a version field in the bpm for" + itemName
            cmd.IndentAndPrint(indentLevel, msg)
            continue
        }
        cmd.PrintDependencies(moduleBpm, indentLevel + 1)
    }
    return;
}

func (cmd *LsCommand) Execute() (error) {
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    bpm := BpmData{};
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(err, "Error: There was a problem loading the bpm.json file")
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    fmt.Println("")
    fmt.Println(bpm.Name)
    cmd.PrintDependencies(bpm, 0)
    fmt.Println("")
    return nil;
}

func (cmd *LsCommand) Initialize() (error) {
    return nil;
}

func NewLsCommand() *cobra.Command {
    myCmd := &LsCommand{}
    cmd := &cobra.Command{
        Use:   "ls",
        Short: "list the dependencies",
        Long:  "list the dependencies",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "ls"
            err := myCmd.Execute();
            if err != nil {
                fmt.Println(err);
                fmt.Println("Finished with errors");
                os.Exit(1)
            }
        },
    }
    return cmd
}
