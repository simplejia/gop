[中文 README](#中文)


# [gop](http://github.com/simplejia/gop) (go REPL)
## Original Intention
* Sometimes when we want to verify a go function quickly, coding in a file is too inefficient. While we have gop, opening a shell environment immediately, it will save the context automatically and enable you to import or export snippet at any time. In addition, it also can complete the code automatically and so on.

## Features
* history record: when gop is started, it will generate .gop folder under your home directory where inputting history is recorded.
* tab complete: when you tap tab, it can complete package and function which needs [gocode](http://github.com/nsf/gocode), if you have already installed gocode in your server but it can not work well, please run`go get -u github.com/nsf/gocode` to update it and install it again.
* It enables you to view code in real time and edit code [! command]
* snippet: import and export template [<,> command]

## Installation
> go get -u github.com/simplejia/gop

## Notice：
* When you input code, it supports continued line
* For the following code, the whole messages will output together when the executing is over
> print(1);time.Sleep(time.Second);print(2)

* You can input `echo 123`, echo is alias of println，You can redefine println variable to use your own print method, for example, defining println as the following: (utils.IprintD is characterized by printing the actual content of a pointer, even if the pointer is nested, fmt.printf can not do that)
```
import "github.com/simplejia/utils"
var println = utils.IprintD 
```
* When you import project package, you had better install package file to pkg directory in advance via go install which can accelerate the executing.
* You can import package in advance and atomically import it in subsequent use
* When gop is started, it will automatically import template code such as $PWD/gop.tmpl or $HOME/.gop/gop.tmpl, you can save your frequently-used code to gop.tmpl

## demo
```
$ gop
Welcome to the Go Partner! [[version: 1.7, created by simplejia]
Enter '?' for a list of commands.
GOP$ ?
Commands:
        ?|help  help menu
        -[dpc][#],[#]-[#],...   pop last/specific (declaration|package|code)
        ![!]    inspect source [with linenum]
        <tmpl   source tmpl
        >tmpl   write tmpl
        [#](...)        add def or code
        reset   reset
        list    tmpl list
        set|get set or get command-line argument
GOP$ for i:=1; i<3; i++ {
.....    print(i)
.....    time.Sleep(time.Millisecond)
.....}
1
2
GOP$ import _ "github.com/simplejia/wsp/demo/mysql"
GOP$ import _ "github.com/simplejia/wsp/demo/redis"
GOP$ import _ "github.com/simplejia/wsp/demo/conf"
GOP$ import "github.com/simplejia/lc"
GOP$ import "github.com/simplejia/wsp/demo/service"
GOP$ lc.Init(1024)
GOP$ demoService := service.NewDemo()
GOP$ demoService.Set("123", "456")
GOP$ time.Sleep(time.Millisecond)
GOP$ echo demoService.Get("123")
456
GOP$ >gop
GOP$ <gop
GOP$ !
        package main

p0:     import _ "github.com/simplejia/wsp/demo/mysql"
p1:     import _ "github.com/simplejia/wsp/demo/redis"
p2:     import _ "github.com/simplejia/wsp/demo/conf"
p3:     import "github.com/simplejia/lc"
p4:     import "github.com/simplejia/wsp/demo/service"
p5:     import "fmt" // imported and not used
p6:     import "strconv" // imported and not used
p7:     import "strings" // imported and not used
p8:     import "time" // imported and not used
p9:     import "encoding/json" // imported and not used
p10:    import "bytes" // imported and not used

        func main() {
c0:             lc.Init(1024)
c1:             demoService := service.NewDemo()
c2:             _ = demoService
c3:             demoService.Set("123", "456")
c4:             time.Sleep(time.Millisecond)
        }

GOP$
```

---
中文
===

# [gop](http://github.com/simplejia/gop) (go REPL)
## 实现初衷
* 有时想快速验证go某个函数的使用，临时写个程序太低效，有了gop，立马开一个shell环境，边写边运行，自动为你保存上下文，还可随时导入导出snippet，另外还有代码自动补全等等特性

## 特性
* history record（gop启动后会在home目录下生成.gop文件夹， 输入历史会记录在此）
* tab complete，可以补全package，补全库函数，需要系统安装有[gocode](http://github.com/nsf/gocode), 如果之前就安装过gocode，如果发现不能自动补全，请执行`go get -u github.com/nsf/gocode`升级重新安装下
* 代码实时查看和编辑功能[!命令功能]
* snippet，可以导入和导出模板[<,>命令功能]

## 安装
> go get -u github.com/simplejia/gop

## 注意：
* 输入代码时，支持续行
* 对于如下代码，只会在执行结束后一并输出
> print(1);time.Sleep(time.Second);print(2)

* 可以通过echo 123这种方式输出, echo是println的简写，你甚至可以重新定义println变量来使用自己的打印方法，比如像我这样定义(utils.IprintD的特点是可以打印出指针指向的实际内容，就算是嵌套的指针也可以，fmt.Printf做不到)：
```
import "github.com/simplejia/utils"
var println = utils.IprintD 
```
* 导入项目package时，最好提前通过go install方式安装包文件到pkg目录，这样可以加快执行速度
* 可以提前import包，后续使用时再自动引入
* gop启动后会自动导入$PWD/gop.tmpl或者$HOME/.gop/gop.tmpl模板代码，可以把常用的代码保存到gop.tmpl里

## demo
```
$ gop
Welcome to the Go Partner! [[version: 1.7, created by simplejia]
Enter '?' for a list of commands.
GOP$ ?
Commands:
        ?|help  help menu
        -[dpc][#],[#]-[#],...   pop last/specific (declaration|package|code)
        ![!]    inspect source [with linenum]
        <tmpl   source tmpl
        >tmpl   write tmpl
        [#](...)        add def or code
        reset   reset
        list    tmpl list
        set|get set or get command-line argument
GOP$ for i:=1; i<3; i++ {
.....    print(i)
.....    time.Sleep(time.Millisecond)
.....}
1
2
GOP$ import _ "github.com/simplejia/wsp/demo/mysql"
GOP$ import _ "github.com/simplejia/wsp/demo/redis"
GOP$ import _ "github.com/simplejia/wsp/demo/conf"
GOP$ import "github.com/simplejia/lc"
GOP$ import "github.com/simplejia/wsp/demo/service"
GOP$ lc.Init(1024)
GOP$ demoService := service.NewDemo()
GOP$ demoService.Set("123", "456")
GOP$ time.Sleep(time.Millisecond)
GOP$ echo demoService.Get("123")
456
GOP$ >gop
GOP$ <gop
GOP$ !
        package main

p0:     import _ "github.com/simplejia/wsp/demo/mysql"
p1:     import _ "github.com/simplejia/wsp/demo/redis"
p2:     import _ "github.com/simplejia/wsp/demo/conf"
p3:     import "github.com/simplejia/lc"
p4:     import "github.com/simplejia/wsp/demo/service"
p5:     import "fmt" // imported and not used
p6:     import "strconv" // imported and not used
p7:     import "strings" // imported and not used
p8:     import "time" // imported and not used
p9:     import "encoding/json" // imported and not used
p10:    import "bytes" // imported and not used

        func main() {
c0:             lc.Init(1024)
c1:             demoService := service.NewDemo()
c2:             _ = demoService
c3:             demoService.Set("123", "456")
c4:             time.Sleep(time.Millisecond)
        }

GOP$
```
