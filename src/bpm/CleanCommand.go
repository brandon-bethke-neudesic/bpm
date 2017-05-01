package main;

import (
    "fmt"
    "os"
    "path"
    "bpmerror"
    "errors"
    "io/ioutil"
)

type CleanCommand struct {
}

func (cmd *CleanCommand) Name() string {
    return "clean"
}

func (cmd *CleanCommand) Trim() (error) {
    if !Options.BpmFileExists() {
        return errors.New("Error: The " + Options.BpmFileName + " file does not exist.");
    }
    bpm := &BpmData{};
    err := bpm.LoadFile(Options.BpmFileName);
    if err != nil {
        return bpmerror.New(nil, "Error: There was a problem loading the bpm.json file")
    }
    if !bpm.HasDependencies() {
        fmt.Println("There are no dependencies. Done.")
        return nil;
    }

    cache := &CleanCache{Items:make([]*CleanCacheItem, 0)};
    err = cache.Build(bpm)
    entries, _ := ioutil.ReadDir(Options.BpmCachePath)
    for _, entry := range entries {
        if entry.IsDir() {
            if !cache.NameExists(entry.Name()) {
                pathToRemove := path.Join(Options.BpmCachePath, entry.Name());
                fmt.Println("Removing item " + pathToRemove)
                os.RemoveAll(pathToRemove);
                continue;
            }
            commitEntries, _ := ioutil.ReadDir(path.Join(Options.BpmCachePath, entry.Name()))
            for _, commitEntry := range commitEntries {
                if commitEntry.IsDir() && !cache.Exists(entry.Name(), commitEntry.Name()) {
                    pathToRemove := path.Join(Options.BpmCachePath, entry.Name(), commitEntry.Name());
                    fmt.Println("Removing item " + pathToRemove)
                    os.RemoveAll(pathToRemove)
                }
            }
        }
    }
    return nil;
}

func (cmd *CleanCommand) Clean() (error) {
    os.RemoveAll(Options.BpmCachePath);
    return nil;
}


func (cmd *CleanCommand) Execute() (error) {
    if Options.Trim {
        return cmd.Trim();
    }
    return cmd.Clean();
}

type CleanCache struct {
    Items []*CleanCacheItem
}

type CleanCacheItem struct {
    Name string
    Commit string
}

func (tc *CleanCache) Add(item *CleanCacheItem) {
    if !tc.Exists(item.Name, item.Commit) {
        tc.Items = append(tc.Items, item);
    }
}

func (tc *CleanCache) NameExists(name string) bool {
    if tc.Items == nil || len(tc.Items) == 0 {
        return false;
    }
    for _, item := range tc.Items {
        if item.Name == name {
            return true;
        }
    }
    return false;
}

func (tc *CleanCache) Exists(name string, commit string) bool {
    if tc.Items == nil || len(tc.Items) == 0 {
        return false;
    }
    for _, item := range tc.Items {
        if item.Name == name && item.Commit == commit {
            return true;
        }
    }
    return false;
}

func (tc *CleanCache) Build(bpm *BpmData) (error) {
    // Always process the keys sorted by name so the processing is consistent
    sortedKeys := bpm.GetSortedKeys();
    for _, depName := range sortedKeys {
        depItem := bpm.Dependencies[depName]
        // Delete previous cached items
        tcItem := &CleanCacheItem{Name: depName, Commit: depItem.Commit}
        tc.Add(tcItem);
        moduleBpm := &BpmData{};
        moduleBpmFilePath := path.Join(Options.BpmCachePath, depName, depItem.Commit, Options.BpmFileName);
        // It should be expected that the bpm.json file may not exist and this isn't a fatal error, just move on.
        err := moduleBpm.LoadFile(moduleBpmFilePath);
        if err != nil {
            return nil;
        }
        return tc.Build(moduleBpm);
    }
    return nil
}
