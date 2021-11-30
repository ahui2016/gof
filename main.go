package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ahui2016/gof/model"
	"github.com/ahui2016/gof/recipes"
	"github.com/ahui2016/gof/util"
	"gopkg.in/yaml.v2"
)

// 需要使用哪些 recipe, 要先在这里注册。
func initRecipes() error {
	return recipes.Register(
		new(recipes.Swap),
	// 在这里添加 recipe
	)
}

const defaultConfigFileName = "gof.yaml"

var (
	tasks      model.Tasks
	useDefault bool
)

var (
	// YAML 文件名
	config = flag.String("f", "", "a YAML config file")

	// -r 的优先级高于 -f (即，如果指定了 -r, 就忽略 -f)
	recipe = flag.String("r", "", "use a recipe with default options")

	// filenames, 优先级高于 YAML 文件里的 names
	names []string
)

func init() {
	initFlag()
	util.Panic(initRecipes())
}

func initFlag() {
	flag.Parse()
	names = flag.Args()

	// 如果命令行指定了 recipe 名称，则不需要 YAML 文件
	if *recipe != "" {
		useDefault = true
		tasks = model.Tasks{AllTasks: []model.Task{{
			Recipe: *recipe,
			Names:  names,
		}}}
	} else {
		// 如果命令行未指定 recipe, 则需要一个 YAML 文件，
		// 如果用户未指定 YAML 文件，则尝试寻找默认的 YAML 文件。
		if strings.TrimSpace(*config) == "" {
			ok, err := util.PathIsExist(defaultConfigFileName)
			util.Panic(err)
			if ok {
				*config = defaultConfigFileName
			} else {
				log.Fatalf("Usage Example:\n    gof -f example.yaml\n    gof -r swap file1 file2")
			}
		}
		tasksFile, err := os.ReadFile(*config)
		util.Panic(err)
		util.Panic(yaml.Unmarshal(tasksFile, &tasks))
	}

	// 如果命令行输入了文件名，则用来覆盖 YAML 文件里的全部文件名
	if len(names) > 0 {
		for i := range tasks.AllTasks {
			tasks.AllTasks[i].Names = names
		}
	}
}

func main() {
	if err := tasks.ExecAll(useDefault); err != nil {
		log.Fatal(err)
	}
}
