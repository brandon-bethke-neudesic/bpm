package main;

import (
    "fmt"
    "os"
    "bpmerror"
    "errors"
    "github.com/spf13/cobra"
)


type UpdateCommand struct {
    Args []string
    Name string
}

func (cmd *UpdateCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    return nil;
}

func (cmd *UpdateCommand) Execute() (error) {
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    bpm := BpmData{};
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(nil, "Error: There was a problem loading the bpm.json file")
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    if cmd.Name != "" && !bpm.HasDependency(cmd.Name) {
        return bpmerror.New(err, "Error: Could not find module " + cmd.Name + " in the dependencies")
    }

    fmt.Println("Processing dependencies for", bpm.Name, "version", bpm.Version);
    // Always process the keys sorted by name so the installation is consistent
    sortedKeys := bpm.GetSortedKeys();
    for _, updateModule := range sortedKeys {
        depItem := bpm.Dependencies[updateModule]
        // If a specific module name was specified then skip the others.
        if cmd.Name != "" && cmd.Name != updateModule {
            continue;
        }
        moduleSourceUrl := depItem.Url;
        if moduleSourceUrl == "" {
            moduleSourceUrl = updateModule;
        }
        moduleSourceUrl, err := MakeRemoteUrl(moduleSourceUrl);

        moduleBpm, cacheItem, err := ProcessModule(moduleSourceUrl);
        if err != nil {
            return err;
        }
        moduleCache.Add(cacheItem)
        fmt.Println("Processing dependencies for", cacheItem.Name, "version", moduleBpm.Version);
        err = ProcessDependencies(moduleBpm)
        if err != nil {
            return err;
        }
        newItem := &BpmDependency{Url: depItem.Url}
        bpm.Dependencies[updateModule] = newItem;

    }
    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            return bpmerror.New(err, "Error: There was an issue performing npm install on the dependencies")
        }
    }

    err = bpm.IncrementVersion();
    if err != nil {
        return err;
    }
    err = bpm.WriteFile(Options.BpmFileName)
    if err != nil {
        return err;
    }

    return nil;
}

func NewUpdateCommand() *cobra.Command {
    myCmd := &UpdateCommand{}
    cmd := &cobra.Command{
        Use:   "update [NAME]",
        Short: "update the specified component",
        Long:  "update the specified component",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "update"

            err := myCmd.Execute();
            if err != nil {
                fmt.Println(err);
                fmt.Println("Finished with errors");
                os.Exit(1)
            }
        },
    }

    flags := cmd.Flags();
    flags.StringVar(&Options.UseLocalPath, "root", "", "")
    flags.StringVar(&Options.ConflictResolutionType, "resolution", "revisionlist", "")
    return cmd
}
