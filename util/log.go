package util

import (
    "github.com/lestrrat-go/file-rotatelogs"
    "github.com/rifflock/lfshook"
    "github.com/sirupsen/logrus"
    "runtime"
    "sync"
)

var (
    Log     = log{
        Dir:"",
        ShowDebug: true,
        Logger: &Logger{
            "",
            logrus.New(),
        },
    }
    logPool = &logPools{
        Pools: make(map[string]*Logger),
    }
)
func init(){
    Log.Level = logrus.DebugLevel
}

func LogInit(dir string, showDebug bool, serverName string, serverNo string) {
    Log = log{
        Dir: dir,
        ShowDebug: showDebug,
        PreName: serverName + serverNo,
    }
    Log.Logger= Log.Get("")
    //
    //Log = &log{
    //    Dir: "",
    //    ShowDebug: true,
    //    Logger: logDefault,
    //    PreName: serverName + serverNo,
    //}
}

type log struct {
    Dir     string
    ShowDebug   bool
    PreName string
    *Logger
}
// 创建新的日志文件
func (l *log)New(name string) (*Logger, error){
    newLog := &Logger{
        name,
        logrus.New(),
    }

    // 测试模式 显示行号
    //newLog.ReportCaller = false
    // 非测试模式， 显示error及以上的错误
    if l.ShowDebug {
        newLog.Level = logrus.TraceLevel
    } else {
        newLog.Level = logrus.TraceLevel
    }
    //logFormat := &logrus.JSONFormatter{
    //    PrettyPrint:     l.ShowDebug,         // 格式化
    //    TimestampFormat: "06-01-02 15:04:05", // 时间格式
    //    DisableHTMLEscape: true,              // 不转义HTML特殊字符
    //}
    logFormat := &logrus.TextFormatter{
        FullTimestamp: true,
        TimestampFormat: "2006-01-02 15:04:05", // 时间格式
    }
    if l.Dir != ""  {
        newLog.Out = &nullWrite{}
        newLog.AddHook(l.newLfsHook(name, logFormat))
    } else {
        newLog.Formatter = logFormat
    }
    return newLog, nil
}
func (l *log)Get(name string) *Logger{
    return logPool.Get( name)
}

type nullWrite struct {}
func (nu *nullWrite)Write(p []byte) (n int, err error){
    return 0, nil
}

// 日志钩子(日志拦截，并重定向)
func (l *log)newLfsHook(name string, logFormat logrus.Formatter) logrus.Hook {
    writerAccess := l.logWriter(name,"access")
    writerDebug := l.logWriter(name,"debug")
    writerError := l.logWriter(name,"error")

    // 可设置按不同level创建不同的文件名
    lfsHook := lfshook.NewHook(lfshook.WriterMap{
        logrus.InfoLevel:  writerAccess,
        logrus.TraceLevel: writerDebug,
        logrus.DebugLevel: writerDebug,
        logrus.WarnLevel:  writerError,
        logrus.ErrorLevel: writerError,
        logrus.FatalLevel: writerError,
        logrus.PanicLevel: writerError,
    }, logFormat)

    return lfsHook
}
func (l *log)logWriter(name, env string) *rotatelogs.RotateLogs {
    if name != "" {
        env = name
    }
    writer, err := rotatelogs.New(
        // 日志文件
        l.Dir + "/" + l.PreName + "." + env  + ".%y%m%d",

        // 日志周期(默认每86400秒/一天旋转一次)
        rotatelogs.WithRotationTime(86400),

        // 清除历史 (WithMaxAge和WithRotationCount只能选其一)
        //rotatelogs.WithMaxAge(time.Hour*24*7), //默认每7天清除下日志文件
        rotatelogs.WithRotationCount(30), //只保留最近的N个日志文件
    )
    if err != nil {
        panic(err)
    }
    return writer
}


type Logger struct {
    name string
    *logrus.Logger
}
func (l *Logger)ShowLine(skip int) *logrus.Entry{
    _, file, line, _ := runtime.Caller(skip)
    return l.With("LOG_FILE", file + ":" + StrParse(line))
}
func (l *Logger)LogPretty(v interface{}, callerSkip int) {
    if Log.Dir != "" {
        l.ShowLine(callerSkip).Debugln(v)
        return
    }
    s, err := JsonIndent(v)
    if err != nil {
        l.Debugln(err)
    }
    l.ShowLine(callerSkip).Debugln(s)
}

func (l *Logger)Log(kvs ...interface{}) error{
    l.With(kvs...).Info()
    return nil
}

func (l *Logger)With(kvs ...interface{}) *logrus.Entry {
    var ret = &logrus.Entry{
        Logger: l.Logger,
    }
    len := len(kvs)
    for i:=1; i<len; i=i+2 {
        ret = ret.WithField(StrParse(kvs[i-1]), kvs[i])
    }
    return ret
}

type logPools struct {
    rwLock sync.RWMutex
    Pools map[string]*Logger
}
func (p *logPools)makeKey(name string) string{
    return Log.PreName + "." + name
}
func (p *logPools)Add(l *Logger){
    p.rwLock.Lock()
    key := p.makeKey(l.name)
    p.Pools[key] = l
    p.rwLock.Unlock()
}
func (p *logPools)Get(name string) *Logger {
    key := p.makeKey(name)
    p.rwLock.RLock()
    ret, ok := p.Pools[key]
    p.rwLock.RUnlock()
    if !ok{
        var err error
        if ret, err = Log.New(name); err ==nil {
            p.Add(ret)
        }
    }
    return ret
}

func LogNothing() logNothing{
    return logNothing{}
}
type logNothing struct {}
func (logNothing) Log(...interface{}) error { return nil}