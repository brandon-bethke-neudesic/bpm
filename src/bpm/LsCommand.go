package main;

import (
    "fmt"
    "path"
    "github.com/spf13/cobra"
    "os"
    "strings"
)

type LsCommand struct {
    TreeOnly bool
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

type WarningItem struct {
    Path string
    Name string
    Message string
}
var warnings []*WarningItem = make([]*WarningItem, 0);

func (cmd *LsCommand) PrintDependenciesTree(bpm *BpmModules, indentLevel int) {
    // Sort the dependency keys so the dependencies always print in the same order
    sorted := bpm.GetSortedKeys();
    for _, itemName := range sorted {
        item := bpm.Dependencies[itemName]
        cmd.IndentAndPrintTree(indentLevel, "")
        cmd.IndentAndPrintTree(indentLevel, "--[" + item.Name + "]")
        // Traverse the hierarchy for the current dependency
        moduleBpm := &BpmModules{};
        err := moduleBpm.Load(path.Join(item.Path, Options.BpmCachePath));
        if err != nil {
            cmd.IndentAndPrint(indentLevel, "Error: Could not load dependencies for " + item.Name)
            continue
        }
        cmd.PrintDependenciesTree(moduleBpm, indentLevel + 1)
    }
}

func (cmd *LsCommand) PrintDependencies(bpm *BpmModules, indentLevel int) {
    // Sort the dependency keys so the dependencies always print in the same order
    sorted := bpm.GetSortedKeys();
    for _, itemName := range sorted {
        item := bpm.Dependencies[itemName]
        cmd.IndentAndPrintTree(indentLevel, "")
        git := &GitExec{Path: item.Path}
        itemCommit, err := git.GetLatestCommit();
        if err != nil {
            itemCommit = "Unknown"
        }
        branch, err := git.GetCurrentBranch()
        if err != nil {
            branch = "Unknown"
        }
        cmd.IndentAndPrintTree(indentLevel, "--[" + item.Name + "] @ " + branch + " " + itemCommit)

        if git.HasChanges() {
            cmd.IndentAndPrintTree(indentLevel, "  WARNING: Contains uncommitted changes")
            warn := &WarningItem{Path: item.Path, Name: item.Name}
            warn.Message = "WARNING: [" + item.Path + "] - There are uncommitted changes.";
            warnings = append(warnings, warn)
        }

        localRemote := git.GetRemote("local");
        if localRemote != nil {
            err := git.Fetch("local");
            if err != nil {
                fmt.Println("Error fetching")
            }

            localGit := &GitExec{Path: localRemote.Url}
            localItemCommit, err := localGit.GetLatestCommit();
            if err != nil {
                localItemCommit = "Unknown"
            }
            localBranch, err := localGit.GetCurrentBranch()
            if err != nil {
                localBranch = "Unknown"
            }
            if localItemCommit != itemCommit {
                cmd.IndentAndPrintTree(indentLevel, "  WARNING: Commit is different. local/" + localBranch + " " + localItemCommit)
                warn := &WarningItem{Path: item.Path, Name: item.Name}
                warn.Message = "WARNING: [" + item.Path + "] - The commit in bpm_modules does not match the commit in the remote.";
                warnings = append(warnings, warn)
                output, err := git.Run("git show -s --pretty=\"%an -  %ad - %B\" " + itemCommit)
                if err != nil {
                    output = "Could not get commit information"
                }
                output = strings.TrimSpace(output);
                warn.Message = warn.Message + "\n         MODULE[" + itemCommit + "] - " + output

                output, err = git.Run("git show -s --pretty=\"%an -  %ad - %B\" local/" + branch)
                if err != nil {
                    output = "Could not get commit information"
                }
                output = strings.TrimSpace(output);
                warn.Message = warn.Message + "\n         REMOTE[" + localItemCommit + "] - " + output
            }
        } else {
            err := git.Fetch(Options.UseRemoteName);
            if err != nil {
                fmt.Println("Error fetching")
            }

            if branch == "Unknown" || branch == "HEAD" {
                branch = "master"
            }
            remoteItemCommit, err := git.GetLatestCommitRemote(Options.UseRemoteName + "/" + branch);
            if err != nil {
                remoteItemCommit = "Unknown"
            }
            if remoteItemCommit != itemCommit {
                cmd.IndentAndPrintTree(indentLevel, "  WARNING: Commit is different. " + Options.UseRemoteName + "/" + branch + " " + remoteItemCommit)
                warn := &WarningItem{Path: item.Path, Name: item.Name}
                warn.Message = "WARNING: [" + item.Path + "] - The commit in bpm_modules does not match the commit in the remote.";

                output, err := git.Run("git show -s --pretty=\"%an -  %ad - %B\" " + itemCommit)
                if err != nil {
                    output = "Could not get commit information"
                }
                output = strings.TrimSpace(output);
                warn.Message = warn.Message + "\n         MODULE[" + itemCommit + "] - " + output

                output, err = git.Run("git show -s --pretty=\"%an -  %ad - %B\" " + Options.UseRemoteName + "/" + branch)
                if err != nil {
                    output = "Could not get commit information"
                }
                output = strings.TrimSpace(output);
                warn.Message = warn.Message + "\n         REMOTE[" + remoteItemCommit + "] - " + output

                warnings = append(warnings, warn)
            }
        }

        // Traverse the hierarchy for the current dependency
        moduleBpm := &BpmModules{};
        err = moduleBpm.Load(path.Join(item.Path, Options.BpmCachePath));
        if err != nil {
            cmd.IndentAndPrint(indentLevel, "Error: Could not load dependencies for " + item.Name)
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
    fmt.Println("[" + Options.RootComponent + "]")
    if cmd.TreeOnly {
        cmd.PrintDependenciesTree(bpm, 0)
    } else {
        cmd.PrintDependencies(bpm, 0)
    }
    fmt.Println("")

    if len(warnings) > 0 {
        fmt.Println("***************")
    }
    for _, warning := range warnings {
        fmt.Println(warning.Message);
        fmt.Println();
    }

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
            err := myCmd.Execute();
            if err != nil {
                fmt.Println(err);
                fmt.Println("Finished with errors");
                os.Exit(1)
            }
        },
    }

    flags := cmd.Flags();
    flags.BoolVar(&myCmd.TreeOnly, "tree-only", false, "Print only the dependency hierarchy")
    flags.StringVar(&Options.UseRemoteName, "remote", "origin", "")
    return cmd;
}
