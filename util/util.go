package util

import (
	"errors"
	"io/fs"
	"os"
)

// Panic panics if err != nil
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// PathIsNotExist 找不到名为 name 的文件时返回 true, 否则返回 false
func PathIsNotExist(name string) (ok bool, err error) {
	_, err = os.Lstat(name)
	if errors.Is(err, fs.ErrNotExist) {
		ok = true
		err = nil
	}
	return
}

// PathIsExist 找到名为 name 的文件时返回 true, 否则返回 false
func PathIsExist(name string) (bool, error) {
	ok, err := PathIsNotExist(name)
	return !ok, err
}
