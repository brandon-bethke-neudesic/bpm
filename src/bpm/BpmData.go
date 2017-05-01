package main;

import (
    "io/ioutil"
    "encoding/json"
    "github.com/blang/semver"
    "fmt"
    "strings"
    "bpmerror"
    "sort"
    "path"
)

/*
{
    "name": "example",
    "version": "1.0.0",
    "dependencies": {
        "bpmdep1": {
            "url": "../bpmdep1.git"
        },
        "bpmdep2": {
            "url": "https://github.com/brandon-bethke-neudesic/bpmdep2.git"
        }
    }
}
*/

type BpmData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
    Dependencies map[string]*BpmDependency `json:"dependencies"`
}

func LoadBpmData(source string) (*BpmData, error) {
    bpm := &BpmData{};
    bpmJsonFile := path.Join(source, Options.BpmFileName);
    err := bpm.LoadFile(bpmJsonFile);
    if err != nil {
        return nil, bpmerror.New(err, "Error: Could not load the bpm.json file for dependency " + source)
    }
    return bpm, nil;
}

func (bpm *BpmData) IncrementVersion() (error){
    bpmVersion, err := semver.Make(bpm.Version);
    if err != nil {
        return bpmerror.New(nil, "Error: Could not read the version field from the bpm file")
    }
    bpmVersion.Patch++;
    bpm.Version = bpmVersion.String();
    return nil;
}

func (bpm *BpmData) HasDependency(name string) bool {
    _, exists := bpm.Dependencies[name];
    return exists;
}

func (bpm *BpmData) HasDependencies() bool {
    return len(bpm.Dependencies) > 0
}

func (bpm *BpmData) String() string {
    bytes, err := json.MarshalIndent(bpm, "", "   ")
    if err != nil {
        fmt.Println(err)
        return "";
    }

    return string(bytes);
}

func (bpm *BpmData) GetSortedKeys() []string {
    sortedKeys := make([]string, len(bpm.Dependencies))
    i := 0
    for k, _ := range bpm.Dependencies {
        sortedKeys[i] = k
        i++
    }
    sort.Strings(sortedKeys)
    return sortedKeys;
}

func (bpm *BpmData) WriteFile(file string) error {
    bytes, err := json.MarshalIndent(bpm, "", "   ")
    if err != nil {
        return err;
    }

    err = ioutil.WriteFile(file, bytes, 0666);
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue writing the file " + file)
    }
    return nil;
}

func (bpm *BpmData) Validate() error {
    if strings.TrimSpace(bpm.Name) == "" {
        return bpmerror.New(nil, "Error: There must be a name field in the bpm.json file")
    }

    if strings.TrimSpace(bpm.Version) == "" {
        return bpmerror.New(nil, "Error: There must be a version field in the bpm.json file")
    }
    return nil;
}

func (bpm *BpmData) LoadFile(file string) error {
    dat, err := ioutil.ReadFile(file)
    if err != nil {
        return err
    }
    jsondata := BpmData{};
    err = json.Unmarshal(dat, &jsondata);
    if err != nil {
        return err
    }
    bpm.Dependencies = jsondata.Dependencies;
    bpm.Name = jsondata.Name
    bpm.Version = jsondata.Version
    return nil;
}

func (bpm *BpmData) Clone(includeModules string) *BpmData {
    newBpm := &BpmData{};
    newBpm.Name = bpm.Name;
    newBpm.Version = bpm.Version;
    newBpm.Dependencies = make(map[string]*BpmDependency);
    for name, v := range bpm.Dependencies {
        newBpm.Dependencies[name] = v
    }

    if includeModules == "all" {
        return newBpm;
    }
    whitelist := strings.Split(includeModules, ",")
    for _, name := range whitelist {
        delete(newBpm.Dependencies, name)
    }
    return newBpm;
}
