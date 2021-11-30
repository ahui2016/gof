package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ahui2016/go-rename/model"
	"github.com/ahui2016/go-rename/recipes"
	"github.com/ahui2016/go-rename/util"
	"gopkg.in/yaml.v2"
)

// 需要使用哪些 recipe, 要先在这里注册。
func initRecipes() error {
	return recipes.Register(
		new(recipes.Swap),
	// 在这里添加 recipe
	)
}

const defaultConfigFileName = "go-rename.yaml"

var config = flag.String("f", "", "a YAML config file")

func init() {
	initFlag()
	util.Panic(initRecipes())
}

func initFlag() {
	flag.Parse()
	if strings.TrimSpace(*config) == "" {
		ok, err := util.PathIsExist(defaultConfigFileName)
		util.Panic(err)
		if ok {
			*config = defaultConfigFileName
		} else {
			log.Fatal("Usage: go-rename -f string")
		}
	}
}

func main() {
	tasksFile, err := os.ReadFile(*config)
	util.Panic(err)
	tasks := model.Tasks{}
	util.Panic(yaml.Unmarshal(tasksFile, &tasks))
	util.Panic(tasks.ExecAll())
}
