package main;

import (
    "io/ioutil"
    "encoding/json"
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

type BpmDependency struct {
    Commit string `json:"commit"`
    Url    string `json:"url"`
}

type BpmData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
    Dependencies map[string]BpmDependency `json:"dependencies"`
}

func (bpm *BpmData) HasDependencies() bool {
    return len(bpm.Dependencies) > 0
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
