package tb

import (
	"errors"
	"strings"

	"github.com/xnumb/tb/log"
	"github.com/xnumb/tb/to"
)

var (
	ErrCmdMaxLen      = errors.New("命令字符串长度不能超过64个字符")
	ErrCmdInvalidPage = errors.New("无法解析cmd字符串中的page为int类型")
)

const (
	CmdEmpty               = "-"
	CmdDel                 = "_del"
	cmdCancelAsk           = "_cancelAsk"
	cmdPrefixCancelConfirm = "_cancelConfirm."
	cmdPrefixConfirmed     = "_confirmed."
)

// genCmd cmd格式为[id]>[page]:[Arg1],[Arg2]...
// 例如: _del | users>2:1,2 | user:2
func genCmd(id string, page int, args ...string) string {
	s := id
	if page > 0 {
		s += ">" + to.S(page)
	}
	if len(args) == 0 {
		return s
	} else {
		s = s + ":" + strings.Join(args, ",")
		if len(s) > 64 {
			log.Err(ErrCmdMaxLen, "cmd", s)
		}
		return s
	}
}

func parseCmd(s string) (string, int, []string) {
	id := ""
	page := 0
	var args []string
	arr := strings.Split(s, ":")
	cmd := arr[0]
	arr2 := strings.Split(cmd, ">")
	id = arr2[0]
	if len(arr2) > 1 {
		p, ok := to.I(arr2[1])
		if !ok {
			log.Err(ErrCmdInvalidPage, "cmd", s)
		}
		page = p
	}
	if len(arr) > 1 {
		args = strings.Split(arr[1], ",")
	}
	return id, page, args
}
