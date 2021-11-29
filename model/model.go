package model

import (
	"fmt"
	"log"
)

// Task 用于对一个或多个文件执行 Options 里描述的操作。
type Task struct {
	Recipe  string
	Options map[string]string
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
