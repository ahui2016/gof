package model

import (
	"fmt"
	"log"

	"github.com/ahui2016/gof/recipes"
)

type Task struct {
	Recipe  string
	Options map[string]string
	Names   []string // file/folder names
}

type Tasks struct {
	// file/folder names, 优先级比 Task 里的 Names 更高。
	Names    []string `yaml:"global-names"`
	AllTasks []Task   `yaml:"all-tasks"`
}

// ExecAll 当 realRun == true 时依次执行每个任务；
// 而当 realRun == false 时则只是依次检查每个任务，不会真的执行。
func (all Tasks) ExecAll(realRun bool) error {
	if len(all.AllTasks) == 0 {
		return fmt.Errorf("no task")
	}
	for _, task := range all.AllTasks {
		recipe, ok := recipes.Get[task.Recipe]
		if !ok {
			return fmt.Errorf("not found recipe: %s", task.Recipe)
		}
		if len(all.Names) > 0 {
			task.Names = all.Names
		}
		recipe.Refresh()
		recipe.Prepare(task.Names, task.Options)
		if err := recipe.Validate(); err != nil {
			return err
		}
		if realRun {
			if err := recipe.Exec(); err != nil {
				return err
			}
		}
	}
	if realRun {
		log.Print("all tasks are finished.")
	} else {
		log.Print("all tasks are validated.")
	}
	return nil
}
