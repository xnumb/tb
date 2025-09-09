package tb

import (
	"github.com/xnumb/tb/log"
	tele "gopkg.in/telebot.v4"
)

type Kbts [][]string

type Kb struct {
	Text string
	Fn   func(ctx tele.Context) error
}

type Kbs []*Kb

// Apply 返回 是否处理相应函数
func (kbs Kbs) Apply(c tele.Context) bool {
	for _, r := range kbs {
		if c.Text() == r.Text {
			err := r.Fn(c)
			if err != nil {
				log.Err(err, "text", r.Text)
			}
			return true
		}
	}
	return false
}
