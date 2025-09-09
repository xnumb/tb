package tb

import tele "gopkg.in/telebot.v4"

type Menu struct {
	ID   string
	Desc string
	Fn   func(tele.Context) error
}
type Menus []*Menu

func (m *Menu) gen() tele.Command {
	return tele.Command{
		Text:        m.ID,
		Description: m.Desc,
	}
}
