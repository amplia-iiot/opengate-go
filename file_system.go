package connectorog

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/amplia-iiot/opengate-go/logger"
)

// open/read a file from localpath.
// If the local file is not found, it will returned the embedbed version the file.
// The embebed file is built when the app is compiled
func OpenFile(filename, path string, embedfiles embed.FS) (fs.File, error) {
	filePath := fmt.Sprintf("%s/%s", path, filename)
	if Exists(filePath) {
		return os.Open(filePath)
	}
	return embedfiles.Open(filename)
}

// read all dir path and return the directory info, with files.
// If the dir is in filesystem return this version but return its embebed copy
// ** relativePath: dir relativePath from the main execution is running. Ej.: data_models/packs
func ReadDir(relativePath string, embedfiles embed.FS) ([]fs.DirEntry, error) {
	if Exists(relativePath) {
		return os.ReadDir(relativePath)
	}
	return embedfiles.ReadDir(".")
}

// return if a file exists.
// path is the path of the file
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	logger.Warn("error reading stat of file:", path, err.Error())
	return false
}
func CreateAFile(fileName string, f embed.FS, embedPath string) (err error) {
	if err = EnsureConfigFolder(embedPath); err != nil {
		return
	}
	return CreateIfNotExist(fileName, f, embedPath)
}
func CreateFiles(f embed.FS, embedPath string) (err error) {
	if err = EnsureConfigFolder(embedPath); err != nil {
		return
	}
	return CreateIfNotExist("", f, embedPath)
}

func CreateOrOverrideFiles(f embed.FS, embedPath string) (err error) {
	os.RemoveAll(embedPath)
	if err = EnsureConfigFolder(embedPath); err != nil {
		return
	}
	return CreateIfNotExist("", f, embedPath)
}

// if fileName is not empty only create de specific file. If fileName is empty create all file in f and embedPath
func CreateIfNotExist(fileName string, f embed.FS, embedPath string) (err error) {
	dir, err := f.ReadDir(".")
	for _, file := range dir {
		if (fileName != "" && fileName == file.Name()) || fileName == "" {
			filePath := fmt.Sprintf("%s/%s", embedPath, file.Name())
			if isCreated := Exists(filePath); !isCreated {
				fileContent, _ := f.ReadFile(file.Name())
				if err = os.WriteFile(filePath, fileContent, fs.ModePerm); err != nil {
					return
				}
			}
		}
	}
	return err
}

func EnsureConfigFolder(embedPath string) error {
	isCreated := Exists(embedPath)
	if !isCreated {
		if err := os.MkdirAll(embedPath, fs.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// The files can be a mixture of embebed files and files present in filesystem.
// read all config path and return the directory info, with files.
// The files can be a mixture of embebed files and files present in filesystem.
// If the file is in filesystem return this version but return its embebed copy
// **dirpath: Ej.: configs/backend
func ReadAll(f embed.FS, dirpath string) ([]fs.DirEntry, error) {
	onmemoryFiles, err := f.ReadDir(".")
	if err != nil {
		return nil, err
	}
	ondiskFiles, err := os.ReadDir(dirpath)
	if err != nil {
		ondiskFiles = make([]fs.DirEntry, 0)
	}
	return MixOnMemoryOnDiskFiles(onmemoryFiles, ondiskFiles), nil
}
func MixOnMemoryOnDiskFiles(onmemoryFiles, ondiskfiles []fs.DirEntry) []fs.DirEntry {
	mixed := ondiskfiles[:]
	for _, file := range onmemoryFiles {
		found := false
		for _, loadedFile := range mixed {
			if loadedFile.Name() == file.Name() {
				found = true
				break
			}
		}
		if !found {
			mixed = append(mixed, file)
		}
	}
	return mixed
}

// typeFile: ".yaml" or ".json", etc
func FilterFiles(typeFile string, files []fs.DirEntry) []fs.DirEntry {
	var jsfiles []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), typeFile) {
			jsfiles = append(jsfiles, file)
		}
	}

	return jsfiles
}

func WriteInFolder(info []byte, fileFolder, fileName string) error {
	mkdirErr := os.MkdirAll(fileFolder, os.ModePerm)
	if mkdirErr != nil {
		return mkdirErr
	}
	return os.WriteFile(fileFolder+fileName, info, 0644)
}
