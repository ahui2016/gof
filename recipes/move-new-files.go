package recipes

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ahui2016/gof/util"
)

/*
 * 建议每个 Recipe 独立一个文件，并且其常量应添加前缀，函数应写成私有方法，
 * 因为全部 recipe 都在 package recipes 里面，要避免冲突。
 */

// MoveNewFiles 实现了 Recipe 接口，用于把一个文件夹内的 n 个最新文件移动到另一个文件夹。
// 只能处理一个文件夹内的第一层文件，不会递归搜索子文件夹。
type MoveNewFiles struct {
	names   []string // names[0] 是目标文件夹, names[1] 是源头文件夹
	options Options  // 暂存 options 待处理
	n       int      // 移动多少个修改日期最新的文件
	suffix  string   // 指定文件名的结尾，空字符串表示不限
	dryRun  bool     // 如果 dryRun 为 true, 则只显示信息，不实际执行
}

func (mv *MoveNewFiles) Name() string {
	return "move-new-files"
}

func (mv *MoveNewFiles) Help() string {
	return `
- recipe: move-new-files # 移动 n 个(修改日期)最新的文件
  options:
    n : 1          # 移动多少个文件
    suffix: ""     # 指定文件名的结尾，空字符串表示不限
    dry-run: "yes" # 设为 yes 时只显示信息；设为 no 时才会实际执行
  names:
  - .\dest\        # 第一个是目标文件夹
  - .\src\         # 第二个是源头文件夹

# 建议先 dry run, 如果有重名文件会提示 "skip"。
# 确认没问题后再把 dry run 的值改为 no。
`
}

func (mv *MoveNewFiles) Refresh() {
	*mv = *new(MoveNewFiles)
}

func (mv *MoveNewFiles) Default() Options {
	return Options{
		"n":       "1",
		"suffix":  "",
		"dry-run": "yes",
	}
}

// Perpare 初始化一些项目，但 mv.n 则需要在 Validate 里初始化。
func (mv *MoveNewFiles) Prepare(names []string, options Options) {
	mv.names = names
	mv.options = options
	mv.suffix = strings.ToLower(options["suffix"])
	mv.dryRun = yesToBool(mv.options["dry-run"])
}

func (mv *MoveNewFiles) Validate() error {
	n, err := strconv.Atoi(mv.options["n"])
	if err != nil {
		return err
	}
	if n < 1 {
		return fmt.Errorf("\"n\" should be 1 or larger")
	}
	mv.n = n

	mv.names, err = namesLimit(mv.names, 2, 2)
	if err != nil {
		return fmt.Errorf("%s: %w", mv.Name(), err)
	}
	return nil
}

func (mv *MoveNewFiles) Exec() error {
	infos, err := mv.getNewFiles()
	if err != nil {
		return err
	}
	if mv.dryRun {
		fmt.Printf("\n**It's a dry run, not a real run.**\n")
	}
	fmt.Printf("\nMove files from [%s] to [%s]\n", mv.names[1], mv.names[0])
	for _, info := range infos {
		target := filepath.Join(mv.names[0], info.Name())
		exists, err := util.PathIsExist(target)
		if err != nil {
			return err
		}
		if exists {
			fmt.Printf("-- skip %s\n", info.Name())
			continue
		} else {
			fmt.Printf("-- move %s\n", info.Name())
		}
		if mv.dryRun {
			continue
		}

		src := filepath.Join(mv.names[1], info.Name())
		if err := os.Rename(src, target); err != nil {
			if err := util.CopyFile(target, src); err != nil {
				return err
			}
			if err := os.Remove(src); err != nil {
				return err
			}
		}
	}
	return nil
}

func (mv *MoveNewFiles) getNewFiles() ([]fs.FileInfo, error) {
	files, err := ioutil.ReadDir(mv.names[1])
	if err != nil {
		return nil, err
	}
	// 只要普通文件，不要文件夹，如果指定了后缀，则只返回指定后缀的文件。
	files = mv.filter(files, func(info fs.FileInfo) bool {
		if !info.Mode().IsRegular() {
			return false
		}
		if mv.suffix == "" {
			return true
		}
		return strings.HasSuffix(strings.ToLower(info.Name()), mv.suffix)
	})
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	return files[:util.IntMin(mv.n, len(files))], nil
}

func (mv *MoveNewFiles) filter(arr []fs.FileInfo, test func(fs.FileInfo) bool) (result []fs.FileInfo) {
	for _, info := range arr {
		if test(info) {
			result = append(result, info)
		}
	}
	return
}
