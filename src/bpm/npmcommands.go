package main;

import (
    "os/exec"
    "strings"
    "bytes"
    "fmt"
)

type NpmCommands struct {
    Path string
}

func (npm *NpmCommands) RunCommand(dir string, command string) error {
    cmd := exec.Command("npm")
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

func (npm *NpmCommands) Uninstall(item string) error {
    fmt.Println("Running npm uninstall on", item);
    npmCommand := "npm uninstall " + item
    err := npm.RunCommand(npm.Path, npmCommand)
    if err != nil {
        return err;
    }
    return nil;
}

func (npm *NpmCommands) InstallUrl(url string) error {
    fmt.Println("Running npm install in", npm.Path, "on", url)
    // git clone <repo url> <destination directory>
    // npm cache clean
    npmCommand := "npm install " + url
    err := npm.RunCommand(npm.Path, npmCommand);
    if err != nil {
        return err;
    }
    return nil;
}


func (npm *NpmCommands) Install() error {
    return npm.InstallUrl("");
}
