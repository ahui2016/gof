package recipes

type Recipe interface {
	// the name of this recipe
	Name() string

	// 在 Prepare 里进行一些初始化，为后续的 Validate 和 Exec 做准备。
	Prepare(names []string, options map[string]string)

	// 必须先执行 Prepare 然后才执行 Validate
	Validate() error

	// 必须先执行 Validate 然后才执行 Exec
	Exec() error
}

var Recipes = map[string]Recipe{
	"swap": new(Swap),
}
