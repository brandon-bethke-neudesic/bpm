
package main;

import  (
    "strings"
    "errors"
    "os/exec"
    "fmt"
    "bytes"
)


type OsExec struct {
    Dir string
    LogOutput bool
}

func (rc *OsExec) Run(command string) (string, error) {
    if command == "" {
        return "", errors.New("The command cannot be empty")
    }
    splitCmd := strings.Split(command, " ")
    cmd := exec.Command(splitCmd[0])
    if rc.Dir != "" {
        cmd.Dir = rc.Dir;
    }
    cmd.Args = splitCmd;
    if rc.LogOutput {
        fmt.Println(cmd.Args)
    }
	var out bytes.Buffer
    var errOut bytes.Buffer
	cmd.Stdout = &out
    cmd.Stderr = &errOut
	err := cmd.Run()
    if rc.LogOutput {
        fmt.Println(out.String())
        fmt.Println(errOut.String())
    }
	if err != nil {
        return out.String() + errOut.String(), err;
	}
    return out.String(), nil
}
