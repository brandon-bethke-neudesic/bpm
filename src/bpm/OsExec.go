
package main;

import  (
    "os/exec"
    "fmt"
    "bufio"
    "bytes"
)

type OsExec struct {
    Dir string
    LogOutput bool
}

func (rc *OsExec) Run(command string) (string, error) {
    if rc.LogOutput {
        fmt.Println(command);
    }
    cmd := exec.Command("sh", "-c", command);
    cmd.Dir = rc.Dir;

    stdOutDone := make(chan bool);
    stdErrDone := make(chan bool);

    var buffer bytes.Buffer
    stdout, err := cmd.StdoutPipe();
    stderr, err := cmd.StderrPipe();
    scannerOut := bufio.NewScanner(stdout)
    go func() {
        for scannerOut.Scan() {
            text := scannerOut.Text();
            if rc.LogOutput {
                fmt.Println(text)
            }
            buffer.WriteString(fmt.Sprintf("%s\n", text))
        }
        stdOutDone <- true;
    }()

    scannerErr := bufio.NewScanner(stderr)
    go func() {
        for scannerErr.Scan() {
            text := scannerErr.Text();
            if rc.LogOutput {
                fmt.Println(text)
            }
            buffer.WriteString(fmt.Sprintf("%s\n", text))
        }
        stdErrDone <- true;
    }()

    err = cmd.Start();
    if err == nil {
        err = cmd.Wait();
    }
    <-stdOutDone
    <-stdErrDone
    return buffer.String(), err;
}