package main

import (
	"os"
)

type FileSystem interface {
	Create(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

type RealFileSystem struct{}

func (fs RealFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (fs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
