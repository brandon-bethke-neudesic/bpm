package main;


import (
    "os"
    "fmt"
    "path"
    "net/url"
    "strings"
    "bpmerror"
    "path/filepath"
    "errors"
)

var moduleCache = ModuleCache{Items:make(map[string]*ModuleCacheItem)};
var Options = BpmOptions {
    BpmCachePath: "bpm_modules",
    BpmFileName: "bpm.json",
    LocalModuleName: "local",
    ExcludeFileList: ".git|.gitignore|.gitmodules|bpm_modules|node_modules",
}

func SliceIndex(limit int, predicate func(i int) bool) int {
    for i := 0; i < limit; i++ {
        if predicate(i) {
            return i
        }
    }
    return -1
}

func PathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    fmt.Println("Error:", err)
    return true
}


func UseLocal(url string) bool {
    if Options.UseLocalPath != "" && !strings.HasPrefix(url, "http") {
        return true;
    }
    return false;
}


func EnsurePath(path string){
    if _, err := os.Stat(path); os.IsNotExist(err) {
        os.MkdirAll(path, 0777)
    }
}

func MakeRemoteUrl(itemUrl string) (string, error) {
    if UseLocal(itemUrl) {
        _, name := filepath.Split(itemUrl)
        return path.Join(Options.UseLocalPath, strings.Split(name, ".git")[0]), nil;
    }
    adjustedUrl := itemUrl;
    if !strings.HasPrefix(adjustedUrl, "http") {
        var remoteUrl string;

        relative := "";
        if Options.UseRemoteUrl != "" {
            remoteUrl = Options.UseRemoteUrl;
        } else {
            git := GitExec{}
            remote := git.GetRemote(Options.UseRemoteName)
            if remote == nil {
                return "", errors.New("Error: There was a problem getting the remote url " + Options.UseRemoteName)
            }
            remoteUrl = remote.Url;
            relative = "..";
        }
        parsedUrl, err := url.Parse(remoteUrl)
        if err != nil {
            return "", bpmerror.New(err, "Error: There was a problem parsing the remote url " + remoteUrl)
        }

        urlPath := strings.Split(parsedUrl.Path, ".git")[0]
        adjustedUrl = parsedUrl.Scheme + "://" + path.Join(parsedUrl.Host, urlPath, relative, itemUrl)
        if !strings.HasSuffix(adjustedUrl, ".git") {
            adjustedUrl = adjustedUrl + ".git";
        }
        fmt.Println("Adjusted Url", adjustedUrl)
    }
    return adjustedUrl, nil;
}

func main() {
    Options.WorkingDir, _ = os.Getwd();
    _, filename := filepath.Split(Options.WorkingDir)
    Options.RootComponent = filename;

    cmd := NewBpmCommand();
    if err := cmd.Execute(); err != nil {
        fmt.Println(err);
        fmt.Println("Finished with errors");
        os.Exit(1)
    }
    os.Exit(0);
}
