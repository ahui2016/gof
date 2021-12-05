package recipes

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ahui2016/gof/util"
)

/*
 * 建议每个 Recipe 独立一个文件，并且其常量应添加前缀，函数应写成私有方法，
 * 因为全部 recipe 都在 package recipes 里面，要避免冲突。
 */

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
	verbose   bool
}

func (o *OneWaySync) Name() string {
	return "one-way-sync"
}

func (o *OneWaySync) Help() string {
	return `
- recipe: one-way-sync # 单向同步
  options:
    dry-run: "yes"     # 设为 yes 时只显示信息；设为 no 时才会实际执行
                       # 建议 dry-run 先设为 yes, 确认没有问题后再改成 no
    add: "yes"         # 是否添加文件
    update: "yes"      # 是否更新文件
    delete: "no"       # 是否删除文件
    by-date: "yes"     # 是否对比文件的修改日期
    by-content: "yes"  # 是否对比文件的内容
    verbose: "yes"     # 如果设为 no, 则在实际执行后不会显示详细信息
  names:
  - .\dest\            # 第一个是目标文件夹
  - .\folder\          # 从第二个开始是源头文件或文件夹
  - .\file.txt

# 注意，names 里的第一个元素是目标文件夹，可使用绝对目录或相对目录。
# 从 names 的第二个元素开始是源头文件或文件夹，只能使用相对目录。
# 使用本 recipe 时，必须先进入源头文件夹，在源头文件夹内执行 gof 命令。
`
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
		"verbose":    "yes",
	}
}

// Perpare 初始化一些项目，但 targetDir 与 srcFiles 则需要在 Validate 里初始化。
func (o *OneWaySync) Prepare(names []string, options map[string]string) {
	o.names = names
	o.dryRun = yesToBool(options["dry-run"])
	o.add = yesToBool(options["add"])
	o.update = yesToBool(options["update"])
	o.delete = yesToBool(options["delete"])
	o.byDate = yesToBool(options["by-date"])
	o.byContent = yesToBool(options["by-content"])
	o.verbose = yesToBool(options["verbose"])
}

func (o *OneWaySync) Validate() (err error) {
	// add/update/delete 至少其中一个必须设为 true
	if !o.add && !o.update && !o.delete {
		log.Println("warning: add/update/delete are all set to false, nothing will be sync'ed.")
	}
	// byDate/byContent 至少其中一个必须设为 true
	if !o.byDate && !o.byContent {
		return fmt.Errorf("by-date and by-content are all set to false, nothing to be compare")
	}

	o.names, err = namesLimit(o.names, 2, DefaultMax)
	if err != nil {
		return fmt.Errorf("%s: %w", o.Name(), err)
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

var (
	ows_addList    []string // 用于 dryRun 显示将要添加的内容
	ows_updateList []string // 用于 dryRun 显示将要更新的内容
	ows_delList    []string // 用于 dryRun 显示将要删除的内容
	ows_added      []string // 标记为已新增的文件夹（用来跳过处理这些文件夹的内容）
	ows_deleted    []string // 标记为已删除的文件夹（用来跳过处理这些文件夹的内容）
)

func (o *OneWaySync) Exec() error {
	// 处理 add 和 update
	for _, srcName := range o.srcFiles {
		if err := o.walk(srcName); err != nil {
			return err
		}
	}

	// 处理 delete
	if o.delete {
		if err := o.walkDelete(); err != nil {
			return err
		}
	}

	// 显示执行结果
	if o.dryRun {
		fmt.Println()
		fmt.Printf("**It's a dry run, not a real run.**\n")
	}
	if o.dryRun || o.verbose {
		fmt.Println()
		fmt.Printf("add (%v)\n", o.add)
		fmt.Println("----------------------")
		o.printArray(ows_addList)
		fmt.Println()
		fmt.Printf("update (%v)\n", o.update)
		fmt.Println("----------------------")
		o.printArray(ows_updateList)
		fmt.Println()
		fmt.Printf("delete (%v)\n", o.delete)
		fmt.Println("----------------------")
		o.printArray(ows_delList)
		fmt.Println()
	}
	return nil
}

func (o *OneWaySync) walkDelete() error {
	return filepath.WalkDir(o.targetDir, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Print("Error in WalkDir")
			return err
		}
		// 跳过 targetDir, 因为只对比 targetDir 的内容。
		if name == o.targetDir {
			return nil
		}
		// 跳过父文件夹已被删除的项目
		if o.subOfFolders(name, ows_deleted) {
			return nil
		}
		srcPath, err := filepath.Rel(o.targetDir, name)
		if err != nil {
			return err
		}
		notExist, err := util.PathIsNotExist(srcPath)
		if err != nil {
			return err
		}
		// 不存在于源头目录中的文件需要删除
		if notExist {
			// 标记需要删除的文件或文件夹
			ows_delList = append(ows_delList, name)
			// 把文件夹标记为已删除，以便跳过处理其内容
			if d.IsDir() {
				ows_deleted = append(ows_deleted, name)
			}
			// 实际删除文件或文件夹
			if !o.dryRun {
				return os.RemoveAll(name)
			}
		}
		return nil
	})
}

