
package main;

import  (
    "os/exec"
    "fmt"
    "bytes"
    "io"
    "os"
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

    var buffer bytes.Buffer
    var w io.Writer;
	if rc.LogOutput {
		w = io.MultiWriter(os.Stdout, &buffer)
	} else {
		w = io.Writer(&buffer);
	}
	cmd.Stdout = w;
	cmd.Stderr = w;
    err := cmd.Run();
    if err != nil {
		return "", err;
    }
    return buffer.String(), err;
}