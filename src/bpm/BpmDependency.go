package main;

import (
    "strings"
    "bpmerror"
)

type BpmDependency struct {
    Commit string `json:"commit"`
    Url    string `json:"url"`
}

func (dep *BpmDependency) Validate() (error) {
    if strings.TrimSpace(dep.Url) == "" {
        return bpmerror.New(nil, "Error: No url specified")
    }
    return nil
}

func (dep *BpmDependency) Equal(item *BpmDependency) bool {
    // If the address is the same, then it's equal.
    if dep == item {
        return true;
    }
    if item.Commit == dep.Commit && item.Url == dep.Url {
        return true;
    }
    return false;
}
