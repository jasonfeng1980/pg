package util

import (
    "time"
)

var (
    TimeLocation = time.FixedZone("CST", 8*3600)  // 东八
    TimeFormatOption = "2006-01-02 15:04:05"
)

// 当前本地时间
func TimeNow() time.Time{
    return time.Now().In(TimeLocation)
}
// 当前本地时间字符串
func TimeNowString() string{
    return time.Now().In(TimeLocation).Format(TimeFormatOption)
}

// 变成字符串
func TimeFormat(tim time.Time)string{
    return tim.Format(TimeFormatOption)
}

// 几天之后，负数为几天之前
func TimeAddDate(beg time.Time, years int, months int, days int) time.Time{
    return beg.AddDate(years, months, days)
}

// 距离多少时间单位
func TimeAdd(beg time.Time, d time.Duration)time.Time{
    return beg.Add(d)
}