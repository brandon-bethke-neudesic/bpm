package main;

import (
)

type NpmExec struct {
    Path string
}

func (npm *NpmExec) Uninstall(item string) error {
    npmCommand := "npm uninstall " + item
    rc := OsExec{Dir: npm.Path, LogOutput: true}
    _, err := rc.Run(npmCommand)
    if err != nil {
        return err;
    }
    return nil;
}

func (npm *NpmExec) InstallUrl(url string) error {
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
