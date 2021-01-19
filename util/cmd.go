package util

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type CmdFunc func(string) (nextTip string, goNext bool)

func CmdWait(tip string, f CmdFunc){
    var goNext bool
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Println(tip)
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
