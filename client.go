package tb

import (
	"strings"
	"time"

	"github.com/xnumb/tb/log"
	"github.com/xnumb/tb/to"
	tele "gopkg.in/telebot.v4"
)

func (t *Tbot) Client() *tele.Bot {
	return t.client
}

func (t *Tbot) SetMenus(menus Menus, userId int64) {
	var cmds []tele.Command
	for _, m := range menus {
		cmds = append(cmds, m.gen())
		t.client.Handle("/"+m.ID, m.Fn)
	}
	if err := t.client.SetCommands(cmds); err != nil {
		log.Err(err)
	}
	u := tele.User{}
	if userId != 0 {
		u.ID = userId
	}
	_ = t.client.SetMenuButton(&u, tele.MenuButtonDefault)
}

func (t *Tbot) registerCallback(btns Btns, asker Asker, expMin int64) {
	t.client.Handle(tele.OnCallback, func(c tele.Context) error {
		s := c.Data()
		if s == CmdEmpty || s == "" {
			return c.Respond()
		}
		msg := c.Message()
		// 过期检测
		if msg == nil {
			return nil
		}
		if expMin > 0 && msg.Unixtime+expMin*60 < time.Now().Unix() {
			if c.Callback() == nil {
				return nil
			}
			return c.RespondAlert("按钮已过期")
			//return c.Respond(&tele.CallbackResponse{
			//	CallbackID: c.Callback().ID,
			//	Text:       "按钮已过期",
			//	ShowAlert:  true,
			//})
		}
		id, page, args := parseCmd(s)
		//log.Info("line", "id", id, "page", page, "args", args)
		sysAction := ""
		if id == CmdDel {
			if err := t.DelMessageFrom(c.Sender().ID, msg.ID); err != nil {
				log.Err(err, "chatId", c.Sender().ID, "msgId", msg.ID)
			}
			return c.Respond()
		} else if id == cmdCancelAsk {
			sid := c.Sender().ID
			_, err := asker.Done(sid)
			if err != nil {
				log.Err(err, "chatId", sid, "msgId", msg.ID)
			}
			// 删除问题
			if err = t.DelMessageFrom(sid, msg.ID); err != nil {
				log.Err(err, "chatId", sid, "msgId", msg.ID)
			}
		} else if strings.HasPrefix(id, cmdPrefixCancelConfirm) {
			sysAction = cmdPrefixCancelConfirm
			id = strings.TrimPrefix(id, cmdPrefixCancelConfirm)
		} else if strings.HasPrefix(id, cmdPrefixConfirmed) {
			sysAction = cmdPrefixConfirmed
			id = strings.TrimPrefix(id, cmdPrefixConfirmed)
		}
		for _, btn := range btns {
			if btn.ID == id {
				argsMap := btn.genArgsMap(args)
				if !btn.Respond {
					_ = c.Respond()
				}
				if sysAction != "" {
					if sysAction == cmdPrefixCancelConfirm {
						if btn.confirmAction != nil {
							return btn.confirmAction.CancelFn(c, argsMap)
						}
					} else if sysAction == cmdPrefixConfirmed {
						if btn.confirmAction != nil {
							return btn.confirmAction.Fn(c, argsMap)
						}
					}
					return nil
				}

				if btn.action != nil {
					return btn.action.Fn(c, argsMap)
				}
				if btn.pageAction != nil {
					return btn.pageAction.Fn(c, page, argsMap)
				}
				if btn.askAction != nil {
					return btn.SendAsk(c, btn.askAction.IsEdit, asker, args...)
				}
				if btn.confirmAction != nil {
					return btn.SendConfirm(c, !btn.confirmAction.IsSend, args...)
				}
				return nil
			}
		}
		return c.Respond()
	})
}

func (t *Tbot) Start() {
	t.client.Start()
}

func (t *Tbot) SendTo(cid int64, p SendParams, opts ...any) (*tele.Message, error) {
	return sendTo(t.client, cid, p, opts...)
}

func (t *Tbot) SendToTopic(cid int64, topicId int, p SendParams, opts ...any) (*tele.Message, error) {
	return sendToTopic(t.client, cid, topicId, p, opts...)
}

func (t *Tbot) SendToUsername(username string, p SendParams, opts ...any) (*tele.Message, error) {
	return sendToUsername(t.client, username, p, opts...)
}

func (t *Tbot) SendAlbum(cid int64, info, medias string, opts ...any) ([]tele.Message, error) {
	return sendAlbumTo(t.client, cid, info, medias, opts...)
}

func (t *Tbot) DelMessageFrom(chatId int64, messageId int) error {
	_, err := t.client.Raw("deleteMessage", Args{
		"chat_id":    to.S(chatId),
		"message_id": to.S(messageId),
	})
	return err
}
