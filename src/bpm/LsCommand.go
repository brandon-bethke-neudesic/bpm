package main;

import (
    "fmt"
    "path"
    "github.com/spf13/cobra"
    "os"
    "strings"
    "sort"
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

type Warnings struct {
	Items map[string] *WarningItem
}

func (warnings *Warnings) GetSortedKeys() []string {
    sortedKeys := make([]string, len(warnings.Items))
    i := 0
    for k, _ := range warnings.Items {
        sortedKeys[i] = k
        i++
    }
    sort.Strings(sortedKeys)
    return sortedKeys;
}

func (warnings *Warnings) Add(dep *BpmDependency, message string) (*WarningItem) {
	if warnings.Items == nil {
		warnings.Items = make(map[string]*WarningItem, 0);
	}
	item, exists := warnings.Items[dep.Path];
	if !exists {
		item = &WarningItem{Name: dep.Name, Path: dep.Path};
		item.AddMessage(message);
		warnings.Items[dep.Path] = item;	
	} else {
		item.AddMessage(message);	
	}	
	return item;
}

type WarningItem struct {
    Name string
    Path string
    Messages []string
}

func (warn *WarningItem) AddMessage(message string) {
	if warn.Messages == nil {
		warn.Messages = make([]string, 0);
	}
	warn.Messages = append(warn.Messages, message);
}

var warnings = Warnings{};

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

    goInto := func(item *BpmDependency, indentLevel int){
        // Traverse the hierarchy for the current dependency
        moduleBpm := &BpmModules{};
        err := moduleBpm.Load(path.Join(item.Path, Options.BpmCachePath));
        if err != nil {
            cmd.IndentAndPrint(indentLevel, "Error: Could not load dependencies for " + item.Name);
            return;
        }
        cmd.PrintDependencies(moduleBpm, indentLevel);        
    }

    // Sort the dependency keys so the dependencies always print in the same order
    sorted := bpm.GetSortedKeys();
    
    for _, itemName := range sorted {
        item := bpm.Dependencies[itemName]
        cmd.IndentAndPrintTree(indentLevel, "")
        git := &GitExec{Path: item.Path}
        itemCommit, err := git.GetLatestCommit();
        if err != nil {
	        cmd.IndentAndPrintTree(indentLevel, "--[" + item.Name + "] Error: Could not get commit information for " + item.Path);
            warnings.Add(item, "Error: Could not get commit information for " + item.Path);
	        goInto(item, indentLevel + 1);
	        continue;	        
        }
        branch, err := git.GetCurrentBranch()
        if err != nil {
	        cmd.IndentAndPrintTree(indentLevel, "--[" + item.Name + "] Error: Could not get branch information for " + item.Path);
            warnings.Add(item, "Error: Could not get branch information for " + item.Path);
	        goInto(item, indentLevel + 1);
	        continue;
        }
        cmd.IndentAndPrintTree(indentLevel, "--[" + item.Name + "] @ " + branch + " " + itemCommit)

        if git.HasChanges() {
            cmd.IndentAndPrintTree(indentLevel, "  WARNING: Contains uncommitted changes")
            warnings.Add(item, "Warning: There are uncommitted changes. Run: 'bpm status " + strings.Replace(item.Path, "bpm_modules/", "", -1) + "'");
        }        
        commitMismatchWarning := "";
        printedModuleCommit := false;
        localRemote := git.GetRemote("local");
        if localRemote != nil {
            err := git.Fetch("local");
            if err != nil {
		        cmd.IndentAndPrintTree(indentLevel, "Error: Could not fetch information for remote local at " + localRemote.Url);
	            warnings.Add(item, "Error: Could not fetch information for remote local at " + localRemote.Url);
		        goInto(item, indentLevel + 1);
		        continue;	                    	
            }

            localGit := &GitExec{Path: localRemote.Url}
            localItemCommit, err := localGit.GetLatestCommit();
            if err != nil {
		        cmd.IndentAndPrintTree(indentLevel, "Error: Could not get remote commit information for " + localRemote.Url);
	            warnings.Add(item, "Error: Could not get remote commit information for " + localRemote.Url);
		        goInto(item, indentLevel + 1);
		        continue;	                    	
            }
            localBranch, err := localGit.GetCurrentBranch()
            if err != nil {
		        cmd.IndentAndPrintTree(indentLevel, "Error: Could not get remote branch information for " + localRemote.Url);
	            warnings.Add(item, "Error: Could not get remote branch information for " + localRemote.Url);
		        goInto(item, indentLevel + 1);
		        continue;	        
            }
            remotePath := "local/" + localBranch

            if localItemCommit != itemCommit {
                cmd.IndentAndPrintTree(indentLevel, "  WARNING: Commit is different. " + remotePath + " " + localItemCommit)                
                commitMismatchWarning = "Warning: The commit in bpm_modules does not match the commit in the remote.\n";
                output, err := git.Run("git show -s --pretty=\"%an -  %ad - %B\" " + itemCommit)
                if err != nil {
                    output = "Error: Could not get commit information"
                }
                output = strings.TrimSpace(output);
                commitMismatchWarning = commitMismatchWarning + "           bpm_module [" + itemCommit + "] - " + output + "\n";
                printedModuleCommit = true;
                output, err = localGit.Run("git show -s --pretty=\"%an -  %ad - %B\" " + localItemCommit)
                if err != nil {
                    output = "Error: Could not get commit information"
                }
                output = strings.TrimSpace(output);
                commitMismatchWarning = commitMismatchWarning + "           " + remotePath + " [" + localItemCommit + "] - " + output + "\n";
            }
        }
        err = git.Fetch(Options.Remote);
        if err != nil {
	        cmd.IndentAndPrintTree(indentLevel, "Error: Could not fetch information for remote " + Options.Remote);
            warnings.Add(item, "Error: Could not fetch information for remote " + Options.Remote);
	        goInto(item, indentLevel + 1);
	        continue;	                    	            	
        }

		remoteBranch := branch;
		// If local item is on HEAD, then we will check the master branch on the remote for it's commit information.		
        if remoteBranch == "HEAD" {
            remoteBranch = "master"
        }

		remotePath := Options.Remote + "/" + remoteBranch;
        remoteItemCommit, err := git.GetLatestCommitRemote(remotePath);
        if err != nil {
	        cmd.IndentAndPrintTree(indentLevel, "Error: Could not get remote commit information for " + remotePath);
            warnings.Add(item, "Error: Could not get remote commit information for " + remotePath);
	        goInto(item, indentLevel + 1);
	        continue;	                    	            	
        }
        if remoteItemCommit != itemCommit {
            cmd.IndentAndPrintTree(indentLevel, "  WARNING: Commit is different. " + remotePath + " " + remoteItemCommit)
            if commitMismatchWarning == "" {			
				commitMismatchWarning = "Warning: The commit in bpm_modules does not match the commit in the remote.\n";
            }
            if !printedModuleCommit {			
	            output, err := git.Run("git show -s --pretty=\"%an -  %ad - %B\" " + itemCommit);
	            if err != nil {
	                output = "Error: Could not get commit information";
	            }
	            output = strings.TrimSpace(output);
				commitMismatchWarning = commitMismatchWarning + "           bpm_module [" + itemCommit + "] - " + output + "\n";
            }
            output, err := git.Run("git show -s --pretty=\"%an -  %ad - %B\" " + remotePath);
            if err != nil {
                output = "Error: Could not get commit information";
            }
            output = strings.TrimSpace(output);
            commitMismatchWarning = commitMismatchWarning + "           " + remotePath + " [" + remoteItemCommit + "] - " + output;
        }
        if commitMismatchWarning != "" {
	        warnings.Add(item, commitMismatchWarning);
        }    
        goInto(item, indentLevel + 1);
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

    if len(warnings.Items) > 0 {
        fmt.Println("***************")
    }
    
    sorted := warnings.GetSortedKeys();
    for _, itemName := range sorted {
        warning := warnings.Items[itemName]
    	fmt.Println("[" + warning.Path + "]");
    	fmt.Println();
    	for _, message := range warning.Messages {
	        fmt.Println("  " + message);
	        fmt.Println();    	
    	}    	
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
    flags.StringVar(&Options.Remote, "remote", "origin", "Use the specified remote instead of origin")
    return cmd;
}
