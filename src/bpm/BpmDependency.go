package main;

import (
)

type BpmDependency struct {
    Url    string `json:"url"`
}

func (dep *BpmDependency) Validate() (error) {
    return nil;
}

func (dep *BpmDependency) Equal(item *BpmDependency) bool {
    // If the address is the same, then it's equal.
    if dep == item {
        return true;
    }
    if item.Url == dep.Url {
        return true;
    }
    return false;
}
