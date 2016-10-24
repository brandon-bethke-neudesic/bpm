package main;

import (
    "fmt"
)

type NpmExec struct {
    Path string
}

func (npm *NpmExec) Uninstall(item string) error {
    fmt.Println("Running npm uninstall on", item);
    npmCommand := "npm uninstall " + item
    rc := OsExec{Dir: npm.Path, LogOutput: true}
    _, err := rc.Run(npmCommand)
    if err != nil {
        return err;
    }
    return nil;
}

func (npm *NpmExec) InstallUrl(url string) error {
    fmt.Println("Running npm install in", npm.Path, "on", url)
    // git clone <repo url> <destination directory>
    // npm cache clean
    npmCommand := "npm install " + url
    rc := OsExec{Dir: npm.Path, LogOutput: true}
    _, err := rc.Run(npmCommand)
    if err != nil {
        return err;
    }
    return nil;
}


func (npm *NpmExec) Install() error {
    return npm.InstallUrl("");
}
