package main

import (
	"flag"
	"fmt"
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

const gofVer = "v0.1.0"

const defaultConfigFileName = "gof.yaml"

var tasks model.Tasks

var (
	showVer = flag.Bool("v", false, "the version of gof")

	// YAML 文件名
	config = flag.String("f", "", "use a YAML config file")

	// -r 的优先级高于 -f (即，如果指定了 -r, 就忽略 -f)
	recipe = flag.String("r", "", "use a recipe with default options")

	dump = flag.Bool("dump", false, "do not run tasks, but print messages")

	// filenames, 优先级高于 YAML 文件里的 names
	names []string
)

func init() {
	util.Panic(initRecipes())
	initFlag()
}

func initFlag() {
	flag.Parse()
	names = flag.Args()

	// 如果有 "-v", 则显示 gof 的版本，并且忽略其它参数，不执行任何操作。
	if *showVer {
		return
	}

	// 如果命令行指定了 recipe 名称，则不需要 YAML 文件
	if *recipe != "" {
		v, ok := recipes.Get[*recipe]
		if !ok {
			log.Fatalf("not found recipe: %s", *recipe)
		}
		tasks = model.Tasks{AllTasks: []model.Task{{
			Recipe:  *recipe,
			Options: v.Default(),
			Names:   names,
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

	// 命令行输入的文件名的优先级比 tasks.Namse 更高。
	if len(names) > 0 {
		tasks.Names = names
	}

	// tasks.Names 的优先级比单个 task 里的 Names 更高。
	if len(tasks.Names) > 0 {
		for i := range tasks.AllTasks {
			tasks.AllTasks[i].Names = nil
		}
	}
}

func main() {
	// 如果有 "-v", 则显示 gof 的版本，并且忽略其它参数，不执行任何操作。
	if *showVer {
		fmt.Printf("gof %s\n", gofVer)
		fmt.Println("source code: https://github.com/ahui2016/gof")
		return
	}
	if *dump {
		util.Panic(printDump(tasks))
		return
	}
	if err := tasks.ExecAll(); err != nil {
		log.Fatal(err)
	}
}

func printDump(in interface{}) error {
	blob, err := yaml.Marshal(&in)
	if err != nil {
		return err
	}
	fmt.Print(string(blob))
	return nil
}
