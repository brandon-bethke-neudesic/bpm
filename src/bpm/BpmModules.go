package main;

import (
    "sort"
    "io/ioutil"
    "bpmerror"
    "path"
)

type BpmModules struct {
    Dependencies map[string]*BpmDependency
}

func (bpm *BpmModules) HasDependency(name string) bool {
    _, exists := bpm.Dependencies[name];
    return exists;
}

func (bpm *BpmModules) HasDependencies() bool {
    return len(bpm.Dependencies) > 0
}

func (bpm *BpmModules) GetSortedKeys() []string {
    sortedKeys := make([]string, len(bpm.Dependencies))
    i := 0
    for k, _ := range bpm.Dependencies {
        sortedKeys[i] = k
        i++
    }
    sort.Strings(sortedKeys)
    return sortedKeys;
}

func (bpm *BpmModules) Load(location string) (error) {
    bpm.Dependencies = make(map[string]*BpmDependency, 0)
    // If the path does not exist, then there are no dependencies. It is not an error to have no depenencies.
    if !PathExists(location) {
        return nil;
    }
    files, err := ioutil.ReadDir(location)
    if err != nil {
        return bpmerror.New(err, "Error: Could not read the bpm modules folder at " + location)
    }
    for _, f := range files {
        if f.IsDir() {
            name := f.Name();
            item := &BpmDependency{Name:name, Path: path.Join(location, name)}
            bpm.Dependencies[name] = item;
        }
    }
    return nil;
}