package utils

import (
	"io/ioutil"
	"os"
)

// RemoveFiles deletes all files in given directory, except those directories on ignore list.
func RemoveFiles(dir string) (err error) {

	dirContents, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fileOrDir := range dirContents {
		fileOrDirFullPath := dir + string(os.PathSeparator) + fileOrDir.Name()
		if fileOrDir.IsDir() {
			if allowDeleteDir(fileOrDir.Name()) {
				err = RemoveFiles(fileOrDirFullPath)
			}
		} else {
			err = os.Remove(fileOrDirFullPath)
		}
	}

	return err
}

// Check if directory can be deleted.
func allowDeleteDir(dir string) bool {
	allowed := true

	switch {
	case dir == ".svn":
		allowed = false
		// Here add cases for other directories, which contents should not be deleted.
	}

	return allowed
}
