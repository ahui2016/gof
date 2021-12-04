package util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"golang.org/x/crypto/blake2b"
)

// WrapErrors 把多个错误合并为一个错误.
func WrapErrors(allErrors ...error) (wrapped error) {
	for _, err := range allErrors {
		if err != nil {
			if wrapped == nil {
				wrapped = err
			} else {
				wrapped = fmt.Errorf("%v | %v", err, wrapped)
			}
		}
	}
	return
}

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

// FindFile returns a better error massage if cannot find the file.
func FindFile(name string) error {
	_, err := os.Lstat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the system cannot find: %s", name)
	}
	return err
}

func StrSliceFilter(arr []string, test func(string) bool) (result []string) {
	for _, s := range arr {
		if test(s) {
			result = append(result, s)
		}
	}
	return
}

// StrIndex returns the index of a string in the slice.
// returns -1 if not found.
func StrIndex(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// Sha256Hex 返回 sha256 的 hex 字符串。
// 虽然函数名是 Sha256, 但实际上采用 BLAKE2b 算法。
func Sha256Hex(data []byte) string {
	sum := blake2b.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// FileSha256Hex 返回文件 name 的 hex 字符串。
// 虽然函数名是 Sha256, 但实际上采用 BLAKE2b 算法。
func FileSha256Hex(name string) (string, error) {
	fileBytes, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return Sha256Hex(fileBytes), nil
}

// https://stackoverflow.com/questions/30376921/how-do-you-copy-a-file-in-go
func CopyFile(destPath, sourcePath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err1 := io.Copy(outputFile, inputFile)
	err2 := outputFile.Sync()
	return WrapErrors(err1, err2)
}

// IntMin computes the minimum of the two int args
func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
