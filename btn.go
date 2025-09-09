package tb

import (
	"errors"

	"github.com/xnumb/tb/emj"
	"github.com/xnumb/tb/log"
	tele "gopkg.in/telebot.v4"
)

var (
	ErrBtnDefinedInvalid   = errors.New("按钮定义错误")
	ErrBtnArgsCountInvalid = errors.New("传参数量与按钮定义不一致")
)

var BtnDelMessage = tele.Btn{
	Text: emj.X + " 关闭",
	Data: CmdDel,
}

type Btn struct {
	// 命令ID, 可以通过Cmd(id来调用) 首字母大写
	// [_sysCmdId.?][cmdId], sysCmdId: ask/confirm/del/(后面是系统自动构建)confirmed/cancelAsk
	ID string
	// 按钮文字, 可以在生成的时候使用GenWithText()等函数临时修改文字,
	// 空字符串则不可以使用Gen()等函数生成tele.Btn
	Text string
	// 参数列表, 在Fn等回调函数会解析成Args类型的参数,可以按key取值
	// 通过Gen()等函数生成tele.Btn的时候传参数量需要定义数量一致
	Args []string
	// 手动处理 Respond
	Respond       bool
	action        *Action
	pageAction    *PageAction
	askAction     *AskAction
	confirmAction *ConfirmAction
}

type Btns []*Btn

type Action struct {
	Fn func(c tele.Context, args Args) error
}

type PageAction struct {
	Fn func(c tele.Context, page int, args Args) error
}

type AskAction struct {
	// 询问文字, 可根据args构建输出
	Q func(c tele.Context, args Args) string
	// 返回的done(bool)代表是否需要完成该ask
	Fn func(c tele.Context, val string, args Args) (bool, error)
	// 询问是否需要编辑当前消息, 默认是新发送问题
	IsEdit bool
	// 是否保留问题, 默认删除消息
	KeepQuiz bool
}

type ConfirmAction struct {
	// 询问文字, 可根据args构建输出
	Q  func(c tele.Context, args Args) string
	Fn func(c tele.Context, args Args) error
	// _confirm 取消按钮的回调 这里用不定参数主要是为了能返回上页可能存在的筛选与page参数
	CancelFn func(c tele.Context, args Args) error
	// _confirm询问是否需要编辑当前消息, 默认是编辑
	IsSend bool
}

func (b *Btn) Link(action Action) *Btn {
	return &Btn{
		ID:      b.ID,
		Text:    b.Text,
		Args:    b.Args,
		Respond: b.Respond,
		action:  &action,
	}
}
func (b *Btn) LinkPage(action PageAction) *Btn {
	return &Btn{
		ID:         b.ID,
		Text:       b.Text,
		Args:       b.Args,
		Respond:    b.Respond,
		pageAction: &action,
	}
}

func (b *Btn) LinkAsk(action AskAction) *Btn {
	return &Btn{
		ID:        b.ID,
		Text:      b.Text,
		Args:      b.Args,
		Respond:   b.Respond,
		askAction: &action,
	}
}
func (b *Btn) LinkConfirm(action ConfirmAction) *Btn {
	return &Btn{
		ID:            b.ID,
		Text:          b.Text,
		Args:          b.Args,
		Respond:       b.Respond,
		confirmAction: &action,
	}
}

func (b *Btn) checkArgs(args []string) {
	if len(b.Args) != len(args) {
		log.Err(ErrBtnArgsCountInvalid, "btn.id", b.ID, "btn.args", b.Args, "input args", args)
	}
}

func (b *Btn) genArgsMap(args []string) Args {
	b.checkArgs(args)
	var argsMap = make(Args)
	for i, arg := range b.Args {
		if i >= len(args) {
			argsMap[arg] = ""
		} else {
			argsMap[arg] = args[i]
		}
	}
	return argsMap
}

