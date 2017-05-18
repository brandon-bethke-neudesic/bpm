package main;

import (
    "strings"
    "path"
    "os"
)


type BpmOptions struct {
    Deep bool
    RootComponent string
    Local string
    Remote string
    RemoteUrl string
    BpmCachePath string
    ExcludeFileList string
    SkipNpm bool
    PackageManager string
    WorkingDir string
    Branch string
}

func (options *BpmOptions) EnsureBpmCacheFolder() {
    if _, err := os.Stat(Options.BpmCachePath); os.IsNotExist(err) {
        os.Mkdir(Options.BpmCachePath, 0777)
    }
}

func (options *BpmOptions) Validate() error {
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
