# gof

a file/folder processor written in Go  
用 Go 语言来写 extension 进行自由定制的文件/文件夹处理器。

带截图的说明： [screenshots.md](screenshots.md)

纯 Go 语言实现，扩展也是使用 Go 来写，通过添加扩展可对文件/文件夹进行随心所欲的操作，比如：

- 对调两个文件的文件名
- 把指定文件备份到指定文件夹，并自动改名
- 把指定文件移动到指定文件夹，并自动删除超过 n 天的旧文件
- 复制文件并且在复制结束后校验文件完整性
- 按你喜欢的方式单向/双向同步两个文件夹（具体就看扩展代码怎样写了）
- ……等等

有什么特殊的需求，都可以自己写 Go 代码来实现。

简而言之，就是本程序搭好了脚手架，处理好了一些通用的逻辑，让你可以专注于实现具体的文件操作代码。  
（比如，本程序可以用 YAML 配置文件来表达一连串操作，你只需要关心单个扩展的具体实现，而通过命令行与 YAML 两种方式输入指令、分析 YAML 文件、按顺序依次执行批量命令等都不需要操心，交由本程序的基础框架去处理）

## 安装方法

配置好 Go 语言环境后，执行以下命令：

```
$ go install github.com/ahui2016/gof@v0.2.1
```

如果有网络问题，请设置 goproxy：

```
$ go env -w GO111MODULE=on
$ go env -w GOPROXY=https://goproxy.cn,direct
```

## 使用方法

本仓库的源代码里提供了一个 examples 文件夹，下载到本地后在 examples 的各个子文件夹里执行以下命令可试验是否安装成功：

```
$ gof -f gof.yaml
```

可以在 yaml 文件里设定需要处理的文件，也可以用命令行指定，例如：

```
$ gof -f gof.yaml file1.txt file2.txt
```

**注意**: 通过命令行指定文件时，如果 YAML 文件里有多个任务，那么每个任务都会统一采用命令行指定的文件。命令行的优先级比 YAML 文件更高。如果试用 `-dump` 参数（详见后文 "任务计划"），可以看到命令行指定文件相当于设定了 global-names.

如果使用 yaml 文件，可以在文件里设定处理方式(recipe) 并且为每个任务设定不同的 options (上面的示例都使用了 yaml 文件)。

也可以不使用 yaml 文件，通过命令行来指定处理方式(recipe):

```
$ gof -r swap file1.txt file2.txt
```

使用 yaml 文件可依次执行多个任务，每个任务可分别设定不同的 options, 而使用参数 `-r` 指定 recipe 则每次只能执行一个任务，并且只能使用默认的 options。

### 任务计划

上面 "使用方法" 中的各种命令均可添加参数 `-dump`, 例如：

```
$ gof -f gof.yaml -dump
```

加了 `-dump` 的命令是安全的，不会真的执行，只会显示任务计划，并且会检查每个任务的参数是否正确。

**注意**: `-dump` 不可跟在被操作的文件名之后，比如下面是**错误**示范

```
$ gof -r swap file1.txt file2.txt -dump
```

上面的命令中 `-dump` 会被当作文件名，因此正确的命令应该是：

```
$ gof -dump -r swap file1.txt file2.txt
```

总之，如果通过命令行来指定被操作的文件，那么被指定的一个或多个文件名应该总是在命令的末尾。

### 帮助信息

- 为了让别人，以及未来一段时间之后的作者自己能迅速了解一个 recipe 的用途，建议每个 recipe 都认真实现 Help() 方法。

- 做法也很简单，大多数情况下直接黏贴一个 YAML 文件的内容并补充一些注释即可，具体请参考项目自带的 recipe (比如 swap.go, one-way-sync.go, move-new-files.go) 里的 Help() 方法。

- 在命令行，用 `gof -help -r swap` 即可查看关于 swap 的说明。

- 用 `gof -list` 可列出全部已经注册的 recipe。

### 一个技巧

使用 `-dump` 功能可非常方便地生成一个 YAML 文件，比如：

```
$ gof -dump -r swap file1.txt file2.txt > gof.yaml
```

即可生成一个 YAML 文件，这样，只需要对新生成的文件稍作修改就可以使用：

```
$ gof -f gof.yaml
```

## 关于 go install 和 GOBIN

如果设置了 GOBIN, 那么程序会被安装在 GOBIN 里，需要手动添加目录到环境变量中。
GOBIN 的具体位置可以用以下命令查看：

```
$ go env GOBIN
```

如果未设置 GOBIN, 请查看 go install 的帮助信息：

```
$ go help install
```

## 添加扩展

本程序采用了很容易添加扩展的设计，添加一个扩展的步骤如下：

1. fork 本仓库以方便修改
2. 在 recipes 文件夹里新建一个 `.go` 文件，第一行内容为 `package recipes`, 在该文件中定义一个 struct 并使其实现 Recipe 接口（参考 recipes 文件夹中已有的文件）
3. 在 main.go 里注册需要用到的 recipe
4. 不是必须，但建议在 examples 文件夹里添加用于测试的文件

完成。

最后，在你修改过的 gof 本地源码文件夹里，执行 `go install` 即可安装你自己定制版本的 gof

## 温馨提示

由于本程序涉及文件操作，实际使用前请先找一些无用文件来试验，确认没问题后再实际使用。建议初期不熟悉的时候多使用 `-dump` 参数（详见上面的 "任务计划" 部分）。

如果使用别人写的扩展，建议在试验前先检查源码。Go 语言很直白，这个检查通常是轻松的。
