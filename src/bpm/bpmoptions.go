package main;

import (
    "strings"
    "path"
    "fmt"
)


type BpmOptions struct {
    Recursive bool
    ConflictResolutionType string
    UseLocal string
    UseRemote string
    BpmCachePath string
    BpmFileName string
    LocalModuleName string
    ExcludeFileList string
    SkipNpmInstall bool
    Finalize bool
    Command SubCommand
}

func (options *BpmOptions) getSubCommand(args []string) (SubCommand){
    if len(args) > 1 {
        command := strings.ToLower(args[1]);
        if command == "init" {
            return &InitCommand{}
        }
        if command == "ls" {
            return &LsCommand{}
        }
        if command == "help" {
            return &HelpCommand{};
        }
        if command == "clean" {
            return &CleanCommand{}
        }
        if command == "update" {
            return &UpdateCommand{}
        }
        if command == "install" {
            return &InstallCommand{}
        }
        if command == "version" {
            return &VersionCommand{}
        }
        if command == "uninstall" {
            return &UninstallCommand{}
        }
        fmt.Println("Unrecognized command", command)
    }
    return &HelpCommand{};
}


func (options *BpmOptions) ParseOptions(args []string) {
    options.Command = options.getSubCommand(args)
    options.SkipNpmInstall = options.GetSkipNpmInstallOption(args);
    options.Recursive = options.GetRecursiveOption(args);
    options.ConflictResolutionType = options.GetConflictResolutionTypeOption(args);
    options.UseRemote = options.GetRemoteOption(args);
    options.UseLocal = options.GetRootOption(args);
    options.Finalize = options.GetFinalizeOption(args)
}

func (options *BpmOptions) GetFinalizeOption(args []string) bool {
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--finalize") == 0 });
    if index == -1 {
        return false;
    }
    return true;
}

func (options *BpmOptions) GetSkipNpmInstallOption(args []string) bool {
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--skipnpm") == 0 });
    if index == -1 {
        return false;
    }
    return true;
}


func (options *BpmOptions) GetRecursiveOption(args []string) bool {
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--recursive") == 0 });
    if index == -1 {
        return false;
    }
    return true;
}

func (options *BpmOptions) GetConflictResolutionTypeOption(args []string) string{
    flagValue := "versioning"
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--resolution=") == 0 });
    if index == -1 {
        return flagValue;
    }
    temp := strings.Split(args[index], "=")
    if len(temp) == 2 {
        flagValue = temp[1];
    }
    return flagValue;
}

func (options *BpmOptions) GetRemoteOption(args []string) string{
    remote := "origin"
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--remote=") == 0 });
    if index == -1 {
        return remote;
    }
    temp := strings.Split(args[index], "=")
    if len(temp) == 2 {
        remote = temp[1];
    }
    return remote;
}

func (options *BpmOptions) GetRootOption(args []string) string{
    root := ""
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--root=") == 0 });
    if index == -1 {
        return root;
    }
    temp := strings.Split(args[index], "=")
    if len(temp) == 2 {
        root = temp[1];
    }
    if strings.Index(root, ".") == 0 || strings.Index(root, "..") == 0 {
        root = path.Join(workingPath, root)
    }
    return root
}
