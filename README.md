# go-rename

自用的文件改名工具（比如两个文件的文件名对调）

暂时只有一个功能：对调两个文件的文件名。

## 安装方法

配置好 Go 语言环境后，执行以下命令：

```
$ go install github.com/ahui2016/go-rename@latest
```

本仓库的源代码里提供了一个 examples 文件夹，下载到本地后在 examples 的各个子文件夹里执行以下命令可试验是否安装成功：

```
$ go-rename -f go-rename.yaml
```

如果在当前文件夹里有一个 go-rename.yaml 文件, 则可以省略 `-f` 参数。

```
$ go-rename
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
