package recipes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahui2016/gof/util"
)

/*
 * 建议每个 Recipe 独立一个文件，并且其常量应添加前缀，函数应写成私有方法，
 * 因为全部 recipe 都在 package recipes 里面，要避免冲突。
 */

// swap_suffix 是用于临时文件名的后缀。
const swap_suffix = "1"

// swap_limit 限制最多可连续添加多少次 suffix，避免文件名无限变长。
const swap_limit = 20

// Swap 实现了 Recipe 接口，用于对调两个文件名（或文件夹名）。
// Swap 只能用于不需要移动文件的情况，比如同一个文件夹（或同一个硬盘分区）内的文件可以操作，
// 而跨硬盘分区的文件则无法处理。
type Swap struct {
	names   []string
	verbose bool
}

func (s *Swap) Name() string {
	return "swap"
}

func (s *Swap) Help() string {
	return `
- recipe: swap    # 对调两个文件名（或文件夹名）
  options:
    verbose: yes  # yes/no, 显示或不显示程序执行的详细过程
  names:          # 文件名（或文件夹名），数量必须是两个，不能多不能少
  - file1.txt
  - file2.txt

# Swap 只能用于不需要移动文件的情况，比如同一个文件夹 (或同一个硬盘分区)
# 内的文件可以操作，而跨硬盘分区的文件则无法处理。
`
}

func (s *Swap) Refresh() {
	*s = *new(Swap)
}

func (s *Swap) Default() Options {
	return Options{
		"verbose": "yes",
	}
}

func (s *Swap) Prepare(names []string, options Options) {
	s.names = names
	s.verbose = yesToBool(options["verbose"])
}

func (s *Swap) Validate() (err error) {
	s.names, err = namesLimit(s.names, 2, 2)
	if err != nil {
		return fmt.Errorf("%s: %w", s.Name(), err)
	}
	for i := range s.names {
		if err := util.FindFile(s.names[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Swap) Exec() error {
	if s.verbose {
		log.Printf("start to swap [%s] and [%s]", s.names[0], s.names[1])
	}
	temp, err := s.tempName(s.names[0])
	if err != nil {
		return err
	}

	if s.verbose {
		log.Printf("-- found a safe temp filename: %s", temp)
		log.Printf("-- rename %s to %s", s.names[0], temp)
	}
	if err := os.Rename(s.names[0], temp); err != nil {
		return err
	}

	if s.verbose {
		log.Printf("-- rename %s to %s", s.names[1], s.names[0])
	}
	if err := os.Rename(s.names[1], s.names[0]); err != nil {
		return err
	}

	if s.verbose {
		log.Printf("-- rename %s to %s", temp, s.names[1])
	}
	if err := os.Rename(temp, s.names[1]); err != nil {
		return err
	}

	log.Printf("swap files OK: %s and %s", s.names[0], s.names[1])
	return nil
}

// addSuffix 给一个文件名添加后缀，使其变成一个临时文件名。
// 比如 abc.js 处理后应变成 abc1.js
func (s *Swap) addSuffix(name string) string {
	clean := filepath.Clean(name)

	// 去掉最后一个分隔符，当作文件来处理。
	// 注意，此时 name 也许是个文件夹，但没关系，在这里可以当作文件来处理。
	if strings.HasSuffix(clean, string(filepath.Separator)) {
		clean = clean[:len(clean)-1]
	}

	dir, base := filepath.Split(clean)
	ext := filepath.Ext(name)
	temp := base[:len(base)-len(ext)] + swap_suffix + ext
	return filepath.Join(dir, temp)
}

// tempName 找出一个可用的临时文件名。
func (s *Swap) tempName(name string) (string, error) {
	for i := 0; i < swap_limit; i++ {
		name = s.addSuffix(name)
		ok, err := util.PathIsNotExist(name)
		if err != nil {
			return "", err
		}
		if ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("no proper temp filename, last try: %s", name)
}
