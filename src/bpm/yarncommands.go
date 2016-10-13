package main;

import (
    "os/exec"
    "strings"
    "bytes"
    "fmt"
)

type YarnCommands struct {
    Path string
    ModulesFolder string
    PackagesRoot string
}

func (yarn *YarnCommands) RunCommand(dir string, command string) error {
    cmd := exec.Command("yarn")
    if dir != "" {
        cmd.Dir = dir;
    }
    cmd.Args = strings.Split(command, " ");
    fmt.Println(cmd.Args)
	var out bytes.Buffer
    var errOut bytes.Buffer
	cmd.Stdout = &out
    cmd.Stderr = &errOut
	err := cmd.Run()
    fmt.Println(out.String())
    fmt.Println(errOut.String())
	if err != nil {
        fmt.Println(err)
        return err;
	}
    return nil
}

func (yarn *YarnCommands) ParseOptions(args []string) {
    yarn.ModulesFolder = Options.GetNameValueOption(args, "--yarn-modules-folder=", "../../../node_modules");
    yarn.PackagesRoot = Options.GetNameValueOption(args, "--yarn-packages-root=", "");
}

func (yarn *YarnCommands) Install() error {
    fmt.Println("Running yarn install in", yarn.Path)
    yarnCommand := "yarn install"
    if yarn.ModulesFolder != "" {
        yarnCommand = yarnCommand + " --modules-folder " + yarn.ModulesFolder
    }
    if yarn.PackagesRoot != "" {
        yarnCommand =  yarnCommand + " --packages-root " + yarn.PackagesRoot
    }
    err := yarn.RunCommand(yarn.Path, yarnCommand);
    if err != nil {
        return err;
    }
    return nil;
}