func (o *OneWaySync) walk(root string) error {
	return filepath.WalkDir(root, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Print("Error in WalkDir")
			return err
		}

		targetPath := filepath.Join(o.targetDir, name)
		notExists, err := util.PathIsNotExist(targetPath)
		if err != nil {
			return err
		}
		srcInfo, err := d.Info()
		if err != nil {
			return err
		}

		// 新增文件
		// 标记即将添加的文件
		if notExists {
			// 父文件夹未被标记为已添加的，才加进 ows_addList 中
			if !o.subOfFolders(targetPath, ows_added) {
				ows_addList = append(ows_addList, targetPath)
			}
			if d.IsDir() {
				ows_added = append(ows_added, targetPath)
			}
		}
		// 实际执行复制文件
		if !o.dryRun {
			// 如果是文件夹
			if d.IsDir() {
				if notExists {
					if err := os.Mkdir(targetPath, os.ModePerm); err != nil {
						return err
					}
				}
				return nil
			}
			// 如果是文件
			if notExists {
				if err := o.copy_setTime(targetPath, name, srcInfo); err != nil {
					return err
				}
			}
		}

		// 更新文件
		// 不需要对比文件夹，不需要对比不存在的文件，不需要对比新文件夹的内容
		if d.IsDir() || notExists || o.subOfFolders(targetPath, ows_added) {
			return nil
		}
		isNeedUpdate := false
		destInfo, err := os.Lstat(targetPath)
		if err != nil {
			return err
		}

		// 对比日期
		if o.byDate {
			if srcInfo.ModTime() != destInfo.ModTime() {
				isNeedUpdate = true
			}
		}
		// 对比内容
		if o.byContent {
			srcSum, err := util.FileSha256Hex(name)
			if err != nil {
				return err
			}
			destSum, err := util.FileSha256Hex(targetPath)
			if err != nil {
				return err
			}
			if srcSum != destSum {
				isNeedUpdate = true
			}
		}

		if isNeedUpdate {
			// 标记即将更新的文件
			ows_updateList = append(ows_updateList, targetPath)
			// 实际执行更新文件
			if !o.dryRun {
				if err := o.copy_setTime(targetPath, name, srcInfo); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// subOfFolders 检查 name 是否在 folders 里的任何一个文件夹中。
// folders 是被标记为已添加或已删除的文件夹。
func (o *OneWaySync) subOfFolders(name string, folders []string) bool {
	for _, folder := range folders {
		if strings.HasPrefix(name, folder) {
			return true
		}
	}
	return false
}

func (o *OneWaySync) printArray(arr []string) {
	if len(arr) == 0 {
		fmt.Println("(none)")
		return
	}
	for i := range arr {
		fmt.Println(arr[i])
	}
}

func (o *OneWaySync) copy_setTime(dest, src string, info fs.FileInfo) error {
	if err := util.CopyFile(dest, src); err != nil {
		return err
	}
	return os.Chtimes(dest, time.Now(), info.ModTime())
}
