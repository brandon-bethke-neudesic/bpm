package main;

import (
    "strings"
    "fmt"
    "regexp"
    "errors"
)

type GitExec struct {
    Path string
}


func (git *GitExec) IsGitRepo() bool {
    return PathExists(".git")
}

func (git *GitExec) HasChanges() bool {
    rc := OsExec{Dir: git.Path, LogOutput: true}
    stdOut, err := rc.Run("git diff-index HEAD --")
    if err != nil {
        return false;
    }
    if strings.TrimSpace(stdOut) == "" {
        return false;
    } else {
        return true;
    }
}

// If commitB is printed, then commitA is an ancestor of commit B
//"git rev-list <commitA> | grep $(git rev-parse <commitB>)"

func (git *GitExec) DetermineAncestor(commit1 string, commit2 string) string {
    rc := OsExec{Dir: git.Path, LogOutput: true}
    stdOut, err := rc.Run("git rev-list " + commit1)
    if err != nil {
        return "";
    }
    if !strings.Contains(stdOut, commit2) {
        stdOut, err = rc.Run("git rev-list " + commit2)
        if err != nil {
            return "";
        }
        if !strings.Contains(stdOut, commit1) {
            return ""
        } else {
            return commit2
        }
    } else {
        return commit1
    }
}

func (git *GitExec) GetRemoteUrl(remoteName string) (string, error) {
    gitCommand := "git remote -v"
    rc := OsExec{Dir: git.Path, LogOutput: true}
    stdOut, err := rc.Run(gitCommand)
    if err != nil {
        return "", err;
    }
    re := regexp.MustCompile(remoteName + "\\s(http.*\\.git)\\s")
    matched := re.FindAllStringSubmatch(stdOut, -1)
    if len(matched) == 0 {
        return "", errors.New("Could not find the remote " + remoteName)
    }
    if len(matched[1]) == 0 {
        return "", errors.New("Could not find the remote " + remoteName)
    }
    return matched[1][1], nil;
}

func (git *GitExec) GetLatestCommit() (string, error) {
    gitCommand := "git log --max-count=1 --pretty=format:%H"
    rc := OsExec{Dir: git.Path, LogOutput: true}
    stdOut, err :=rc.Run(gitCommand)
    if err != nil {
        return "", err;
    }
    return stdOut, nil;
}

func (git *GitExec) Init() error {
    fmt.Println("Initializing empty git repository")
    gitCommand := "git init";
    rc := OsExec{Dir: git.Path, LogOutput: true}
    _, err := rc.Run(gitCommand);
    return err;
}

func (git *GitExec) AddRemote(name string, url string) error {
    fmt.Println("Adding remote", url, "as", name)
    gitCommand := "git remote add " + name + " " + url
    rc := OsExec{Dir: git.Path, LogOutput: true}
    _, err := rc.Run(gitCommand)
    return err;
}

func (git *GitExec) InitAndFetch(url string) error {
    err := git.Init();
    if err != nil {
        return err;
    }
    err = git.AddRemote("origin", url)
    if err != nil {
        return err;
    }
    fmt.Println("Fetching...")
    gitCommand := "git fetch --all"
    rc := OsExec{Dir: git.Path, LogOutput: true}
    _, err = rc.Run(gitCommand);
    return err
}

func (git *GitExec) Checkout(commit string) (error) {
    fmt.Println("Checking out commit", commit)
    gitCommand := "git checkout " + commit
    rc := OsExec{Dir: git.Path, LogOutput: true}
    _, err := rc.Run(gitCommand);
    return err;
}

func (git *GitExec) SubmoduleUpdate(init bool, recursive bool) (error) {
    fmt.Println("Updating submodules...")
    gitCommand := "git submodule update ";
    if init {
        gitCommand = gitCommand + "--init "
    }
    if recursive {
        gitCommand = gitCommand + "--recursive"
    }
    rc := OsExec{Dir: git.Path, LogOutput: true}
    _, err := rc.Run(gitCommand)
    return err;
}

func (git *GitExec) InitAndCheckout(url string, commit string) error {
    err := git.InitAndFetch(url)
    if err != nil {
        return err;
    }
    err = git.Checkout(commit)
    if err != nil {
        return err;
    }
    return git.SubmoduleUpdate(true, true)
}

func (git *GitExec) Clone(url string, commit string) error {
    fmt.Println("Cloning", url, "...")
    // git clone <repo url> <destination directory>
    gitCommand := "git clone " + url + " " + commit
    rc := OsExec{Dir: git.Path, LogOutput: true}
    _, err := rc.Run(gitCommand);
    return err;
}
