package main;

import (
    "fmt"
    "path"
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

func (cmd *LsCommand) PrintDependencies(bpm *BpmModules, indentLevel int) {
    // Sort the dependency keys so the dependencies always print in the same order
    sortedKeys := bpm.GetSortedKeys();
    for _, itemName := range sortedKeys {
        cmd.IndentAndPrintTree(indentLevel, "")
        itemPath := path.Join(Options.BpmCachePath, itemName)
        cmd.IndentAndPrintTree(indentLevel, "--" + itemName);
        // Recursively get dependencies in the current dependency
        moduleBpm := &BpmModules{};
        err := moduleBpm.Load(path.Join(itemPath, Options.BpmCachePath));
        if err != nil {
            cmd.IndentAndPrint(indentLevel, "Error: Could not load the bpm.json file for dependency " + itemName)
            continue
        }
        cmd.PrintDependencies(moduleBpm, indentLevel + 1)
    }
    return;
}

func (cmd *LsCommand) Execute() (error) {
    bpm := &BpmModules{}
    err := bpm.Load(Options.BpmCachePath);
    if err != nil {
        return err;
    }

    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    fmt.Println("")
    fmt.Println(Options.RootComponent)
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
