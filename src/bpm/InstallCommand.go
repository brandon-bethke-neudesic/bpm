package main;

import (
    "os"
    "fmt"
    "strings"
    "bpmerror"
    "github.com/spf13/cobra"
    "path/filepath"
    "net/url"
    "path"
)

type InstallCommand struct {
    Args []string
    Component string
    Name string
}

func (cmd *InstallCommand) Initialize() (error) {
    if len(cmd.Args) > 0 {
        cmd.Component = cmd.Args[0];
    }

    if strings.HasPrefix(cmd.Component, "http") && Options.UseLocalPath != "" {
        return bpmerror.New(nil, "Error: The root option should not be used with a full url path when installing a component. Ex: bpm install http://www.github.com/neudesic/myproject.git")
    }

    return nil;
}

func (cmd *InstallCommand) Execute() (error) {
    Options.EnsureBpmCacheFolder();

    bpm := BpmModules{}
    err := bpm.Load(Options.BpmCachePath);
    if err != nil {
        return err;
    }
    name := cmd.Component;
    if strings.HasPrefix(cmd.Component, "http") {
        parsedUrl, err := url.Parse(cmd.Component);
        if err != nil {
            return bpmerror.New(err, "Error: Could not parse the component url");
        }
        _, name = filepath.Split(parsedUrl.Path)
        name = strings.Split(name, ".git")[0]
    } else if strings.HasPrefix(cmd.Component, "../")  || strings.HasPrefix(cmd.Component, "./") {
        _, name = filepath.Split(cmd.Component);
        name = strings.Split(name, ".git")[0];
    }

    if cmd.Name != "" {
        name = cmd.Name;
    }

    if name != "" && bpm.HasDependency(name){
        bpm.Dependencies[name].Install();
    } else if name != "" {
        newItem := &BpmDependency{Name:name, Path: path.Join(Options.BpmCachePath, name), Url: cmd.Component}
        newItem.Install();
    } else {
        if !bpm.HasDependencies() {
            fmt.Println("There are no dependencies")
            return nil;
        }

        fmt.Println("Installing all dependencies for " + Options.RootComponent);
        for _, item := range bpm.Dependencies {
            err = item.Install();
            if err != nil {
                return err;
            }
        }
    }

    if !Options.SkipNpmInstall {
        err = moduleCache.Install()
        if err != nil {
            return err;
        }
    }
    return nil;
}


func NewInstallCommand() *cobra.Command {
    myCmd := &InstallCommand{}
    cmd := &cobra.Command{
        Use:   "install [COMPONENT]",
        Short: "install the specified item or all the dependencies",
        Long:  "install the specified item or all the dependencies",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            myCmd.Args = args;
            return myCmd.Initialize();
        },
        Run: func(cmd *cobra.Command, args []string) {
            Options.Command = "install"
            err := myCmd.Execute();
            if err != nil {
                fmt.Println(err);
                fmt.Println("Finished with errors");
                os.Exit(1)
            }
        },
    }

    flags := cmd.Flags();
    flags.StringVar(&Options.UseLocalPath, "root", "", "A relative local path where the dependent repos can be found. Ex: bpm install --root=..")
    flags.StringVar(&myCmd.Name, "name", "", "When installing a new component, use the specified name for the component instead of the name from the .git url. Ex: bpm install https://www.github.com/sample/js-sample.git --sample")
    flags.BoolVar(&Options.SkipNpmInstall, "skipnpm", false, "Do not perform a npm install")
    flags.StringVar(&Options.UseRemoteName, "remote", "origin", "")
    return cmd;
}