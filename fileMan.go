package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func getFile(path string) (string, string) {
	var iName string
	var cmd string

	for len(iName) == 0 {
		files, err := os.ReadDir(path)
		if err != nil {
			fmt.Println("Error Reading Current Directory")
			return iName, path
		}

		showFiles(files, &path)
		fmt.Print("Note: Scupper file names cannot have spaces.\nLine #: ")
		fmt.Scanln(&cmd)
		if isInt(cmd) {
			num, er := strconv.Atoi(cmd)
			if er != nil {
				fmt.Println("getFileName(): " + er.Error())
			} else if files[num-1].IsDir() {
				path = filepath.Join(path, files[num-1].Name())
			} else {
				if num <= len(files) && num > 0 {
					iName = files[num-1].Name()
				}
			}
		} else if cmd == ".." {
			path = filepath.Dir(path)
		}
	}

	return iName, path
}

func showFiles(files []os.DirEntry, dir *string) {
	fmt.Println("\nFile List: " + *dir)
	fmt.Println("    ..  Go to parent directory")
	for i, entry := range files {
		info, _ := entry.Info()
		tm := info.ModTime()
		yr, mo, day := tm.Date()
		fmt.Printf("    %2d  %4d-%02d-%02d  %02d:%02d  %s", i+1, yr, mo, day,
			tm.Hour(), tm.Minute(), entry.Name())
		if entry.IsDir() {
			fmt.Println("/")
		} else {
			fmt.Print("\n")
		}
	}
}

func openFile(name string, path string, mode int) *os.File {
	fmt.Println("Opening: " + filepath.Join(path, name))
	file, e := os.OpenFile(filepath.Join(path, name), mode, 0644)
	if e != nil {
		fmt.Println("Error opening file: " + e.Error())
		return nil
	}

	return file
}
