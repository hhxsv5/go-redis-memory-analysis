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

func (file *File) Read() ([]byte, error) {
	return ioutil.ReadFile(file.filename)
}

func (file *File) Write(data []byte, perm os.FileMode) (error) {
	return ioutil.WriteFile(file.filename, data, perm)
}

func (file *File) Append(data []byte, perm os.FileMode) (int, error) {
	fp, err := os.OpenFile(file.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, perm)

	if err != nil {
		return 0, err
	}
	defer fp.Close()

	length, err := fp.Write(data)
	return length, err
}
