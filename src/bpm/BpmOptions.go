package main;

import (
    "strings"
    "path"
    "fmt"
    "os"
    "errors"
    "bpmerror"
)


type BpmOptions struct {
    Recursive bool
    ConflictResolutionType string
    UseLocalPath string
    UseRemoteName string
    UseRemoteUrl string
    BpmCachePath string
    BpmFileName string
    LocalModuleName string
    ExcludeFileList string
    SkipNpmInstall bool
    Finalize bool
    PackageManager string
    WorkingDir string
    Trim bool
    UseParentUrl bool
    Command SubCommand
}

func (options *BpmOptions) EnsureBpmCacheFolder() {
    if _, err := os.Stat(Options.BpmCachePath); os.IsNotExist(err) {
        os.Mkdir(Options.BpmCachePath, 0777)
    }
}

func (options *BpmOptions) DoesBpmFileExist() (error) {
    if _, err := os.Stat(Options.BpmFileName); os.IsNotExist(err) {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    return nil
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


func (options *BpmOptions) Parse(args []string) {
    options.Command = options.getSubCommand(args)
    options.SkipNpmInstall = options.GetBoolOption(args, "--skipnpm")
    options.Recursive = options.GetBoolOption(args, "--recursive")
    options.ConflictResolutionType = options.GetNameValueOption(args, "--resolution=", "versioning")
    options.UseRemoteName = options.GetNameValueOption(args, "--remote=", "origin")
    options.UseRemoteUrl = options.GetNameValueOption(args, "--remoteurl=", "")
    options.UseLocalPath = options.GetRootOption(args);
    options.Finalize = options.GetBoolOption(args, "--finalize")
    options.PackageManager = options.GetNameValueOption(args, "--pkgm=", "npm")
    options.WorkingDir, _ = os.Getwd();
    options.Trim = options.GetBoolOption(args, "--trim")
    options.UseParentUrl = options.GetBoolOption(args, "--useparenturl")
}

func (options *BpmOptions) Validate() error {
    if options.Recursive && options.UseLocalPath == "" {
        return bpmerror.New(nil, "Error: The --recursive option can only be used with the --root= option")
    }
    if options.Trim && options.Command.Name() != "clean" {
        return bpmerror.New(nil, "Error: The --trim option can only be used with the clean command")
    }
    return nil
}

func (options *BpmOptions) GetNameValueOption(args []string, name string, defaultValue string) (string) {
    flagValue := defaultValue
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], name) == 0 });
    if index == -1 {
        return flagValue;
    }
    temp := strings.Split(args[index], "=")
    if len(temp) == 2 {
        flagValue = temp[1];
    }
    return flagValue;
}

func (options *BpmOptions) GetBoolOption(args []string, name string) (bool) {
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], name) == 0 });
    if index == -1 {
        return false;
    }
    return true;
}

func (options *BpmOptions) GetRootOption(args []string) string{
    root := options.GetNameValueOption(args, "--root=", "")
    if strings.Index(root, ".") == 0 || strings.Index(root, "..") == 0 {
        root = path.Join(options.WorkingDir, root)
    }
    return root
}
