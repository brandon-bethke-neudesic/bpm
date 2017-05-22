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

func (cmd *UpdateCommand) Description() (string) {
    return `
    
Update a dependency in bpm_modules then run npm install. Updating means to fetch and merge changes from origin/master unless otherwise specified.
Only direct dependencies are updated. Dependencies of dependencies are not updated.

When updating, all changes in tracked files in the dependency will be discarded.

Option --root
	
	The root option tells bpm to add a local remote to the dependency for the specified relative location and then fetch the changes from the local remote.
	The local remote url will be the current working directory plus the specified location plus the name of the dependency. 
	Ex: bpm update bpmdep1 --root=..
	The local remote url in this case would be CWD/../bpmdep1
		
	Any committed changes will be fetched and merged.
	Any staged or untracked changes will be copied.
	If the local remote is on a branch, then the dependency will also be switched to use this branch.
		
Option --remote

	The remote option instructs bpm to use specified remote when fetching changes rather than the default origin.
	This option is not to be used with --root since they are mutually exclusive.
	
Option --branch

	The branch option instructs bpm to use the specified branch when merging changes rather than default master.
	This option is not to be used with --root since the branch the local remote is on always has precedence.
		
Versioning

	The version number in package.json will always be updated and will be modified with a build number. Ex: 0.0.1-1495480065
	Where 1495480065 represents the number of seconds since Jan 1st 1970 UTC.

	When using the --root option and when copying changes, if any, from a local dependency repository, then the package.json version number in that dependenecy will also be updated.

Examples:

	# Update all direct dependencies and using the local path as the source repo
	bpm update --root=../js	
	bpm update --root=../js --skipnpm
	
	# Update just the groupscale depdendency
	bpm update groupscale --root=../js --skipnpm

	# Update the groupscale dependency and use the base url specified by the remote brandon as the source repo
	bpm update groupscale --remote brandon
	
	# Update the groupscale dependency and use the base url specified
	bpm update groupscale --remoteurl https://neudesic.timu.com/projects/timu/code/brandon-bethke --skipnpm
`
    	
}

func (cmd *UpdateCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Name = cmd.Args[0];
    }
    
    if Options.Local != "" && Options.RemoteUrl != "" {
    	return errors.New("Error: The --root option and --remoteurl option should not be used together.");
    }
    
    return nil;
}

func (cmd *UpdateCommand) Execute() (error) {
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
    	if bpm.HasDependency(name){
	        err = bpm.Dependencies[name].Update();
	        if err != nil {
	            return err;
	        }
    	} else {
	        return errors.New("Error: The dependency " + name + " has not been installed or is not a direct dependency.")    		
    	}
    } else {
        fmt.Println("Scanning " + Options.RootComponent);
        for _, item := range bpm.Dependencies {
        	fmt.Println("Found " + item.Name);
            err = item.Update();
            if err != nil {
                return err;
            }
        }
    }

    if !Options.SkipNpm {
        err = moduleCache.Install()
        if err != nil {
            return bpmerror.New(err, "Error: There was an issue performing npm install on the dependencies")
        }
    }

    return nil;
}

func NewUpdateCommand() *cobra.Command {
    myCmd := &UpdateCommand{}
    cmd := &cobra.Command{
        Use:   "update [NAME]",
        Short: "update the specified dependency",
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

    flags := cmd.Flags();
    flags.StringVar(&Options.RemoteUrl, "remoteurl", "", "Use the specified remote base url instead of the base url from origin. Ex: bpm update --remoteurl https://neudesic.timu.com/projects/timu/code/xcom")        
    flags.StringVar(&Options.Local, "root", "", "A relative local path where the dependent repos can be found. Ex: bpm install --root=..")
    flags.BoolVar(&Options.Deep, "deep", false, "Update all dependencies in the bpm modules hierachy");
    flags.BoolVar(&Options.SkipNpm, "skipnpm", false, "Do not perform a npm install")
    flags.BoolVar(&Options.IgnoreMissingLocal, "iml", false, "Ignore Missing Local. When the url for remote local does not exist, use the specified remote (default \"origin\")")
    flags.StringVar(&Options.Remote, "remote", "origin", "Use the specified remote instead of origin when not using --root")
    flags.StringVar(&Options.Branch, "branch", "master", "Use the specified branch instead master when not using --root");

    return cmd
}
