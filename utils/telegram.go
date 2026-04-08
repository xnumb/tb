package utils

import (
	"errors"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

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

// ToHTML 将当前上下文中的消息正文（与 c.Text() / c.Entities() 一致，含 caption）转为 Bot API 的 HTML。
// customEmojiOnly 为 true 时只把 custom_emoji 实体包成 <tg-emoji>，其余文本与手写标签原样保留；
// 为 false 时按实体生成 <b>、<code>、<tg-spoiler>、<tg-emoji>、<a>、<blockquote> 等标签。
func ToHTML(c tele.Context, customEmojiOnly bool) string {
	msg := c.Message()
	text := c.Text()
	if text == "" {
		return ""
	}
	if msg == nil {
		if customEmojiOnly {
			return text
		}
		return html.EscapeString(text)
	}
	raw := c.Entities()
	if customEmojiOnly {
		return textWithCustomEmojiHTML(msg, text, raw)
	}
	return messageEntitiesToFullHTML(msg, text, raw)
}

func textWithCustomEmojiHTML(msg *tele.Message, text string, raw tele.Entities) string {
	var ents []tele.MessageEntity
	for _, e := range raw {
		if e.Type == tele.EntityCustomEmoji {
			ents = append(ents, e)
		}
	}
	if len(ents) == 0 {
		return text
	}
	sort.Slice(ents, func(i, j int) bool {
		if ents[i].Offset != ents[j].Offset {
			return ents[i].Offset < ents[j].Offset
		}
		return ents[i].Length < ents[j].Length
	})
	var b strings.Builder
	cur := 0
	n := utf16Len(text)
	for _, e := range ents {
		if e.Offset < cur {
			continue
		}
		if e.Offset > cur {
			b.WriteString(substringByUTF16(text, cur, e.Offset))
		}
		inner := msg.EntityText(e)
		if e.CustomEmojiID == "" {
			b.WriteString(inner)
		} else {
			b.WriteString(`<tg-emoji emoji-id="`)
			b.WriteString(e.CustomEmojiID)
			b.WriteString(`">`)
			b.WriteString(htmlEscapeMinimal(inner))
			b.WriteString(`</tg-emoji>`)
		}
		cur = e.Offset + e.Length
		if cur > n {
			cur = n
		}
	}
	if cur < n {
		b.WriteString(substringByUTF16(text, cur, n))
	}
	return b.String()
}

type entityStackItem struct {
	e   tele.MessageEntity
	end int
}

func messageEntitiesToFullHTML(msg *tele.Message, text string, raw tele.Entities) string {
	a := utf16.Encode([]rune(text))
	n := len(a)
	var ents []tele.MessageEntity
	for _, e := range raw {
		if e.Length <= 0 || !entityNeedsHTMLStack(e.Type) {
			continue
		}
		end := e.Offset + e.Length
		if e.Offset < 0 || end > n {
			continue
		}
		ents = append(ents, e)
	}
	if len(ents) == 0 {
		return html.EscapeString(text)
	}

	openAt := make(map[int][]tele.MessageEntity)
	for _, e := range ents {
		openAt[e.Offset] = append(openAt[e.Offset], e)
	}
	for off := range openAt {
		es := openAt[off]
		sort.Slice(es, func(i, j int) bool {
			return es[i].Length > es[j].Length
		})
		openAt[off] = es
	}

	var b strings.Builder
	var stack []entityStackItem
	for i := 0; i < n; {
		for len(stack) > 0 && stack[len(stack)-1].end == i {
			top := stack[len(stack)-1]
			b.WriteString(entityCloseTag(top.e))
			stack = stack[:len(stack)-1]
		}
		for _, e := range openAt[i] {
			open := entityOpenTag(e, msg)
			if open == "" {
				continue
			}
			b.WriteString(open)
			stack = append(stack, entityStackItem{e: e, end: e.Offset + e.Length})
		}

		r1 := a[i]
		var r rune
		var w int
		if r1 >= 0xd800 && r1 <= 0xdbff && i+1 < n {
			r2 := a[i+1]
			if r2 >= 0xdc00 && r2 <= 0xdfff {
				r = utf16.DecodeRune(rune(r1), rune(r2))
				w = 2
			} else {
				r = rune(r1)
				w = 1
			}
		} else {
			r = rune(r1)
			w = 1
		}
		b.WriteString(html.EscapeString(string(r)))
		i += w
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		b.WriteString(entityCloseTag(top.e))
		stack = stack[:len(stack)-1]
	}
	return b.String()
}

func entityNeedsHTMLStack(t tele.EntityType) bool {
	switch t {
	case tele.EntityBold, tele.EntityItalic, tele.EntityUnderline, tele.EntityStrikethrough,
		tele.EntityCode, tele.EntityCodeBlock, tele.EntityTextLink, tele.EntitySpoiler,
		tele.EntityCustomEmoji, tele.EntityBlockquote, tele.EntityEBlockquote,
		tele.EntityTMention, tele.EntityURL, tele.EntityEmail, tele.EntityPhone:
		return true
	default:
		return false
	}
}

func entityOpenTag(e tele.MessageEntity, msg *tele.Message) string {
	switch e.Type {
	case tele.EntityBold:
		return "<b>"
	case tele.EntityItalic:
		return "<i>"
	case tele.EntityUnderline:
		return "<u>"
	case tele.EntityStrikethrough:
		return "<s>"
	case tele.EntityCode:
		return "<code>"
	case tele.EntityCodeBlock:
		if e.Language != "" {
			return `<pre><code class="language-` + htmlEscapeAttrClass(e.Language) + `">`
		}
		return "<pre>"
	case tele.EntityTextLink:
		return `<a href="` + htmlEscapeAttr(e.URL) + `">`
	case tele.EntitySpoiler:
		return "<tg-spoiler>"
	case tele.EntityCustomEmoji:
		if msg == nil || e.CustomEmojiID == "" {
			return ""
		}
		return `<tg-emoji emoji-id="` + htmlEscapeAttr(e.CustomEmojiID) + `">`
	case tele.EntityBlockquote:
		return "<blockquote>"
	case tele.EntityEBlockquote:
		return "<blockquote expandable>"
	case tele.EntityTMention:
		if e.User == nil {
			return ""
		}
		return `<a href="tg://user?id=` + strconv.FormatInt(e.User.ID, 10) + `">`
	case tele.EntityURL:
		inner := msg.EntityText(e)
		return `<a href="` + htmlEscapeAttr(inner) + `">`
	case tele.EntityEmail:
		inner := msg.EntityText(e)
		return `<a href="mailto:` + htmlEscapeAttr(inner) + `">`
	case tele.EntityPhone:
		inner := msg.EntityText(e)
		href := strings.TrimSpace(inner)
		href = strings.ReplaceAll(href, " ", "")
		return `<a href="tel:` + htmlEscapeAttr(href) + `">`
	default:
		return ""
	}
}

func entityCloseTag(e tele.MessageEntity) string {
	switch e.Type {
	case tele.EntityBold:
		return "</b>"
	case tele.EntityItalic:
		return "</i>"
	case tele.EntityUnderline:
		return "</u>"
	case tele.EntityStrikethrough:
		return "</s>"
	case tele.EntityCode:
		return "</code>"
	case tele.EntityCodeBlock:
		if e.Language != "" {
			return "</code></pre>"
		}
		return "</pre>"
	case tele.EntityTextLink, tele.EntityTMention, tele.EntityURL, tele.EntityEmail, tele.EntityPhone:
		return "</a>"
	case tele.EntitySpoiler:
		return "</tg-spoiler>"
	case tele.EntityCustomEmoji:
		return "</tg-emoji>"
	case tele.EntityBlockquote, tele.EntityEBlockquote:
		return "</blockquote>"
	default:
		return ""
	}
}

func htmlEscapeMinimal(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func htmlEscapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func htmlEscapeAttrClass(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteString("&#x")
			b.WriteString(strconv.FormatInt(int64(r), 16))
			b.WriteString(";")
		}
	}
	return b.String()
}

func utf16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func substringByUTF16(s string, start, end int) string {
	if s == "" || start >= end {
		return ""
	}
	a := utf16.Encode([]rune(s))
	if start < 0 {
		start = 0
	}
	if end > len(a) {
		end = len(a)
	}
	if start >= end {
		return ""
	}
	return string(utf16.Decode(a[start:end]))
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
