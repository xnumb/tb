package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/xnumb/tb/emj"
	tele "gopkg.in/telebot.v4"
)

var ErrFormatError = errors.New("解析格式错误")

func CheckBotTokenFmt(s string) bool {
	bt := strings.Split(s, ":")
	if len(bt) != 2 {
		return false
	}
	_, err := strconv.ParseInt(bt[0], 10, 64)
	if err != nil {
		return false
	}
	return len(bt[1]) == 35
}

const BtnsFormatIntro = "请按照格式输入按钮配置命令 如: \n" +
	"<code>按钮1 https://www.google1.com 按钮2 https://www.google2.com 按钮3 https://www.google3.com\n" +
	"按钮4 https://www.google4.com 按钮5 https://www.google5.com\n" +
	"按钮6 https://www.google6.com</code>\n" +
	emj.Warn + " 生成的按钮与命令格式的顺序与换行一致\n" +
	emj.Warn + " 链接与按钮之间需要一个空格\n" +
	emj.Warn + " 按钮名称请不要插入超过一个空格"

func GenUrlBtns(s string) ([]tele.Row, error) {
	if !strings.Contains(s, "https://") {
		return nil, ErrFormatError
	}
	var rows []tele.Row
	rs := strings.Split(s, "\n")
	for _, r := range rs {
		if r == "" {
			continue
		}
		row := parseUrlBtnsLine(r)
		if len(row) == 0 {
			return nil, ErrFormatError
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseUrlBtnsLine(input string) tele.Row {
	re := regexp.MustCompile(`(\S+\s*\S+)\s+(https://\S+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	links := tele.Row{}

	for _, match := range matches {
		name := strings.TrimSpace(match[1])
		link := match[2]
		links = append(links, tele.Btn{
			Text: name,
			URL:  link,
		})
	}
	return links
}