func (b *Btn) gen(text string, page int, args ...string) tele.Btn {
	btn := tele.Btn{
		Text: "ErrorBtn",
		Data: CmdEmpty,
	}
	if text == "" {
		log.Err(ErrBtnDefinedInvalid, "btn.id", b.ID, "btn文字为空")
		return btn
	}
	if page < 0 {
		log.Err(ErrBtnDefinedInvalid, "btn.id", b.ID, "page参数设置为负数")
		return btn
	}
	if b.pageAction != nil && page == 0 {
		log.Err(ErrBtnDefinedInvalid, "btn.id", b.ID, "需要设置page参数")
		return btn
	}
	b.checkArgs(args)
	return tele.Btn{
		Text: text,
		Data: genCmd(b.ID, page, args...),
	}
}

func (b *Btn) G(args ...string) tele.Btn {
	return b.gen(b.Text, 0, args...)
}
func (b *Btn) P(page int, args ...string) tele.Btn {
	return b.gen(b.Text, page, args...)
}

func (b *Btn) T(text string, args ...string) tele.Btn {
	return b.gen(text, 0, args...)
}

func (b *Btn) TP(text string, page int, args ...string) tele.Btn {
	return b.gen(text, page, args...)
}

func (b *Btn) SendAsk(c tele.Context, isEdit bool, asker Asker, args ...string) error {
	action := b.askAction
	if action == nil {
		log.Err(ErrBtnDefinedInvalid, "btn.id", b.ID)
		return nil
	}
	sid := c.Sender().ID
	cmd := genCmd(b.ID, 0, args...)
	msgId := c.Message().ID // 不是按钮触发的没有消息 todo 这里看是否有Message.id 是不是0
	if err := asker.Set(sid, cmd, msgId); err != nil {
		log.Err(err, "sid", sid, "cmd", cmd, "msgId", msgId)
		return err
	}
	argsMap := b.genArgsMap(args)
	txt := action.Q(c, argsMap)
	if txt == "" {
		log.Err(nil, "btn.id", b.ID, "args", argsMap)
		txt = "未定义问题内容"
	}
	return Send(c, SendParams{
		IsEdit: isEdit,
		Info:   txt,
		Rows: []tele.Row{
			{
				{
					Text: emj.X + " 取消",
					Data: cmdCancelAsk,
				},
			},
		},
	}, tele.ForceReply)
}

func (b *Btn) SendConfirm(c tele.Context, isEdit bool, args ...string) error {
	action := b.confirmAction
	if action == nil {
		log.Err(ErrBtnDefinedInvalid, "btn.id", b.ID)
		return nil
	}
	argsMap := b.genArgsMap(args)
	info := action.Q(c, argsMap)
	if info == "" {
		log.Err(nil, "btn.id", b.ID, "args", argsMap)
		return nil
	}
	return Send(c, SendParams{
		IsEdit: isEdit,
		Info:   info,
		Rows: []tele.Row{
			{
				{
					Text: emj.Check + " 确认",
					Data: genCmd(cmdPrefixConfirmed+b.ID, 0, args...),
				},
				{
					Text: emj.X + " 取消",
					Data: genCmd(cmdPrefixCancelConfirm+b.ID, 0, args...),
				},
			},
		},
	})
}

func (btns Btns) CheckAsker(c tele.Context, asker Asker) bool {
	sid := c.Sender().ID
	cmd := asker.Get(sid)
	if cmd == "" {
		return false
	}
	id, _, args := parseCmd(cmd)
	for _, btn := range btns {
		if btn.ID == id {
			action := btn.askAction
			if action == nil {
				log.Err(ErrBtnDefinedInvalid, "btn.id", btn.ID)
				return false
			}
			argsMap := btn.genArgsMap(args)
			done, err := action.Fn(c, c.Text(), argsMap)
			if err != nil {
				log.Err(err, "sid", sid, "cmd", cmd, "args", argsMap)
			}
			if !done {
				return true
			}
			msgId, err := asker.Done(sid)
			if err != nil {
				log.Err(err, "sid", sid, "cmd", cmd, "msgId", msgId)
			}
			if !action.KeepQuiz && msgId > 0 {
				if err = DelMessageFrom(c, sid, msgId+1); err != nil {
					log.Err(err, "sid", sid, "cmd", cmd, "msgId", msgId)
				}
			}
			return true
		}
	}
	return false
}
