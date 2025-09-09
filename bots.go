package tb

import (
	"errors"

	"github.com/xnumb/tb/log"
	tele "gopkg.in/telebot.v4"
)

var ErrBotStartFailed = errors.New("bot启动失败")

type Tbots struct {
	Bots    map[uint]*Tbot
	Gen     func(token string) *Tbot
	OnStart func(id uint, u *tele.User) // u为nil则表示启动失败或者stop
}

func NewTbots(gen func(string) *Tbot, onStart func(id uint, u *tele.User)) *Tbots {
	return &Tbots{
		Bots:    make(map[uint]*Tbot),
		Gen:     gen,
		OnStart: onStart,
	}
}

func (m *Tbots) Start(id uint, token string) bool {
	if _, exist := m.Bots[id]; exist {
		return false
	}
	b := m.Gen(token)
	if b == nil {
		m.OnStart(id, nil)
		log.Err(ErrBotStartFailed, "id", id)
		return false
	}
	m.OnStart(id, b.client.Me)
	m.Bots[id] = b
	go b.Start()
	return true
}

func (m *Tbots) Stop(id uint) {
	if b, ok := m.Bots[id]; ok {
		b.client.Stop()
		m.OnStart(id, nil)
	}
}

func (m *Tbots) Get(id uint) *Tbot {
	if b, ok := m.Bots[id]; ok {
		return b
	}
	return nil
}
