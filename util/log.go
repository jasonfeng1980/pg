package util

import (
    "fmt"
    "github.com/go-kit/kit/log"
    "github.com/jasonfeng1980/pg/conf"
    "github.com/jasonfeng1980/pg/ecode"
    "os"
    "strings"
    "sync"
    "time"
)

const DEFAULT_LOG_OS_STDOUT = "DEFAULT_LOG_OS_STDOUT"
var (
    logDir = ""
    logShowDebug = false
    serverName = ""
    defaultLog = log.NewLogfmtLogger(os.Stdout)
    logMask = logGetMask()
    TimesLayout = log.TimestampFormat(
        func() time.Time { return time.Now() },
        "2006-01-02 15:04:05",
    )
    logPools = &logPoll{
        pool: make(map[string]Logger),
    }
)
func init(){
    defaultLog = log.With(defaultLog, "ts", TimesLayout)
    //defaultLog = log.With(defaultLog, "caller", log.Caller(4))
    logPools.pool[DEFAULT_LOG_OS_STDOUT] = Logger{
       name: DEFAULT_LOG_OS_STDOUT,
       Logger: defaultLog,
    }
    logPools.osFile = append(logPools.osFile, os.Stdout)
}


func LogInit(conf *conf.Config) {
    logDir = conf.LogDir
    if logDir!="" &&  logDir[:1] != "/" {
        logDir = conf.ServerRoot + "/" + logDir
    }
    logShowDebug = conf.LogShowDebug
    serverName = conf.ServerName + conf.ServerNo
    logPools = &logPoll{
        pool: make(map[string]Logger),
    }

}

// 不做任何事的LOG
func LogNothing()log.Logger{
    return log.NewNopLogger()
}
// 获取指定名称的日志句柄
func LogHandle(name string) Logger{
    return logPools.Get(name)
}
// 关闭所有的日志链接
func LogClose(){
    logPools.Close()
}


type Logger struct {
    name string
    log.Logger
}

func (l Logger)Log(kvs ...interface{}) error{
    if logDir != "" { // 如果指定了存储文件
        if l.name == "DEBUG" && !logShowDebug { // 判断是否允许显示DEBUG日志
            return nil
        }
        newMask := logGetMask() // 获取当前的日期mask
        if newMask != logMask { // 如果mask不一致，就重新指向文件
            logPools.Reload(newMask)
            l.Logger = logPools.Get(l.name)
        }
    }
    return l.Logger.Log(kvs...)
}
func (l Logger)Logf(format string, args ...interface{}) error{
    newArg := fmt.Sprintf(format, args...)
    return l.Log("msg", newArg)
}

type logPoll struct {
    rwLock sync.RWMutex
    pool   map[string]Logger
    osFile  []*os.File
}
// 重新指向加载文件目录
func (p *logPoll) Reload(mask string) {
    logMask = mask // 重新指定mask
    if logDir == "" { // 没有指定存储文件，就不处理
        return
    }
    var logOsFile []*os.File
    // 循环pool，替换写入文件
    for n, _ := range p.pool {
        if n == DEFAULT_LOG_OS_STDOUT {
            continue
        }
        w := p.newLog(n)
        logOsFile = append(logOsFile, w)
    }
    // 关闭之前的文件句柄
    p.Close()

    // 创建新的文件关闭
    p.osFile = logOsFile
}
func (p *logPoll) Close(){
    for _, v := range p.osFile{
        v.Close()
    }
}
func (p *logPoll) Get(name string) Logger{
    name = strings.ToUpper(name)
    // 判断POOL里是否有
    p.rwLock.RLock()
    _, ok := p.pool[name]
    p.rwLock.RUnlock()
    if  !ok { // 没有就创建一个新的
        // 创建新的POOL
        w := p.newLog(name)
        if w != nil {
            p.osFile = append(p.osFile, w)
        }
    }

    return p.pool[name]
}

func (p *logPoll)newLog(name string) *os.File{
    var (
        logger log.Logger
        w *os.File
        err error
    )
    name = strings.ToUpper(name)

    p.rwLock.Lock()
    if logDir == "" || name == DEFAULT_LOG_OS_STDOUT { // 没有指定日志文件或者是OS_STDOUT
        logger = defaultLog
        w = nil
    } else {
        if !FilePathExists(logDir){ // 路径不正确
            panic(ecode.UtilWrongDir.Error(logDir))
        }
        file := fmt.Sprintf("%s/%s.%s.%s", logDir, serverName, name, logMask)
        w, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|0666)
        if err != nil { // 无法创建日志文件
            panic(err)
        }
        logger = log.NewJSONLogger(w)
        logger = log.With(logger, "ts", TimesLayout)
        logger = log.With(logger, "caller", log.Caller(4))
    }

    p.pool[name] = Logger{
        name: name,
        Logger: logger,
    }
    p.rwLock.Unlock()
    return w
}

func logGetMask() string{
    return time.Now().Format("060102")
}