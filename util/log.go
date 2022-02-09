package util

import (
    "github.com/natefinch/lumberjack"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
    "strings"
    "time"
)

var l *Logger
var ln *Logger

func Log() *Logger {
    return ln
}

// 初始化
func init() {
    LogInit("", "", "debug")
}

type Logger struct {
    *zap.Logger
    S *zap.SugaredLogger
}

var zapLevel zapcore.Level

func LogInit(filename string, serverName string, level string) {
    var (
        writerInfo zapcore.WriteSyncer
        writerErr  zapcore.WriteSyncer
    )
    // 获取zapLevel
    switch strings.ToLower(level) {
    case "error":
        zapLevel = zapcore.ErrorLevel
    case "warn":
        zapLevel = zapcore.WarnLevel
    case "info":
        zapLevel = zapcore.InfoLevel
    case "debug":
        zapLevel = zapcore.DebugLevel
    default:
        zapLevel = zapcore.DebugLevel
    }
    // 获取writer
    if filename != "" {
        writerInfo = zapcore.AddSync(&lumberjack.Logger{
            Filename:   filename + "/" + serverName + ".access",
            MaxSize:    200,
            MaxBackups: 30,
            MaxAge:     30,
            Compress:   false,
            LocalTime:  true,
        })
        writerErr = zapcore.AddSync(&lumberjack.Logger{
            Filename:   filename + "/" + serverName +".error",
            MaxSize:    200,
            MaxBackups: 30,
            MaxAge:     30,
            Compress:   false,
            LocalTime:  true,
        })
    } else {
        writerInfo = os.Stdout
        writerErr = os.Stdout
    }
    // 获取encoder
    encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
        TimeKey:     "T",
        LevelKey:    "L",
        CallerKey:   "C",
        MessageKey:  "M",
        LineEnding:  zapcore.DefaultLineEnding,
        EncodeLevel: zapcore.LowercaseLevelEncoder,
        EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
            enc.AppendString(t.Format("2006-01-02 15:04:05"))
        },
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    })

    // 区分正常日志 和错误日志
    levelInfo := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl >= zapLevel && lvl < zapcore.ErrorLevel
    })
    levelErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl >= zapLevel && lvl >= zapcore.ErrorLevel
    })

    //core := zapcore.NewCore(encoder, writer, zapLevel)
    core := zapcore.NewTee(
        zapcore.NewCore(encoder, writerInfo, levelInfo),
        zapcore.NewCore(encoder, writerErr, levelErr),
    )

    zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
    zapLogLn := zapLog.WithOptions(zap.AddCallerSkip(1))
    l = &Logger{
        zapLog,
        zapLog.Sugar(),
    }
    ln = &Logger{
        zapLogLn,
        zapLogLn.Sugar(),
    }
}

func makeZapField(kvs ...interface{}) []zap.Field{
    len := len(kvs)
    var ret []zap.Field
    for i := 1; i < len; i = i + 2 {
        ret = append(ret, zap.String(Str(kvs[i-1]), Str(kvs[i])))
    }
    return ret
}

//////////////////////////////////
// 以下是快捷方式
//////////////////////////////////


func (* Logger) L(msgs ...interface{}) {
    Logs(msgs...)
}
func Logs(msgs ...interface{}) {
    strList := make([]string, len(msgs))
    for _, m := range msgs {
        strList = append(strList, Str(m))
    }
    ln.Debug(strings.Join(strList, "  "))
}
func (* Logger) D(msg interface{}, kvs ...interface{}) {
    Debug(msg, kvs...)
}
func Debug(msg interface{}, kvs ...interface{}) {
    ln.Debug(Str(msg), makeZapField(kvs...)...)
}

func (* Logger) I(msg interface{}, kvs ...interface{}) {
    Info(msg, kvs...)
}
func Info(msg interface{}, kvs ...interface{}) {
    ln.Info(Str(msg), makeZapField(kvs...)...)
}

func (* Logger) W(msg interface{}, kvs ...interface{}) {
    Warn(msg, kvs...)
}
func Warn(msg interface{}, kvs ...interface{}) {
    ln.Warn(Str(msg), makeZapField(kvs...)...)
}

func (* Logger) E(msg interface{}, kvs ...interface{}) {
    Error(msg, kvs...)
}
func Error(msg interface{}, kvs ...interface{}) {
    ln.Error(Str(msg), makeZapField(kvs...)...)
}

func (* Logger) P(msg interface{}, kvs ...interface{}) {
    Panic(msg, kvs...)
}
func Panic(msg interface{}, kvs ...interface{}) {
    ln.Panic(Str(msg), makeZapField(kvs...)...)
}
