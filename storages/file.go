package storages

import (
	"io/ioutil"
	"os"
)

type File struct {
	filename string
}

func NewFile(filename string) (*File) {
	return &File{filename}
}

func (file *File) Read(length uint64) ([]byte, error) {
	return ioutil.ReadFile(file.filename)
}

func (file *File) Write(data []byte, perm os.FileMode) (error) {
	return ioutil.WriteFile(file.filename, data, perm)
}
