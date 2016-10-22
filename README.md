# [gop](http://github.com/simplejia/gop) (go REPL)
## 实现初衷
* 有时想快速验证go某个函数的使用，临时写个程序太低效，有了gop，立马开一个shell环境，边写边运行，自动为你保存上下文，还可随时导入导出snippet，另外还有代码自动补全等等特性

## 特性
* history record（gop启动后会在home目录下生成.gop文件夹， 输入历史会记录在此）
* tab complete，可以补全package，补全库函数，需要系统安装有[gocode](http://github.com/nsf/gocode)
* r|w两种模式切换，r是默认模式，对用户输入实时解析运行，执行w命令切换到w模式，w模式下，只有当执行run命令时，代码才会真正执行
* 代码实时查看和编辑功能[!命令功能]
* snippet，可以导入和导出模板[<,>命令功能]

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
[r]$ ?
Commands:
        ?|help  help menu
        -[dpc][#],[#]-[#],...   pop last/specific (declaration|package|code)
        ![!]    inspect source [with linenum]
        <tmpl   source tmpl
        >tmpl   write tmpl
        [#](...)        add def or code
        run     run source
        compile compile source
        w       write source mode on
        r       write source mode off
        reset   reset
        list    tmpl list
[r]$ for i:=1; i<3; i++ {
.....    print(i)
.....    time.Sleep(time.Millisecond)
.....}
1
2
[r]$ import _ "github.com/simplejia/wsp/demo/mysql"
[r]$ import _ "github.com/simplejia/wsp/demo/redis"
[r]$ import _ "github.com/simplejia/wsp/demo/conf"
[r]$ import "github.com/simplejia/lc"
[r]$ import "github.com/simplejia/wsp/demo/service"
[r]$ lc.Init(1024)
[r]$ demoService := service.NewDemo()
[r]$ demoService.Set("123", "456")
[r]$ time.Sleep(time.Millisecond)
[r]$ echo demoService.Get("123")
456
[r]$ >gop
[r]$ <gop
[r]$ !
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

[r]$
```

## LICENSE
gop is licensed under the Apache Licence, Version 2.0
(http://www.apache.org/licenses/LICENSE-2.0.html)
