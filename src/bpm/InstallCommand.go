package main;

import (
    "os"
    "github.com/spf13/cobra"
    "fmt"
    "errors"
    "bpmerror"
)

type InstallCommand struct {
    Args []string
    Name string
}

func (cmd *InstallCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    return nil;
}

func (cmd *InstallCommand) Description() (string){
	return `
	
Run npm install on each dependency in the hierarchy. The install command will not first perform an update.

If there is a situation where the same dependency is being installed multiple times, the version number in package.json 
will be compared and the higher version will be installed.

Examples:

	bpm install
	bpm install bpmdep1
`
}

func (cmd *InstallCommand) Execute() (error) {	
    bpm := BpmModules{}
    err := bpm.Load(Options.BpmCachePath);
    if err != nil {
        return err;
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies")
        return nil;
    }
    name := cmd.Name;
	// Check if updating a single depedency, otherwise update all of them.
    if name != "" {	
    	if bpm.HasDependency(name) {
	        err = bpm.Dependencies[name].Install();
	        if err != nil {
	            return err;
	        }
    	} else {
	        return errors.New("Error: The dependency " + name + " has not been installed or is not a direct dependency.")    	
    	}
    } else {    	
        fmt.Println("Scanning " + Options.RootComponent);
        for _, item := range bpm.Dependencies {
        	fmt.Println("Found " + item.Path);
            err = item.Install();
            if err != nil {
                return err;
            }
        }
    }

    err = moduleCache.Install()
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue performing npm install on the dependencies")
    }
    return nil;
}

func NewInstallCommand() *cobra.Command {
    myCmd := &InstallCommand{}
    cmd := &cobra.Command{
        Use:   "install [NAME]",
        Short: "install the specified dependency",
        Long:  myCmd.Description(),
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

	//flags := cmd.Flags();    
    //flags.StringVar(&Options.PackageManager, "pkgm", "npm", "Use the specified package manager to install the component. npm or yarn")

    return cmd
}
