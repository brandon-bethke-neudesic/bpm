package main;

import (
    "io/ioutil"
    "encoding/json"
    "bytes"
    "bpmerror"
    "github.com/blang/semver"
    "path"
    "time"
    "strconv"
)

type PackageJson struct {
    Path string
    Version semver.Version
    raw map[string]interface{}
}

func (pj *PackageJson) UpdateVersion() (error){
    if len(pj.Version.Pre) == 0 {
        str := strconv.FormatInt(time.Now().Unix(), 10)
        version, _ := semver.NewPRVersion(str)
        pj.Version.Pre = append(pj.Version.Pre, version);
    } else {
        str := strconv.FormatInt(time.Now().Unix(), 10)
        pj.Version.Pre[0], _ = semver.NewPRVersion(str);
    }
    return nil;
}

func (pj *PackageJson) Load() error {
    dat, err := ioutil.ReadFile(path.Join(pj.Path, "package.json"));
    if err != nil {
        return err
    }
    err = json.Unmarshal(dat, &pj.raw);
    if err != nil {
        return err;
    }

    version, _ := pj.raw["version"].(string);
    if version == "" {
        version = "0.0.1-1";
    }
    pj.Version, _ = semver.Make(version);
    return nil;
}

func (pj *PackageJson) Save() error {
    pj.raw["version"] = pj.Version.String();
    buf := new(bytes.Buffer)
    enc := json.NewEncoder(buf)
    enc.SetEscapeHTML(false)
    enc.SetIndent("", "    ");
    _ = enc.Encode(pj.raw)
    file := path.Join(pj.Path, "package.json");
    err := ioutil.WriteFile(file, buf.Bytes(), 0666);
    if err != nil {
        return bpmerror.New(err, "Error: There was an issue writing the file " + file)
    }

    return nil;
}