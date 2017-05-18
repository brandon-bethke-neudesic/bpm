#### Overview

bpm is a tool to help to keep submodules in the `bpm_modules` folder updated with local and remote changes. It will also npm install these dependencies.


#### Build bpm

	./build.sh

This script will create a symbolic link for bpm in the /usr/local/bin folder.
The script will also create a linux binary.
If the `TPM_CLUSTER_HOME` variable is set then the linux version will be copied to `TPM_CLUSTER_HOME/includes/bpm/bin/linux` 


#### Install

	bpm install [NAME]

npm install the direct dependency with the specified name. If no name is specified, then all dependencies will be installed.


#### Update

	bpm update [NAME] [OPTIONS]
	 
Update the direct dependency with the specified name. If no name is specified, then all direct dependencies will be updated.

#### Listing

	bpm ls 
	
List the dependency heirarchy.

#### Add

	bpm add [URL] [OPTIONS]

Add the specified URL as a submodule in the bpm_modules folder.
If the --root option is specified then the module will also be updated.

#### Remove

	bpm remove [NAME] [OPTIONS]
	
Remove the specified submodule. If no items is specified, then all submodules will be removed.

#### Status

	bpm status NAME
	
Perform a git status on the specified dependency. The name can be a path but exclude the bpm_modules dirtories.

Examples:
	
	bpm status bpmdep2
	bpm status bpmdep3/bpmdep1