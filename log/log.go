package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xnumb/tb/emj"
	"github.com/xnumb/tb/utils"
)

var (
	projectRootOnce sync.Once
	projectRoot     string
)

var (
	consumerRootOnce sync.Once
	consumerRoot     string
)

func findProjectRoot() {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	// b is the path to this file (log.go)
	root := filepath.Dir(b)
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			projectRoot = root
			return
		}
		parent := filepath.Dir(root)
		if parent == root { // Reached root
			return
		}
		root = parent
	}
}

func findConsumerRoot(firstFrameFile string) {
	root := filepath.Dir(firstFrameFile)
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			consumerRoot = root
			return
		}
		parent := filepath.Dir(root)
		if parent == root { // Reached root
			consumerRoot = "" // Not found
			return
		}
		root = parent
	}
}

func getFilteredStackTrace() string {
	projectRootOnce.Do(findProjectRoot) // Ensure tb projectRoot is initialized

	var stackTrace []string
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	foundFirstFrame := false

	for {
		frame, more := frames.Next()
		if frame.File == "" {
			break
		}

		// Skip frames from the logging package itself
		if projectRoot != "" && strings.HasPrefix(frame.File, projectRoot) {
			continue
		}

		// Skip go runtime internal functions
		if strings.HasPrefix(frame.Function, "runtime.") {
			continue
		}

		// Skip main.go
		if filepath.Base(frame.File) == "main.go" {
			continue
		}

		// Find and cache the consumer's project root from the first valid frame
		if !foundFirstFrame {
			consumerRootOnce.Do(func() { findConsumerRoot(frame.File) })
			foundFirstFrame = true
		}

		filePath := frame.File
		// Make path relative to consumer's project root if possible
		if consumerRoot != "" {
			if rel, err := filepath.Rel(consumerRoot, frame.File); err == nil {
				filePath = rel
			}
		}

		stackTrace = append(stackTrace, fmt.Sprintf("\n\t%s:%d", filePath, frame.Line))

		if !more {
			break
		}
	}
	return strings.Join(stackTrace, "")
}

func Info(ps ...any) {
	printLog("INFO", ps...)
}

func Warn(ps ...any) {
	printLog("WARN", ps...)
}

func Debug(ps ...any) {
	printLog("DEBUG", ps...)
}

func handleErr(err error, logType string, ps ...any) {
	msg := "nil"
	if err != nil {
		msg = err.Error()
	}
	ps = append([]any{msg, nil}, ps...)
	printLog(logType, ps...)
}

func Err(err error, ps ...any) {
	handleErr(err, "ERR", ps...)
}

func ErrSkip(err error, skip int, ps ...any) {
	// Note: with stack traces, ErrSkip's skip is less meaningful, but we preserve the signature.
	// The filtering logic is now primary. We'll just call the base handler.
	handleErr(err, "ERR", ps...)
}

func Fatal(err error, ps ...any) {
	handleErr(err, "FATAL", ps...)
	os.Exit(1)
}

func printLog(logType string, ps ...any) {
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

	stack := ""
	if logType != "INFO" && logType != "WARN" {
		stack = getFilteredStackTrace()
	}
	logTypeText := ""
	switch logType {
	case "INFO":
		logTypeText = emj.Green
	case "WARN":
		logTypeText = emj.Yellow
	case "DEBUG":
		logTypeText = emj.Blue
	case "ERR":
		logTypeText = emj.Red
	case "FATAL":
		logTypeText = emj.Oh
	}
	logTypeText += "[" + logType + "]"

	fmt.Printf("%s %s %s%s\n", logTypeText, t.Format(time.DateTime), strings.Join(infos, " "), stack)
}
