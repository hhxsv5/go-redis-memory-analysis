package storages

import (
	"io/ioutil"
	"os"
)

type File struct {
	filename string
	fp       *os.File
}

func NewFile(filename string, flag int, perm os.FileMode) (*File, error) {
	fp, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return &File{filename, fp}, nil
}

func (file *File) ReadAll() ([]byte, error) {
	return ioutil.ReadFile(file.filename)
}

func (file *File) WriteAll(data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(file.filename, data, perm)
}

func (file *File) Append(data []byte) (int, error) {
	length, err := file.fp.Write(data)
	return length, err
}

func (file *File) Truncate() error {
	fi, err := os.Stat(file.filename)
	if err != nil {
		return err
	}
	return file.fp.Truncate(fi.Size())
}

func (file *File) Close() {
	_ = file.fp.Close()
}
