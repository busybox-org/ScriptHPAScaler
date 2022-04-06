# 插件开发指南
利用 go 包的 init 特性，将插件注册，并在主程序中调用。

# 添加插件
1. 创建插件目录，并在插件目录中添加插件的文件
2. 在plugins/README.md中添加插件的说明
3. 需要实现插件的 init 函数，并在 init 函数中注册插件。
4. 实现在 plugins 中 interface 的方法
5. 在 `plugins/standard/imports.go` 中添加插件的导入