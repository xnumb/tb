package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"tb/utils"
	"time"
)

func Info(ps ...any) {
	printLog(0, "INFO", ps...)
}

func Warn(ps ...any) {
	printLog(0, "WARN", ps...)
}

func Debug(ps ...any) {
	printLog(2, "DEBUG", ps...)
}

func handleErr(err error, skip int, logType string, ps ...any) {
	msg := "nil"
	if err != nil {
		msg = err.Error()
	}
	ps = append([]any{msg, nil}, ps...)
	printLog(skip, logType, ps...)
}

func Err(err error, ps ...any) {
	handleErr(err, 3, "ERR", ps...)
}

func ErrSkip(err error, skip int, ps ...any) {
	handleErr(err, skip, "ERR", ps...)
}

func Fatal(err error, ps ...any) {
	handleErr(err, 3, "FATAL", ps...)
	os.Exit(1)
}

func printLog(skip int, logType string, ps ...any) {
	t := utils.GetNow()
	if len(ps)%2 == 1 {
		ps = append(ps, nil)
	}
	var infos []string
	for i := 0; i < len(ps); i += 2 {
		key := ps[i]
		val := ps[i+1]
		if val == nil {
			infos = append(infos, fmt.Sprintf("%+v", key))
		} else {
			infos = append(infos, fmt.Sprintf("%v=%+v", key, val))
		}
	}
	path := ""
	if skip > 0 {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			fmt.Println("Can't get caller info.")
			return
		}
		// 尝试将绝对路径转换为相对路径
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, file); err == nil {
				file = rel
			}
		}
		path = fmt.Sprintf(" %s:%d", file, line)
	}
	fmt.Printf("[%s] %s %s%s\n", logType, t.Format(time.DateTime), strings.Join(infos, " "), path)
}
