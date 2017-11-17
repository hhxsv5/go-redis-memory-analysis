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

func (file *File) ReadAll() ([]byte, error) {
	return ioutil.ReadFile(file.filename)
}

func (file *File) WriteAll(data []byte, perm os.FileMode) (error) {
	return ioutil.WriteFile(file.filename, data, perm)
}

func (file *File) Append(data []byte, perm os.FileMode, clear bool) (int, error) {
	flag := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if clear {
		flag |= os.O_TRUNC
	}

	fp, err := os.OpenFile(file.filename, flag, perm)

	if err != nil {
		return 0, err
	}
	defer fp.Close()

	length, err := fp.Write(data)
	return length, err
}
