package util

import (
    "testing"
)

func TestLogger_With(t *testing.T) {
    //Log.With("a", 111, "b", "ccc").Info()
    //
    //l := Log.Get("test")
    //pc,file,line,ok := runtime.Caller(0)
    //l.With("pc", pc, "file", file, "line", line, "ok", ok).Println("test", "222aa")
    //l.Log("a", 111, "b", 222)

}

func TestLog_New(t *testing.T) {
    //Log.Dir = "../log"
    Log.Debug = false

    l := Log.Get("test")
    l.Trace("trace")
    l.Log("log", 333, "tt", 5555)   // info
    l.With("a", 1333, "b", 666).Debug("debug")
    l.Warningln("warn")
    l.With("a", 1, "b", 22).Error("error")
}

