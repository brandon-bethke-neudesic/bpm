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

func (yarn *YarnCommands) GetYarnModulesFolder(args []string) string {
    flagValue := "../../../node_modules"
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--yarn-modules-folder=") == 0 });
    if index == -1 {
        return flagValue;
    }
    temp := strings.Split(args[index], "=")
    if len(temp) == 2 {
        flagValue = temp[1];
    }
    return flagValue;
}

func (yarn *YarnCommands) GetYarnPackagesRoot(args []string) string {
    flagValue := ""
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--yarn-packages-root=") == 0 });
    if index == -1 {
        return flagValue;
    }
    temp := strings.Split(args[index], "=")
    if len(temp) == 2 {
        flagValue = temp[1];
    }
    return flagValue;
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
