package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyz"

func RandStr(n int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func ParseYaml(path string, obj any) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(f, obj)
	if err != nil {
		return err
	}
	return nil
}

func Md5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	bytes := hash.Sum(nil)
	md5Str := hex.EncodeToString(bytes)
	return md5Str
}

func Hex(s string) (string, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckVarcharLen(s string, maxLen int) bool {
	return utf8.RuneCountInString(s) > maxLen
}

func GetPrefixCmdVal(cmd, prefix string) string {
	if !strings.HasPrefix(cmd, prefix) {
		return ""
	}
	val := strings.TrimPrefix(cmd, prefix)
	val = strings.TrimPrefix(val, ":")
	val = strings.TrimPrefix(val, "：")
	val = strings.TrimSpace(val)
	return val
}

// ExecShell output, errMsg, error
func ExecShell(content string) (string, string, error) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "temp_script_*.sh")
	if err != nil {
		return "", "", err
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件在使用后被删除

	// 将 content 写入临时文件
	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", "", err
	}

	// 关闭临时文件
	err = tempFile.Close()
	if err != nil {
		return "", "", err
	}

	// 赋予临时文件执行权限
	err = os.Chmod(tempFile.Name(), 0755)
	if err != nil {
		return "", "", err
	}

	// 打开日志文件
	logFile, err := os.Create("command_output.log")
	if err != nil {
		return "", "", err
	}
	defer logFile.Close()

	cmd := exec.Command(tempFile.Name())
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(&out, logFile)
	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(&stderr, logFile)

	err = cmd.Run()
	if err != nil {
		return "", stderr.String(), err
	}
	return out.String(), "", nil
}
