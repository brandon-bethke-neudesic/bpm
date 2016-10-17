package main;

import (
    "os/exec"
    "strings"
    "bytes"
    "fmt"
    "regexp"
    "errors"
)

type GitCommands struct {
    Path string
}


func (git *GitCommands) HasChanges() bool {
    stdOut, err := git.RunCommand(git.Path, "git diff-index HEAD --")
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

func (git *GitCommands) DetermineAncestor(commit1 string, commit2 string) string {
    stdOut, err := git.RunCommand(git.Path, "git rev-list " + commit1)
    if err != nil {
        return "";
    }
    if !strings.Contains(stdOut, commit2) {
        stdOut, err = git.RunCommand(git.Path, "git rev-list " + commit2)
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

func (git *GitCommands) GetRemoteUrl(remoteName string) (string, error) {
    gitCommand := "git remote -v"
    stdOut, err := git.RunCommand(git.Path, gitCommand)
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

func (git *GitCommands) GetLatestCommit() (string, error) {
    gitCommand := "git log --max-count=1 --pretty=format:%H"
    stdOut, err := git.RunCommand(git.Path, gitCommand)
    if err != nil {
        return "", err;
    }
    return stdOut, nil;
}

func (git *GitCommands) RunCommand(dir string, command string) (string, error) {
    cmd := exec.Command("git")
    if dir != "" {
        cmd.Dir = dir;
    }
    cmd.Args = strings.Split(command, " ");
    fmt.Println(cmd.Args)
	var out bytes.Buffer
    var stdErr bytes.Buffer
	cmd.Stdout = &out
    cmd.Stderr = &stdErr
	err := cmd.Run()
    fmt.Println(out.String())
    fmt.Println(stdErr.String())
	if err != nil {
        fmt.Println(err)
        return out.String(), err;
	}
    return out.String(), nil
}

func (git *GitCommands) Init() error {
    fmt.Println("Initializing empty git repository")
    gitCommand := "git init";
    _, err := git.RunCommand(git.Path, gitCommand);
    return err;
}

func (git *GitCommands) AddRemote(name string, url string) error {
    fmt.Println("Adding remote", url, "as", name)
    gitCommand := "git remote add " + name + " " + url
    _, err := git.RunCommand(git.Path, gitCommand)
    return err;
}

func (git *GitCommands) InitAndFetch(url string) error {
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
    _, err = git.RunCommand(git.Path, gitCommand);
    return err
}

func (git *GitCommands) Checkout(commit string) (error) {
    fmt.Println("Checking out commit", commit)
    gitCommand := "git checkout " + commit
    _, err := git.RunCommand(git.Path, gitCommand);
    return err;
}

func (git *GitCommands) SubmoduleUpdate(init bool, recursive bool) (error) {
    fmt.Println("Updating submodules...")
    gitCommand := "git submodule update ";
    if init {
        gitCommand = gitCommand + "--init "
    }
    if recursive {
        gitCommand = gitCommand + "--recursive"
    }
    _, err := git.RunCommand(git.Path, gitCommand)
    return err;
}

func (git *GitCommands) InitAndCheckout(url string, commit string) error {
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

func (git *GitCommands) Clone(url string, commit string) error {
    fmt.Println("Cloning", url, "...")
    // git clone <repo url> <destination directory>
    gitCommand := "git clone " + url + " " + commit
    _, err := git.RunCommand(git.Path, gitCommand);
    return err;
}
