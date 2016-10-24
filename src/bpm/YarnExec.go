package main;

import (
    "fmt"
)

type YarnExec struct {
    Path string
    ModulesFolder string
    PackagesRoot string
}

func (yarn *YarnExec) ParseOptions(args []string) {
    yarn.ModulesFolder = Options.GetNameValueOption(args, "--yarn-modules-folder=", "../../../node_modules");
    yarn.PackagesRoot = Options.GetNameValueOption(args, "--yarn-packages-root=", "");
}

func (yarn *YarnExec) Install() error {
    fmt.Println("Running yarn install in", yarn.Path)
    yarnCommand := "yarn install"
    if yarn.ModulesFolder != "" {
        yarnCommand = yarnCommand + " --modules-folder " + yarn.ModulesFolder
    }
    if yarn.PackagesRoot != "" {
        yarnCommand =  yarnCommand + " --packages-root " + yarn.PackagesRoot
    }
    rc := OsExec{Dir: yarn.Path, LogOutput: true}
    _, err := rc.Run(yarnCommand)
    if err != nil {
        return err;
    }
    return nil;
}
