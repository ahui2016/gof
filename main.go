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
		new(recipes.OneWaySync),
		new(recipes.MoveNewFiles),
	// 在这里添加 recipe
	)
}

const gofVer = "v0.2.1"

var tasks model.Tasks

var (
	showVer = flag.Bool("v", false, "the version of gof")

	// YAML 文件名
	config = flag.String("f", "", "use a YAML config file")

	// -r 的优先级高于 -f (即，如果指定了 -r, 就忽略 -f)
	recipe = flag.String("r", "", "use a recipe with default options")
	help   = flag.Bool("help", false, "print a brief overview of a recipe")
	list   = flag.Bool("list", false, "print out all registered recipes")

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

	// 如果有 "-v" 或 "-list" 或 "-help", 则显示相关信息，并且忽略其它参数，不执行任何操作。
	if *showVer || *list {
		return
	}

	// 如果命令行指定了 recipe 名称，则不需要 YAML 文件
	if *recipe != "" {
		v := getRecipe(*recipe)
		tasks = model.Tasks{AllTasks: []model.Task{{
			Recipe:  *recipe,
			Options: v.Default(),
			Names:   names,
		}}}
	} else {
		// 如果命令行未指定 recipe, 则需要一个 YAML 文件，
		if strings.TrimSpace(*config) == "" {
			log.Fatalf("\nUsage Example:\n    gof -f example.yaml\n    gof -r swap file1 file2")
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
	// 如果有 "-v" 或 "-list" 或 "-help", 则显示相关信息，并且忽略其它参数，不执行任何操作。
	if *showVer {
		fmt.Printf("gof %s\n", gofVer)
		fmt.Println("source code: https://github.com/ahui2016/gof")
		return
	}
	if *list {
		recipesNames := []string{}
		for k := range recipes.Get {
			recipesNames = append(recipesNames, k)
		}
		fmt.Print("registered recipes: ")
		fmt.Print(strings.Join(recipesNames, ", "))
		fmt.Println()
		return
	}
	if *help {
		if *recipe == "" {
			fmt.Println("-help: print a brief overview of a recipe")
			fmt.Println("use -r to specify a recipe, for example: gof -help -r swap")
			fmt.Println("use -list to list out all registered recipes")
		} else {
			v := getRecipe(*recipe)
			fmt.Println(v.Help())
		}
		return
	}

	if *dump {
		util.Panic(printDump(tasks))
	}
	if err := tasks.ExecAll(!*dump); err != nil {
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

func getRecipe(name string) recipes.Recipe {
	v, ok := recipes.Get[*recipe]
	if !ok {
		log.Fatalf("not found recipe: %s\nuse -list to list out all registered recipes", *recipe)
	}
	return v
}
