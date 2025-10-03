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

// isTBLibraryFile checks if a file belongs to the tb library
// This works for both local development and go mod cache scenarios
func isTBLibraryFile(filePath string) bool {
	// Check for local development path (when projectRoot is set)
	if projectRoot != "" && strings.HasPrefix(filePath, projectRoot) {
		return true
	}

	// Check for go mod cache path pattern
	// Pattern: /path/to/go/pkg/mod/github.com/xnumb/tb@version/
	if strings.Contains(filePath, "github.com/xnumb/tb@") {
		return true
	}

	// Alternative pattern for go mod cache (without @version)
	if strings.Contains(filePath, "github.com/xnumb/tb/") {
		// Additional check to ensure it's actually in a mod cache or vendor directory
		if strings.Contains(filePath, "/pkg/mod/") || strings.Contains(filePath, "/vendor/") {
			return true
		}
	}

	return false
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

		// Skip frames from the tb library itself
		// This covers both local development and go mod cache scenarios
		if isTBLibraryFile(frame.File) {
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
	if err != nil {
		ps = append([]any{err.Error(), nil}, ps...)
	}
	printLog(logType, ps...)
}

func Err(err error, ps ...any) {
	handleErr(err, "ERR", ps...)
}

func Fatal(err error, ps ...any) {
	handleErr(err, "FATAL", ps...)
	os.Exit(1)
}

func printLog(logType string, ps ...any) {
	t := utils.GetNow()

	var infos []string
	startIndex := 0

	// If there's an odd number of args, the first is a standalone message
	if len(ps)%2 == 1 {
		infos = append(infos, fmt.Sprintf("%+v", ps[0]))
		startIndex = 1
	}

	// Process the rest as key-value pairs
	for i := startIndex; i < len(ps); i += 2 {
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
