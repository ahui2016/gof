# go-rename

自用的文件改名工具（比如两个文件的文件名对调）

暂时只有一个功能：对调两个文件的文件名。

## 安装方法

配置好 Go 语言环境后，执行以下命令：

```
$ go install github.com/ahui2016/go-rename@latest
```

本仓库的源代码里提供了一个 example 文件夹，下载到本地后在 example 文件夹里执行以下命令可试验是否安装成功：

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
