package recipes

import "fmt"

type Recipe interface {

	// the name of this recipe
	Name() string

	// 清空数据，确保每次被调用的 Recipe 都是全新的，未被污染的。
	Refresh()

	// Default 返回默认的 options
	Default() map[string]string

	// 在 Prepare 里进行一些初始化，为后续的 Validate 和 Exec 做准备。
	Prepare(names []string, options map[string]string)

	// 必须先执行 Prepare 然后才执行 Validate
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
