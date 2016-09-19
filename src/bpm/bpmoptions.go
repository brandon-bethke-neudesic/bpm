package main;

import (
    "strings"
    "path"
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
}

func GetSkipNpmInstall(args []string) bool {
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--skipnpm") == 0 });
    if index == -1 {
        return false;
    }
    return true;
}


func GetRecursiveFlag(args []string) bool {
    index := SliceIndex(len(args), func(i int) bool { return strings.Index(args[i], "--recursive") == 0 });
    if index == -1 {
        return false;
    }
    return true;
}

func GetConflictResolutionType(args []string) string{
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

func GetRemote(args []string) string{
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

func GetLocal(args []string) string{
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
