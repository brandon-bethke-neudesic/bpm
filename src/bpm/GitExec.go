package main;

import (
    "strings"
    "bufio"
    "path"
    "os"
)

type GitExec struct {
    Path string
    LogOutput bool
}


func (git *GitExec) IsGitRepo() bool {
    return PathExists(".git")
}

func (git *GitExec) getChangedFiles() ([]string, error) {
    files := make([]string, 0);

    gitCommand := "git diff-index --name-only HEAD --"
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(gitCommand)
    if err != nil {
        return nil, err;
    }
    lineScanner := bufio.NewScanner(strings.NewReader(output))
    for lineScanner.Scan() {
        files = append(files, lineScanner.Text());
    }

    // Get untracked files


    gitCommand = "git ls-files -o"
    if PathExists(path.Join(git.Path, ".gitignore")){
        gitCommand = gitCommand + " --exclude-from=.gitignore"
    }
    rc = OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err = rc.Run(gitCommand)
    if err != nil {
        return nil, err;
    }
    lineScanner = bufio.NewScanner(strings.NewReader(output))
    for lineScanner.Scan() {
        files = append(files, lineScanner.Text());
    }
    return files, nil;
}

func (git *GitExec) LsFiles() ([]string, error) {
    files, err := git.getChangedFiles();
    if err != nil {
        return nil, err;
    }
    // Remove deleted files from the list
    gitCommand := "git ls-files -d"
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(gitCommand)
    if err != nil {
        return nil, err;
    }
    lineScanner := bufio.NewScanner(strings.NewReader(output))
    for lineScanner.Scan() {
        for i, item := range files {
            if item == lineScanner.Text() {
                files = append(files[:i], files[i+1:]...);
            }
        }
    }
    return files, nil;
}

func (git *GitExec) DeletedFiles() ([]string, error) {
    files := make([]string, 0);
    // Get deleted files
    gitCommand := "git ls-files -d"
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(gitCommand)
    if err != nil {
        return nil, err;
    }
    lineScanner := bufio.NewScanner(strings.NewReader(output))
    for lineScanner.Scan() {
        files = append(files, lineScanner.Text());
    }
    return files, nil;
}

func (git *GitExec) HasChanges() bool {
    files, err := git.getChangedFiles();
    if err != nil {
        return false;
    }
    return len(files) > 0
}

// If commitB is printed, then commitA is an ancestor of commit B
//"git rev-list <commitA> | grep $(git rev-parse <commitB>)"

func (git *GitExec) DetermineAncestor(commit1 string, commit2 string) string {
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
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

type GitRemote struct {
    Name string
    Url string
}

func (git *GitExec) Run(command string) (string, error) {;
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(command)
    if err != nil {
        return "", err;
    }
    return output, nil;
}

func (git *GitExec) RenameRemote(oldName string, newName string) error {
    _, err := git.Run("git remote rename " + oldName + " " + newName)
    return err;
}

func (git *GitExec) GetRemotes() ([]*GitRemote, error) {
    remotes := make([]*GitRemote, 0)

    gitCommand := "git remote"
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(gitCommand)
    if err != nil {
        return nil, err;
    }
    lineScanner := bufio.NewScanner(strings.NewReader(output))
    for lineScanner.Scan() {
        gitRemote := git.GetRemote(lineScanner.Text());
        remotes = append(remotes, gitRemote);
    }
    return remotes, nil;
}

func (git *GitExec) RemoteExists(name string) bool {
    gitCommand := "git remote get-url " + name
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand)
    if err != nil {
        return false;
    }
    return true;
}

func (git *GitExec) GetRemote(name string) (*GitRemote) {
    gitCommand := "git remote get-url " + name
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(gitCommand)
    if err != nil {
        return nil;
    }
    gitRemote := &GitRemote{Name: name}
    lineScanner := bufio.NewScanner(strings.NewReader(output))
    for lineScanner.Scan() {
        gitRemote.Url = lineScanner.Text();
    }
    return gitRemote;
}

func (git *GitExec) GetLatestCommit() (string, error) {
    gitCommand := "git log --max-count=1 --pretty=format:%H"
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    stdOut, err :=rc.Run(gitCommand)
    if err != nil {
        return "", err;
    }
    return strings.TrimSpace(stdOut), nil;
}

func (git *GitExec) AddSubmodule(url, location string) (error) {
    gitCommand := "git submodule add " + url + " " + location;
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);
    return err;
}

func (git *GitExec) DeleteSubmodule(location string) (error) {
    gitCommand := "git submodule deinit -f " + location;
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);

    if err == nil {
        gitCommand = "git rm -f " + location
        rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
        _, err = rc.Run(gitCommand);
    }

    if err == nil {
        gitCommand = "rm -rf .git/modules/" + location
        rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
        _, err = rc.Run(gitCommand);
    }

    if err == nil {
        os.RemoveAll(location)
    }
    return err;
}

func (git *GitExec) Init() error {
    gitCommand := "git init";
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);
    return err;
}

func (git *GitExec) GetCurrentBranch() (string, error) {
    gitCommand := "git symbolic-ref --short HEAD"
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    output, err := rc.Run(gitCommand)
    return strings.TrimSpace(output), err;
}

func (git *GitExec) DeleteRemote(name string) {
    gitCommand := "git remote remove " + name
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    rc.Run(gitCommand)
}

func (git *GitExec) AddRemote(name string, url string) error {
    gitCommand := "git remote add " + name + " " + url
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand)
    return err;
}

func (git *GitExec) Fetch(param string) error {
    gitCommand := "git fetch " + param
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);
    return err
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
    err = git.Fetch("--all");
    if err != nil {
        return err;
    }
    return nil;
}

func (git *GitExec) Checkout(name string) (error) {
    gitCommand := "git checkout " + name
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);
    return err;
}

func (git *GitExec) SubmoduleUpdate(init bool, recursive bool) (error) {
    gitCommand := "git submodule update ";
    if init {
        gitCommand = gitCommand + "--init "
    }
    if recursive {
        gitCommand = gitCommand + "--recursive"
    }
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
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

func (git *GitExec) Pull(remote string, branch string) error {
    gitCommand := "git pull " + remote + " " + branch;
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);
    return err;
}

func (git *GitExec) Clone(url string, name string) error {
    // git clone <repo url> <destination directory>
    gitCommand := "git clone " + url + " " + name
    rc := OsExec{Dir: git.Path, LogOutput: git.LogOutput}
    _, err := rc.Run(gitCommand);
    return err;
}
