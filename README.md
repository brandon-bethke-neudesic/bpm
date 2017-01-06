Another package manager for nodejs projects

How to build the bpm

./build.sh

This script will create a symbolic link for bpm in the /usr/local/bin folder.
It also creates the linux and darwin version of the app.

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

name: The name of the component. Generally this is the same name as the git repo.

version: The version number of the component. Versioning is used to resolve conflicts that may arise when computing the dependency tree.

dependencies: A map of the dependencies. The map key is the name of the dependency

dependency map item

url: The full or relative URL to the repository

commit: The commit hash

Install dependencies
Dependencies specified in the bpm.json are installed using the `bpm install` command. bpm requires git and the repository where the bpm command is run must be a git repository with at least a remote of origin.

    bpm install [--remote=myremote] | [--root=mypath]

Example:

    bpm install
    bpm install --root=../js
    bpm install --remote=brandon

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
- run a supported package manager install on my-component/hash. (npm or yarn)

In this example, the URL is relative. Dependency URLs can be a full URL or a relative URL. For any dependency that has a relative url, the `--remote` option will be used to resolve the relative url to a full url. origin is the default remote. Therefore, if the origin is http://github.com/user/my-component.git, then the dependency url will be resolved to http://github.com/user/my-depencency-1.git

The dependency url and commit are required to correctly install a dependency.

When the --root option is used, instead of downloading the code from the dependency url, bpm will attempt to locate the dependency on the local disk relative to the specified root.

Given the command
    bpm install --root=../mydependencies

bpm will expect the following folder structure on disk and ignore the specified url and commit hash for each dependency.

    mycomponent\
    mydependencies\
        my-dependency-1\

Install a single existing dependency.

    bpm install [name]

Example:

    bpm install my-dependency-1

Install a new dependency.
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

    bpm update [dependencyName] [--remote=myremote | --root=mypath] [--recursive]

Example:

    bpm update
    bpm update --root=../js
    bpm update mortar
    bpm update mortar --remote=brandon
    bpm update mortar --root=../mydependencies
    bpm update mortar --root=../mydependencies --recursive

The version number in the bpm.json will be incremented automatically when a dependency has changed.

When using the `--root` option, uncommited local changes are always copied to bpm_modules. If the repository contains outstanding changes, then the commit hash of the dependency will be updated to 'local', which is obviously invalid. This is to let the developer know that they need to finalize the dependency commit. To finalize, it is expected that local changes will be committed and then by running the update command again, and if there are no detected changes, then the commit hash will be updated to the latest.

The --recursive option only works when specified with the --root option. bpm will recursively go through all dependencies and update the commit hashes based on the last local commit hash for the dependency. The version number of sub-dependencies are also incremented.


Uninstall a dependency.
Removed the dependency from bpm_modules and performs an npm uninstall [depName]

    bpm uninstall [dependencyName]

Example:

    bpm uninstall mortar


Create a default bpm.json

    bpm init <modulename>

Example:

    bpm init my-component

Clean the bpm_modules

    bpm clean

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

Dependency conflict resolution.

It is possible that the dependency tree will contain multiple reference to the same dependency. It is also possible that the commit hash for those dependencies will be different. In this case, the version number of the dependency in the dependency's bpm.json file will be compared and the latest version will be used.

Example:

- my-component has dependency-2 with commit X, and the dependency-2's bpm.json file for commit X is version 1.0.0
- my-component has dependency-1 and dependency-1 also contains dependency-2 but with commit Y, and dependency-2's bpm.json file for commit Y is version 1.0.1

In this case, version 1.0.1 of dependency-2 will be used.

Using the --resolution option, it is possible to specify an alternative conflict resolution strategy.

The option `--resolution=revisionlist` will attempt to determine which commit is the latest commit using the git revision history.

The option `--skipnpm` will skip the package manager install phase.

The option `--remoteurl=https://host/path.git` will cause bpm to use the specified url as the remote url for all relative path rather than the remote name.

Supported Package Managers

bpm supports npm and yarn. To specify a package manger use the --pkgm= option. By default npm is used.

    bpm --pkgm=yarn [--yarn-packages-root=] [--yarn-modules-folder=]
    bpm --pkgm=npm
