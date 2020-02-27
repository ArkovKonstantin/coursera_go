package main

import (
	"fmt"
	"io"
	"io/ioutil"
	//"log"
	"os"
	"sort"
	"strings"
)

func filter(files *[]os.FileInfo) {
	n := 0
	for _, v := range *files {
		if v.IsDir() {
			(*files)[n] = v
			n++
		}
	}
	*files = (*files)[:n]
}

func getPref(path string, lastDirCount int, lastItem bool) string {
	c := strings.Count(path, "/") - 1
	var pref string
	for i := 0; i < c; i++ {
		if i < c-lastDirCount {
			pref += "│\t"
		} else {
			pref += "\t"
		}
	}
	if lastItem == true {
		pref += "└───"
	} else {
		pref += "├───"
	}
	return pref
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	lastDirCount := strings.Count(path, "&")
	path = path[:len(path)-lastDirCount]
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	// check printFiles
	if !printFiles {
		filter(&files)
	}
	//sorting
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	for i, file := range files {
		newP := path + string(os.PathSeparator) + file.Name()
		if file.IsDir() {
			var lastItem bool
			if i == len(files)-1 {
				lastItem = true
				newP += "&" // add suffix indicating last dir
			}
			//fmt.Printf("%s\n", getPref(newP, lastDirCount, lastItem) + file.Name())
			fmt.Fprintf(out, "%s\n", getPref(newP, lastDirCount, lastItem)+file.Name())
			for j := 0; j < lastDirCount; j++ { // recover suffixes
				newP += "&"
			}
			err = dirTree(out, newP, printFiles)
			if err != nil {
				return err
			}
		} else {
			if printFiles {
				var lastItem bool
				if i == len(files)-1 {
					lastItem = true
				}

				if file.Size() == 0 {
					//fmt.Printf("%s (%s)\n", getPref(newP,lastDirCount, lastItem) + file.Name(), "empty")
					fmt.Fprintf(out, "%s (%s)\n", getPref(newP, lastDirCount, lastItem)+file.Name(), "empty")
				} else {
					//fmt.Printf("%s (%db)\n", getPref(newP,lastDirCount, lastItem) + file.Name(), file.Size())
					fmt.Fprintf(out, "%s (%db)\n", getPref(newP, lastDirCount, lastItem)+file.Name(), file.Size())
				}
			}
		}
	}
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}