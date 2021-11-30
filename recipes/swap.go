package recipes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahui2016/go-rename/util"
)

/*
 * 建议每个 Recipe 独立一个文件，并且其常量和函数都添加前缀，
 * 因为全部 Recipe 都在 package recipes 里面，要避免冲突。
 */

// swap_suffix 是用于临时文件名的后缀。
const swap_suffix = "1"

// swap_limit 限制最多可连续添加多少次 suffix，避免文件名无限变长。
const swap_limit = 20

// Swap 实现了 Recipe 接口
type Swap struct {
	names   []string
	verbose bool
}

func (s *Swap) Name() string {
	return "swap"
}

func (s *Swap) Prepare(names []string, options map[string]string) {
	s.names = names
	if options["verbose"] == "yes" {
		s.verbose = true
	}
}

func (s *Swap) Validate() error {
	if len(s.names) != 2 {
		return fmt.Errorf("%s: needs two filenames", s.Name())
	}
	s.names[0] = strings.TrimSpace(s.names[0])
	s.names[1] = strings.TrimSpace(s.names[1])
	if s.names[0] == "" || s.names[1] == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	return nil
}

func (s *Swap) Exec() error {
	if s.verbose {
		log.Printf("start to swap [%s] and [%s]", s.names[0], s.names[1])
	}
	temp, err := swap_tempName(s.names[0])
	if err != nil {
		return err
	}

	if s.verbose {
		log.Printf("found a safe temp filename: %s", temp)
		log.Printf("rename %s to %s", s.names[0], temp)
	}
	if err := os.Rename(s.names[0], temp); err != nil {
		return err
	}

	if s.verbose {
		log.Printf("rename %s to %s", s.names[1], s.names[0])
	}
	if err := os.Rename(s.names[1], s.names[0]); err != nil {
		return err
	}

	if s.verbose {
		log.Printf("rename %s to %s", temp, s.names[1])
	}
	if err := os.Rename(temp, s.names[1]); err != nil {
		return err
	}

	log.Printf("swap files OK: %s and %s", s.names[0], s.names[1])
	return nil
}

// swap_addSuffix 给一个文件名添加后缀，使其变成一个临时文件名。
// 比如 abc.js 处理后应变成 abc1.js
func swap_addSuffix(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name + swap_suffix
	}
	return name[:len(name)-len(ext)] + swap_suffix + ext
}

// swap_tempName 找出一个可用的临时文件名。
func swap_tempName(name string) (string, error) {
	for i := 0; i < swap_limit; i++ {
		name = swap_addSuffix(name)
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
