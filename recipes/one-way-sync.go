package recipes

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

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
	// add/update/delete 至少其中一个必须设为 true
	if !o.add && !o.update && !o.delete {
		log.Println("warning: add/update/delete are all set to false, nothing will be sync'ed.")
	}
	// byDate/byContent 至少其中一个必须设为 true
	if !o.byDate && !o.byContent {
		return fmt.Errorf("by-date and by-content are all set to false, nothing to be compare")
	}

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

var (
	ows_addList    []string
	ows_updateList []string
	ows_delList    []string
	ows_deleted    []string
)

func (o *OneWaySync) Exec() error {
	// 处理 add 和 update
	for _, srcName := range o.srcFiles {
		if err := ows_walk(srcName, o); err != nil {
			return err
		}
	}
	// 处理 delete
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
		if util.StrIndex(ows_deleted, name) >= 0 {
			return nil
		}
		if d.IsDir() {
			deleted, err := filepath.Glob(name + "*")
			if err != nil {
				return err
			}
			ows_deleted = append(ows_deleted, deleted...)
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
			if o.dryRun {
				ows_delList = append(ows_delList, name)
			} else {
				return os.RemoveAll(name)
			}
			fmt.Printf("Delete: %s\n", name)
		}
		return nil
	})
}

func ows_walk(srcRoot string, o *OneWaySync) error {
	return filepath.WalkDir(srcRoot, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Print("Error in WalkDir")
			return err
		}

		targetPath := filepath.Join(o.targetDir, name)
		notExists, err := util.PathIsNotExist(targetPath)
		if err != nil {
			return err
		}

		// 新增文件
		if o.dryRun {
			ows_addList = append(ows_addList, targetPath)
		} else {
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
				_, err := os.Create(targetPath)
				return err
			}
		}

		// 更新文件
		isNeedUpdate := false
		isTimeDiff := false
		isContentDiff := false

		// 对比日期
		if o.byDate {
			srcInfo, err := d.Info()
			if err != nil {
				return err
			}
			destInfo, err := os.Lstat(targetPath)
			if err != nil {
				return err
			}
			srcTime := srcInfo.ModTime().Format(time.RFC3339)
			destTime := destInfo.ModTime().Format(time.RFC3339)
			if srcTime != destTime {
				isTimeDiff = true
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
				isContentDiff = true
			}
		}

		if o.byDate && o.byContent {
			if isTimeDiff && isContentDiff {
				isNeedUpdate = true
			}
		} else {
			if isTimeDiff || isContentDiff {
				isNeedUpdate = true
			}
		}
		if isNeedUpdate {
			if o.dryRun {
				ows_updateList = append(ows_updateList, targetPath)
			} else {
				return util.CopyFile(targetPath, name)
			}
		}
		return nil
	})
}
