package main

import (
	"fmt"
	"io"
	"io/ioutil"
	//"log"
	"os"
	"sort"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	seq := make([]int, 0)
	err := printTree(out, path, printFiles, seq)
	if err != nil {
		return err
	}
	return nil
}

func printTree(out io.Writer, path string, printFiles bool, seq []int) error {
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
		prefix := ""
		for _, j := range seq {
			if j == 1 {
				prefix += "\t"
			} else if j == 0 {
				prefix += "|\t"
			}
		}
		if i == len(files)-1 {
			prefix += "└───"
		} else {
			prefix += "├───"
		}
		if file.IsDir() {
			fmt.Fprintf(out, "%s\n", prefix+file.Name())
			if i == len(files)-1 {
				err := printTree(out, path+string(os.PathSeparator)+file.Name(), printFiles, append(seq, 1))
				if err != nil {
					return nil
				}
			} else {
				err := printTree(out, path+string(os.PathSeparator)+file.Name(), printFiles, append(seq, 0))
				if err != nil {
					return nil
				}
			}
		} else {
			if file.Size() == 0 {
				fmt.Fprintf(out, "%s (%s)\n", prefix+file.Name(), "empty")
			} else {
				fmt.Fprintf(out, "%s (%db)\n", prefix+file.Name(), file.Size())
			}
		}
	}
	return nil
}

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
