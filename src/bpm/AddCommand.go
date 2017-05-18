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
    "errors"
)

type AddCommand struct {
    Args []string
    Component string
    Name string
}

func (cmd *AddCommand) Initialize() (error) {
    if len(cmd.Args) == 0 {
        return errors.New("Error: A url is required. Ex: bpm add ../myrepo.git");
    }
    cmd.Component = cmd.Args[0];
    return nil;
}

func (cmd *AddCommand) Execute() (error) {
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

        return errors.New("Error: This component is already added")
    } else if name != "" {
        newItem := &BpmDependency{Name:name, Path: path.Join(Options.BpmCachePath, name), Url: cmd.Component}
        newItem.Add();
    }
    return nil;
}

func NewAddCommand() *cobra.Command {
    myCmd := &AddCommand{}
    cmd := &cobra.Command{
        Use:   "add [URL]",
        Short: "add the specified item as a dependency",
        Long:  "add the specified item as a dependency",
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
    flags.StringVar(&Options.RemoteUrl, "remoteurl", "", "Use the specified remote base url instead of the base url from origin. Ex: bpm add ../myrepository.git --remoteurl https://neudesic.timu.com/projects/timu/code/xcom")        
    flags.StringVar(&Options.Local, "root", "", "A relative local path where the dependent repos can be found. Ex: bpm add ../myrepository.git --root=..")
    flags.StringVar(&myCmd.Name, "name", "", "When installing a new component, use the specified name for the component instead of the name from the .git url. Ex: bpm install https://www.github.com/sample/js-sample.git --name sample")
    flags.StringVar(&Options.Remote, "remote", "origin", "Use the specified remote as the base url when not using --root")
    flags.StringVar(&Options.Branch, "branch", "master", "Use the specified branch instead master when not using --root");
    return cmd;
}