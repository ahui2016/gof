package recipes

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ahui2016/gof/util"
)

// OneWaySync 实现了 Recipe 接口，用于单向同步。
// 把 srcFiles (包括文件和文件夹) 同步到 distFolder,
// distFolder 里没有的文件就 add, 已有的就对比差异按需 update, 多余的则 delete,
// 其中 add, update, delete 都可以单独控制, true 才执行, false 则不执行（至少一项为 true）。
// 对于 update 的情况，可选择是否对比日期、是否对比内容（至少对比其中一项）。
// 如果 dryRun 为 true, 则只显示信息，不实际执行。
type OneWaySync struct {
	names     []string
	targetDir string
	srcFiles  []string
	dryRun    bool
	add       bool
	update    bool
	delete    bool
	byDate    bool
	byContent bool
}

func (o *OneWaySync) Name() string {
	return "one-way-sync"
}

func (o *OneWaySync) Refresh() {
	*o = *new(OneWaySync)
}

func (o *OneWaySync) Default() map[string]string {
	return map[string]string{
		"dry-run":    "yes",
		"add":        "yes",
		"update":     "yes",
		"delete":     "no",
		"by-date":    "no",
		"by-content": "yes",
	}
}

// Perpare 初始化一些项目，但 targetDir 与 srcFiles 则需要在 Validate 里初始化。
func (o *OneWaySync) Prepare(names []string, options map[string]string) {
	o.names = names
	o.dryRun = options["dry-run"] == "yes"
	o.add = options["add"] == "yes"
	o.update = options["update"] == "yes"
	o.delete = options["delete"] == "yes"
	o.byDate = options["by-date"] == "yes"
	o.byContent = options["by-content"] == "yes"
}

func (o *OneWaySync) Validate() error {
	// 清除空字符串
	o.names = util.StrSliceFilter(o.names, func(name string) bool {
		return name != ""
	})

	// 检查 o.names 的数量
	if len(o.names) < 2 {
		log.Print("file/folder names: ", o.names)
		return fmt.Errorf("%s: needs at least two file/folder names", o.Name())
	}

	// 确保每个文件/文件名都真实存在
	for i := range o.names {
		if err := util.FindFile(o.names[i]); err != nil {
			return err
		}
	}

	// 确保 o.names[0] 是文件夹
	info, err := os.Lstat(o.names[0])
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("the target is not a folder: %s", o.names[0])
	}

	// 初始化
	o.targetDir = o.names[0]
	o.srcFiles = o.names[1:]
	return nil
}

func (o *OneWaySync) Exec() error {
	if runtime.GOOS == "windows" {
		for _, srcName := range o.srcFiles {
			if err := walk(srcName); err != nil {
				return err
			}
		}
	} else {
		for _, srcName := range o.srcFiles {
			info, err := os.Stat(o.srcFiles[0])
			if err != nil {
				return err
			}
			if info.IsDir() {
				if err := walk(srcName); err != nil {
					return err
				}
			} else {
				fmt.Printf("visited file: %q\n", srcName)
			}
		}
	}
	return nil
}

func walk(root string) error {
	return filepath.WalkDir(root, func(name string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Print("Error in WalkDir")
			return err
		}
		if info.IsDir() {
			fmt.Printf("visited dir: %q\n", name)
		} else {
			fmt.Printf("visited file: %q\n", name)
		}
		return nil
	})
}
