package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		log.Fatal("Usage: go run main.go <path> [-f]")
	}

	path := os.Args[1] // path = .
	flagFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(os.Stdout, path, flagFiles)
	if err != nil {
		log.Fatal(err)
	}
}

func dirTree(out io.Writer, path string, flagFiles bool) error {
	if flagFiles {
		return printWithFiles(out, path, "", "")
	}

	return printWithoutFiles(out, path, "", "")
}

func printWithFiles(out io.Writer, fullpath, fullpref, dirname string) error {
	dirContent, err := os.ReadDir(fullpath)
	if err != nil {
		return err
	}

	sort.Slice(dirContent, func(i, j int) bool {
		return dirContent[i].Name() < dirContent[j].Name()
	})

	subpref := "├───"
	for index, entity := range dirContent {
		if index == len(dirContent)-1 {
			subpref = "└───"
		}

		if entity.IsDir() {
			fmt.Fprintf(out, "%s%s\n", fullpref, subpref+entity.Name())
		} else {
			x, _ := os.Stat(fullpath + "/" + entity.Name())
			switch size := x.Size(); size {
			case 0:
				fmt.Fprintf(out, "%s%s (empty)\n", fullpref, subpref+entity.Name())
			default:
				fmt.Fprintf(out, "%s%s (%db)\n", fullpref, subpref+entity.Name(), x.Size())
			}
		}

		fullprefArg := ""
		if index == len(dirContent)-1 {
			fullprefArg = fullpref + "\t"
		} else {
			fullprefArg = fullpref + "│\t"
		}
		err := error(nil)
		if entity.IsDir() {
			err = printWithFiles(out, fullpath+"/"+entity.Name(), fullprefArg, entity.Name())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type Dir struct {
	name  string
	index int
}

func printWithoutFiles(out io.Writer, fullpath, fullpref, dirname string) error {
	dirContent, err := os.ReadDir(fullpath)
	if err != nil {
		return err
	}

	sort.Slice(dirContent, func(i, j int) bool {
		return dirContent[i].Name() < dirContent[j].Name()
	})

	dirSet := make([]Dir, 0)
	lastDir := 0
	for i, entity := range dirContent {
		if entity.IsDir() {
			dirSet = append(dirSet, Dir{name: entity.Name(), index: i})
			lastDir = i
		}
	}

	subpref := "├───"
	for _, dir := range dirSet {
		if dir.index == lastDir {
			subpref = "└───"
		}

		fmt.Fprintf(out, "%s%s\n", fullpref, subpref+dir.name)

		fullprefArg := ""
		if dir.index == lastDir {
			fullprefArg = fullpref + "\t"
		} else {
			fullprefArg = fullpref + "│\t"
		}
		err := printWithoutFiles(out, fullpath+"/"+dir.name, fullprefArg, dir.name)
		if err != nil {
			return err
		}
	}
	return nil
}
