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
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }

        cmdString = strings.TrimSuffix(cmdString, "\n")
        tip, goNext = f(cmdString)
        if !goNext {
            break
        }
    }
}

