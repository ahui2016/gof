package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ahui2016/go-rename/util"
)

// suffix 是用于临时文件名的后缀。
const suffix = "1"

// limit 限制最多可连续添加多少次 suffix，避免文件名无限变长。
const limit = 20

type Options struct {
	Verbose  bool
	Commands []string
}

// Task 用于对一个或多个文件执行 Options 里描述的操作。
type Task struct {
	Options Options
	Names   []string
}

type Tasks struct {
	AllTasks []Task `yaml:"all-tasks"`
}

// Exec 执行 Task 的默认操作 (即 Commands[0])。
func (t Task) Exec() error {
	if len(t.Options.Commands) == 0 {
		return fmt.Errorf("all-tasks.options.commands is empty")
	}
	if t.Options.Commands[0] != "swap" {
		return fmt.Errorf("not support：%s", t.Options.Commands[0])
	}
	if len(t.Names) != 2 {
		return fmt.Errorf("need two file name")
	}

	// 暂时只支持 swap 指令。对调 Names[0] 与 Names[1] 这两个文件的文件名。
	if t.Options.Verbose {
		log.Printf("start to swap [%s] and [%s]", t.Names[0], t.Names[1])
	}
	temp, err := tempName(t.Names[0])
	if err != nil {
		return err
	}

	if t.Options.Verbose {
		log.Printf("found a safe temp-file name: %s", temp)
		log.Printf("rename %s to %s", t.Names[0], temp)
	}
	if err := os.Rename(t.Names[0], temp); err != nil {
		return err
	}

	if t.Options.Verbose {
		log.Printf("rename %s to %s", t.Names[1], t.Names[0])
	}
	if err := os.Rename(t.Names[1], t.Names[0]); err != nil {
		return err
	}

	if t.Options.Verbose {
		log.Printf("rename %s to %s", temp, t.Names[1])
	}
	if err := os.Rename(temp, t.Names[1]); err != nil {
		return err
	}

	log.Printf("swap files OK: %s and %s", t.Names[0], t.Names[1])
	return nil
}

func (all Tasks) ExecAll() error {
	if len(all.AllTasks) == 0 {
		return fmt.Errorf("no task")
	}
	for _, task := range all.AllTasks {
		if err := task.Exec(); err != nil {
			return err
		}
	}
	log.Print("all tasks are finished.")
	return nil
}

// addSuffix 给一个文件名添加后缀，使其变成一个临时文件名。
// 比如 abc.js 处理后应变成 abc1.js
func addSuffix(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name + suffix
	}
	return name[:len(name)-len(ext)] + suffix + ext
}

// tempName 找出一个可用的临时文件名。
func tempName(name string) (string, error) {
	for i := 0; i < limit; i++ {
		name = addSuffix(name)
		ok, err := util.PathIsNotExist(name)
		if err != nil {
			return "", err
		}
		if ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("no proper temp-file name, last try: %s", name)
}
