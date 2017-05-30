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
	PackageManager: "npm",
    BpmCachePath: "bpm_modules",
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

func PathExists(location string) (bool) {
    _, err := os.Stat(location)
    if err != nil && os.IsNotExist(err){
	    return false;
    }
    return true;
}

func UseLocal(url string) bool {
    if Options.Local != "" && !strings.HasPrefix(url, "http") {
        return true;
    }
    return false;
}

func UpdatePackageJsonVersion(location string) (error) {
    pj := PackageJson{Path: location}
    err := pj.Load();
    if err == nil {
        err = pj.UpdateVersion();
    }
    if err == nil {
        err = pj.Save();
    }
    return err;
}

func MakeRemoteUrl(itemUrl string) (string, error) {
    if UseLocal(itemUrl) {
        _, name := filepath.Split(itemUrl)
        return path.Join(Options.Local, strings.Split(name, ".git")[0]), nil;
    }
    adjustedUrl := itemUrl;
    if !strings.HasPrefix(adjustedUrl, "http") {
        var remoteUrl string;

        relative := "";
        if Options.RemoteUrl != "" {
            remoteUrl = Options.RemoteUrl;
        } else {
            git := GitExec{LogOutput: true}
            remote := git.GetRemote(Options.Remote)
            if remote == nil {
                return "", errors.New("Error: There was a problem getting the remote url " + Options.Remote)
            }
            remoteUrl = remote.Url;
            if !strings.HasPrefix(adjustedUrl, "./") && !strings.HasPrefix(adjustedUrl, "../"){
	            relative = "..";
            }
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
        fmt.Println("Using remote Url", adjustedUrl)
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
