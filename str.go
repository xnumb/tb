package tb

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/xnumb/tb/emj"
	tele "gopkg.in/telebot.v4"
)

func Info(title string, lines ...string) string {
	title = WrapTitle(title, emj.D, 0)
	lines = append([]string{title}, lines...)
	return Infos(lines...)
}

func Infos(lines ...string) string {
	return strings.Join(lines, "\n")
}

// WrapTitle count: 0代表自动
func WrapTitle(title, emoji string, count int) string {
	if count == 0 {
		l := utf8.RuneCountInString(title)
		const maxLen = 14
		if l >= maxLen {
			return title
		} else {
			c := (maxLen - l) / 2
			s := strings.Repeat(emoji, c)
			return fmt.Sprintf("%s %s %s", s, title, s)
		}
	} else {
		s := strings.Repeat(emoji, count)
		return fmt.Sprintf("%s %s %s", s, title, s)
	}
}

func BoolStr(v bool, tStr, fStr string) string {
	if v {
		return tStr
	}
	return fStr
}

func CheckStr(v bool, s string) string {
	return BoolStr(v, emj.Check+" "+s, emj.Uncheck+" "+s)
}
func RadioStr(v bool, s string) string {
	return BoolStr(v, emj.Check+" "+s, emj.Radio+" "+s)
}

func GroRow(t string) tele.Row {
	return GroRowFmt(t, emj.Down, 1)
}

func GroRowFmt(title, emoji string, count int) tele.Row {
	return tele.Row{
		{
			Text: WrapTitle(title, emoji, count),
			Data: CmdEmpty,
		},
	}
}
