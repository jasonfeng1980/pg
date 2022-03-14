package util

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type ConsoleTipFunc func(string) (nextTip string, goNext bool)
// 提示->等待用户输入->然后运行
func ConsoleTip(tip string, f ConsoleTipFunc){
    var goNext bool
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print(tip, ": ")
        fmt.Print("$ ")
        cmdString, err := reader.ReadString('\n')
        l := len(cmdString)
        if cmdString[l-2:l] == "\r\n" { // windows 系统 需要去掉\r\n
            cmdString = cmdString[0:l-2]
        }
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }

        cmdString = strings.TrimSuffix(cmdString, "\n")
        // 如果输入的是exit 或者 quit 就退出
        if ListHave([]string{"exit", "quit"}, strings.ToLower(cmdString)) {
            break
        }
        tip, goNext = f(cmdString)
        if !goNext {
            break
        }
    }
}

