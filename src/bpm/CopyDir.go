// Copy a whole directory tree. Current implementation is a packaged version of Jaybill McCarthy's code which can be found at http://jayblog.jaybill.com/post/id/26
package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type CopyDir struct {
	Exclude string
}

// Copies file source to destination dest.
func (cp *CopyDir) CopyFile(source string, dest string) (error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, si.Mode())
		}

	}
	return nil
}


func (cp *CopyDir) contains(s []string, e string) bool {
    for _, a := range s {
        if strings.Compare(a, e) == 0 {
            return true
        }
    }
    return false
}

// Recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func (cp *CopyDir) Copy(source string, dest string) (err error) {

	var excludeNames = []string{}
	if cp.Exclude != "" {
		excludeNames = strings.Split(cp.Exclude, "|")
	}

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return &CustomError{"Source is not a directory"}
	}
	// ensure dest dir does not already exist
	_, err = os.Open(dest)
	//if !os.IsNotExist(err) {
	//	return &CustomError{"Destination already exists"}
	//}
	// create dest dir
	//err = os.MkdirAll(dest, fi.Mode())
	//if err != nil {
	//	return err
	//}
	if os.IsNotExist(err) {
		// create dest dir
		err = os.MkdirAll(dest, fi.Mode())
		if err != nil {
			return err
		}
	}
	entries, err := ioutil.ReadDir(source)
	for _, entry := range entries {

		// Skip file names that match the excluded names
		if cp.contains(excludeNames, entry.Name()) {
			continue;
		}

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			copyDir := CopyDir{Exclude:cp.Exclude}
			err = copyDir.Copy(sfp, dfp);
			if err != nil {
				log.Println(err)
			}
		} else {
			// perform copy
			err = cp.CopyFile(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return
}

// A struct for returning custom error messages
type CustomError struct {
	What string
}

// Returns the error message defined in What as a string
func (e *CustomError) Error() string {
	return e.What
}
