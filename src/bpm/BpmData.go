package main;

import (
    "io/ioutil"
    "encoding/json"
    "github.com/blang/semver"
    "fmt"
    "strings"
    "bpmerror"
)

/*
{
    "name": "example",
    "version": "1.0.0",
    "dependencies": {
        "bpmdep1": {
            "commit": "cd4a1ae3fb81c7a0b032c5f359b0e0691be933a9",
            "url": "https://github.com/brandon-bethke-neudesic/bpmdep1.git"
        },
        "bpmdep2": {
            "commit": "abebc61f36b61e68d946392cf8457683ea20abc5",
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
