Another package manager for nodejs projects

How to build the bpm

./build.sh

This script will create a symbolic link for bpm in the /usr/local/bin folder.


Overview of the bpm.json file structure.

    {
        "name" : "my-component",
        "version" : "1.0.0",
        "dependencies" : {
            "my-depencency-1" : {
                "url" : "../my-depencency-1.git",
                "commit" : "90b0a2da501451cf55ee07f9faeb3f8707af6011"
            }
        }
    }

name: The name of the component

version: The version number of the component. This is used to resolve conflicts that may arise when computing the dependency tree.

dependencies: A map of the dependencies where the map key is the name of the dependency

dependency map item

url: The full or relative URL to the repository

commit: The commit hash

Build the bpm dependencies specified in bpm.json. bpm depends on git and therefore must be installed. The repository where the bpm command is run must be a git repository with at least a remote of origin.

    bpm [--remote=origin] | [--root=.]

Example:

    bpm
    bpm --root=../js
    bpm --remote=brandon

For example, given the following bpm.json for the repository my-component

    {
        "name" : "my-component",
        "version" : "1.0.0",
        "dependencies" : {
            "my-depencency-1" : {
                "url" : "../my-depencency-1.git",
                "commit" : "90b0a2da501451cf55ee07f9faeb3f8707af6011"
            }
        }
    }

The bpm command will:

- create a directory in the bpm_modules folder for my-component
- git fetch the repository at the specified URL
- git checkout the specified commit hash in a subfolder
- copy the subfolder to the node_modules folder as my-component
- run npm install on my-component

In this example the URL is a relative URL. Dependency URLs can be a full URL or a relative URL. For any dependency that has a relative url, the remote option will be used to resolve the relative url to a full url. origin is the default remote. Therefore, if the origin is http://github.com/user/my-component.git, then the dependency url will be resolved to http://github.com/user/my-depencency-1.git

The dependency url and commit are required fields unless the --root option is used.

When the --root option is used, instead of downloading the code from the dependency url, bpm will attempt to locate the dependency on the local disk relative to the specified root.
Given the command
    bpm --root=../mydependencies

bpm will expect the following folder structure on disk and ignore the specified url and commit hash for each dependency.

    mycomponent\
    mydependencies\
        my-depencency-1\


The version number is used to resolve conflicts when the dependency tree contains the same dependency but with different commit hashes.

Install a new dependency and save to the bpm.json. The repository being installed must contain a bpm.json file
The url parameter can be a relative URL or a full URL. The relative URL will be relative to the specified remote or origin by default

    bpm install [url] [commit] [--remote=origin]

Example:

    bpm install ../mortar.git
    bpm install ../mortar.git ad2c7c47362fc682079307cbb1db7ef944997364
    bpm install ../mortar.git --remote=brandon
    bpm install https://neudesic.timu.com/projects/timu/code/master/mortar.git

The version number in the bpm.json will be incremented automatically when the install command is used. If the commit is not specified then the last commit hash will be used.

Update the commit of existing dependency to the latest

    bpm update [dependencyName] [--remote=origin]

Example:

    bpm update mortar
    bpm update mortar --remote=brandon


The version number in the bpm.json will be incremented automatically when the update command is used.


Create a default bpm.json

    bpm --new=module

Example:

    bpm --new=my-component

Clean the bpm_modules

    bpm --clean

Sample bpm.json files for repositories login-client, mortar and null-query

my-component bpm.json

    {
        "name" : "my-component",
        "version" : "1.0.0",
        "dependencies" : {
            "my-depencency-1" : {
                "url" : "../my-depencency-1.git",
                "commit" : "90b0a2da501451cf55ee07f9faeb3f8707af6011"
            },
            "my-depencency-2":{
                "url" :"../my-dependency-2.git",
                "commit" :"d5f3dfd6625ba0c92709198c319ec2276471610e"
            }
        }
    }

my-depencency-1 bpm.json

    {
        "name" : "my-depencency-1",
        "version" : "1.0.1",
        "dependencies" : {
            "my-depencency-2" : {
                "url" : "../my-depencency-2.git",
                "commit" : "d5f3dfd6625ba0c92709198c319ec2276471610e"
            }
        }
    }

my-depencency-2 bpm.json

    {
        "name" : "my-depencency-2",
        "version" : "1.0.0",
        "dependencies": {}
    }


Expected layout on disk for the above example since relative URLs are used.

    dev\
        my-component\
        my-depencency-1\
        my-depencency-2\


// disk
When bpm is run a bpm_modules folder will be created which contains copies of the repositories which are then copied into the node_modules folder

    my-component\
        package.json
        bpm.json
        bpm_modules\
            my-depencency-1\
                    90b0a2da501451cf55ee07f9faeb3f8707af6011\
            my-depencency-2\
                    d5f3dfd6625ba0c92709198c319ec2276471610e\
        node_modules\
            my-depencency-1\
            my-depencency-2\
        src\
            main.js


When the bpm command is run with the --root option, then the bpm_modules dependencies folders will contain a local folders

    bpm_modules\
        my-depencency-1\
            local\


bpm module cache

When the bpm command is run any existing items in the bpm_module cache will be used and the dependency will not be redownloaded. Additionally, if the local folder exists for the component in the module cache, then this cache item will always be used even if the cache does not contain an item for the dependency hash

bmp update module-name [--root=xxx] // get latest of module, put in node_modules, update json with commit, increment version number of json file, if root is specified look in file path root/module-name, update each sub dep in node_modules with --root logic


Version resolution

It is possible that the dependency tree will contain multiple reference to the same dependency. It is also possible that commit hash for those dependencies will be different. In this case, the version number of the dependency in the dependency's bpm.json file will be compared and the latest version will be used.

Example:

- my-component has dependency-2 with commit X, and the dependency-2's bpm.json file for commit X is version 1.0.0
- my-component has dependency-1 and dependency-1 also contains dependency-2 but with commit Y, and dependency-2's bpm.json file for commit Y is version 1.0.1

In this case, version 1.0.1 of dependency-2 will be used.
