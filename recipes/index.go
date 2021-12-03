package recipes

import "fmt"

type Recipe interface {

	// Name of this recipe.
	// 注意，应返回一个便于命令行输入的名字，比如中间不要有空格。
	Name() string

	// Help messages.
	// 方便在命令行查看每个 recipe 的用途。如果没有写清楚，使用者（包括一段时间之后的作者自己）
	// 就需要查看源文件才能知道具体使用方法了。通常用一个带注释的 YAML 文件即可，例子可参考项目自带的 recipe.
	Help() string

	// 清空数据，确保每次被调用的 Recipe 都是全新的，未被污染的。
	Refresh()

	// Default 返回默认的 options
	Default() map[string]string

	// 在 Prepare 里进行一些初始化，为后续的 Validate 和 Exec 做准备。
	// 但由于有些参数需要检查后才能初始化（避免 panic），因此一部分初始化要放在 Validate 里实施。
	Prepare(names []string, options map[string]string)

	// 必须先执行 Prepare 然后才执行 Validate.
	// 注意: 在 Validate 只能读取文件信息，不可修改文件，包括文件内容、日期、权限等等任何修改都不允许。
	// 必须保证 Validate 是安全的，不会对文件进行任何修改的。
	Validate() error

	// 必须先执行 Validate 然后才执行 Exec
	Exec() error
}

var Get = make(map[string]Recipe)

func Register(recipes ...Recipe) error {
	for _, recipe := range recipes {
		_, ok := Get[recipe.Name()]
		if ok {
			return fmt.Errorf("%s already exists", recipe.Name())
		}
		Get[recipe.Name()] = recipe
	}
	return nil
}
