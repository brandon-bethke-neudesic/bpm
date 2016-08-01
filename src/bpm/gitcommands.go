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

func (git *GitCommands) GetRemoteUrl(remoteName string) (string, error) {
    gitCommand := "git remote -v"
    stdOut, stdErr, err := git.RunCommand(git.Path, gitCommand)
    if err != nil {
        return stdErr, err;
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
    stdOut, stdErr, err := git.RunCommand(git.Path, gitCommand)
    if err != nil {
        return stdErr, err;
    }
    return stdOut, nil;
}


// git init
// git remote add origin http://blah
// git fetch origin
// git checkout 709c0b4b817b136c63c1896c6af10eb5962005bd

func (git *GitCommands) RunCommand(dir string, command string) (string, string, error) {
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
	if err != nil {
        fmt.Println(stdErr.String())
        fmt.Println(out.String())
        fmt.Println(err)
        return out.String(), stdErr.String(), err;
	}
    return out.String(), "", nil
}

func (git *GitCommands) Init() error {
    fmt.Println("Initializing empty git repository")
    gitCommand := "git init";
    _,stdErr, err := git.RunCommand(git.Path, gitCommand);
    if err != nil {
        fmt.Println(stdErr)
    }
    return err;
}

func (git *GitCommands) AddRemote(name string, url string) error {
    fmt.Println("Adding remote", url, "as", name)
    gitCommand := "git remote add " + name + " " + url
    _,stdErr, err := git.RunCommand(git.Path, gitCommand)
    if err != nil {
        fmt.Println(stdErr)
    }
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
    _,stdErr,err := git.RunCommand(git.Path, gitCommand);
    if err != nil {
        fmt.Println(stdErr)
    }
    return err
}

func (git *GitCommands) Checkout(commit string) (error) {
    fmt.Println("Checking out commit", commit)
    gitCommand := "git checkout " + commit
    _,stdErr, err := git.RunCommand(git.Path, gitCommand);
    if err != nil {
        fmt.Println(stdErr)
    }
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
    _, stdErr, err := git.RunCommand(git.Path, gitCommand)
    if err != nil {
        fmt.Println(stdErr)
    }
    return err;
}

func (git *GitCommands) InitAndCheckoutCommit(url string, commit string) error {
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
    _,_, err := git.RunCommand(git.Path, gitCommand);
    return err;
}
