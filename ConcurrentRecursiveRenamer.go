/*
	Author: Luís Ferreirinha N51127
*/

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Stores info about a file that is going to be renamed by main
type FileRenamingInfo struct {
	depth        int
	originalPath string
	newPath      string
}

func openAndReadDir(path string) []fs.FileInfo {
	// Attempt to open directory
	rootFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	// Attempt to read directory contents
	files, err := rootFile.Readdir(0)
	if err != nil {
		fmt.Println(err)
	}
	return files
}

func renamer(depth int, parentDir string, newParentDir string, rootFile fs.FileInfo, targetString string, replaceString string, report chan<- []FileRenamingInfo) {

	// List of renaming operations
	var listRenames []FileRenamingInfo

	// Increase current depth level
	depth++

	// New file name
	newFileName := strings.ReplaceAll(rootFile.Name(), targetString, replaceString)

	// Since we need to read the contents of this directory if the file is a directory,
	// we need to keep track of the original directory path before the renaming operations
	currentPath := filepath.Join(parentDir, rootFile.Name())

	// Since the parent directory might be renamed before this file is renamed, we need to pass down the new directory name
	// so that all new paths are valid when the file renaming is done sequentially by the main thread
	nextParentDir := filepath.Join(newParentDir, rootFile.Name())

	// If newFilename is different than original filename we create a new path and append it to list of renaming operations
	if newFileName != rootFile.Name() {
		// Calculate new file path if it contains targetString
		newPath := filepath.Join(newParentDir, newFileName)
		// Original file path after the parent directory is renamed by main thread
		newParentOldFileName := nextParentDir
		listRenames = append(listRenames, FileRenamingInfo{depth, newParentOldFileName, newPath})
		// This is the new parent directory for the next thread
		nextParentDir = newPath
	}

	// If current file is a directory we iterate throught the files and create new Goroutines
	if rootFile.IsDir() {

		files := openAndReadDir(currentPath)

		child_report := make(chan []FileRenamingInfo)
		routines := 0

		for _, file := range files {
			go renamer(depth, currentPath, nextParentDir, file, targetString, replaceString, child_report)
			routines++
		}

		// Wait for all child routines to report and append their reported files to listRenames
		for i := 0; i < routines; i++ {
			childListRenames := <-child_report
			for _, fileRename := range childListRenames {
				listRenames = append(listRenames, fileRename)
			}
		}
	}

	// Report to parent directory thread that this routine finished its job
	report <- listRenames

}

func main() {

	root := os.Args[1]
	targetString := os.Args[2]
	replaceString := os.Args[3]

	files := openAndReadDir(root)

	// One directional channel for routines to report when they are done
	report := make(chan []FileRenamingInfo)
	routines := 0

	// Depth in the current file tree (root = 0)
	depth := 0

	// for each file we create a new goroutine
	for _, file := range files {
		go renamer(depth, root, root, file, targetString, replaceString, report)
		routines++
	}

	var allRenamingOperations []FileRenamingInfo

	// wait for all routines to report their found files
	for i := 0; i < routines; i++ {
		childListRenames := <-report
		for _, fileRename := range childListRenames {
			allRenamingOperations = append(allRenamingOperations, fileRename)
		}
	}

	// Sort slice of FileRenamingInfo by level
	// This way main will rename files in higher levels first, before renaming those in lower levels
	// and avoids conflicts in renaming because the calculated newPaths in lower levels presume that the
	// higher levels have been renamed
	sort.Slice(allRenamingOperations, func(i, j int) bool {
		return allRenamingOperations[i].depth < allRenamingOperations[j].depth
	})

	// Sequentially rename all found target files
	for _, fileRename := range allRenamingOperations {
		//fmt.Printf("Orignal path: %s\nNew Path: %s\n", fileRename.originalPath, fileRename.newPath)
		err := os.Rename(fileRename.originalPath, fileRename.newPath)
		if err != nil {
			fmt.Println(err)
		}
	}

}
