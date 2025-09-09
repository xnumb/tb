package tb

import (
	"strings"

	"github.com/xnumb/tb/emj"
	"github.com/xnumb/tb/log"
	"github.com/xnumb/tb/to"
	"github.com/xnumb/tb/utils"

	tele "gopkg.in/telebot.v4"
)

type SendPage struct {
	// 第几页
	No int
	// 总页数
	Count int
	// 数据条数
	Total int64
	// 进入当前页面的按钮/命令ID
	CmdId string
	// 进入当前页面的按钮/命令的参数
	CmdArgs []string
	// 数据条数为0时的提示文字
	EmptyText string
	// 忽略空数据(也就是不显示暂无数据, 用作文本分页)
	IgnoreEmpty bool
	// 是否需要数字分页
	NumMode bool
}

type SendParams struct {
	IsEdit  bool
	IsReply bool
	// 文字信息, 可以使用Head()生成, 可以拼接多组Head()的结果
	Info string
	// 顶部按钮
	HeadRows []tele.Row
	// 主按钮区域的按钮
	Rows []tele.Row
	// 底部按钮
	FootRows []tele.Row
	// 分页信息
	Page *SendPage
	// fileId/url/path
	Pic string
	// fileId/url/path
	Vod string
	// 底部菜单更新 不为空则忽略rows...
	KbBtns Kbts
	// 原生Markup 如果存在则忽略rows与KbBtns
	RawMarkup *tele.ReplyMarkup
}

// ReceiveMedia text, mediaId
func ReceiveMedia(c tele.Context) (string, string) {
	text := c.Text()
	mediaId := ""
	pic := c.Message().Photo
	if pic != nil {
		mediaId = pic.FileID
	}
	vod := c.Message().Video
	if vod != nil {
		mediaId = "_" + vod.FileID
	}
	return text, mediaId
}

func (sp *SendParams) SetMedia(text, mediaId string, btns string, needDelBtn bool) bool {
	var msgBtns []tele.Row
	picId := ""
	vodId := ""
	if strings.HasPrefix(mediaId, "_") {
		vodId = strings.TrimPrefix(mediaId, "_")
	} else {
		picId = mediaId
	}
	if btns != "" {
		msgBtns, _ = utils.GenUrlBtns(btns)
	}
	if needDelBtn {
		if len(msgBtns) == 0 {
			msgBtns = []tele.Row{
				{
					BtnDelMessage,
				},
			}
		} else {
			msgBtns = append(msgBtns, tele.Row{
				BtnDelMessage,
			})
		}
	}
	if text == "" && picId == "" && vodId == "" {
		return false
	}
	sp.Info = text
	sp.Pic = picId
	sp.Vod = vodId
	sp.Rows = msgBtns
	return true
}

func buildSend(p *SendParams, opts ...any) (any, []any) {
	markup := &tele.ReplyMarkup{}
	if p.RawMarkup != nil {
		markup = p.RawMarkup
	} else if p.KbBtns != nil {
		var kbBtns []tele.Row
		for _, r := range p.KbBtns {
			row := tele.Row{}
			for _, b := range r {
				row = append(row, tele.Btn{
					Text: b,
				})
			}
			kbBtns = append(kbBtns, row)
		}
		markup.ResizeKeyboard = true
		markup.Reply(kbBtns...)
	} else {
		// 构建 rows
		var rows []tele.Row
		if p.HeadRows != nil {
			rows = append(rows, p.HeadRows...)
		}
		pp := p.Page
		if len(p.Rows) == 0 && pp != nil && !pp.IgnoreEmpty {
			btnText := emj.Empty + " 暂无数据"
			if pp.EmptyText != "" {
				btnText = pp.EmptyText
			}
			rows = append(rows, tele.Row{
				tele.Btn{
					Text: btnText,
					Data: CmdEmpty,
				},
			})
		} else {
			rows = append(rows, p.Rows...)
		}

		// 构建 pageRow
		if pp != nil && pp.Count > 1 {
			var pageRow tele.Row
			if pp.NumMode {
				if pp.Count <= 7 {
					for i := 1; i <= pp.Count; i++ {
						if pp.No == i {
							pageRow = append(pageRow, tele.Btn{
								Text: "(" + to.S(i) + ")",
								Data: CmdEmpty,
							})
						} else {
							pageRow = append(pageRow, tele.Btn{
								Text: to.S(i),
								Data: genCmd(pp.CmdId, i, pp.CmdArgs...),
							})
						}
					}
				} else {
					startNo := 0
					if pp.No <= 3 {
						startNo = 1
					} else if pp.No >= pp.Count-2 {
						startNo = pp.Count - 4
					} else {
						startNo = pp.No - 2
					}
					firstBtn := tele.Btn{
						Text: emj.First,
						Data: genCmd(pp.CmdId, 1, pp.CmdArgs...),
					}
					lastBtn := tele.Btn{
						Text: emj.Last,
						Data: genCmd(pp.CmdId, pp.Count, pp.CmdArgs...),
					}
					pageRow = tele.Row{firstBtn}
					for i := startNo; i <= startNo+4; i++ {
						if pp.No == i {
							pageRow = append(pageRow, tele.Btn{
								Text: "(" + to.S(i) + ")",
								Data: CmdEmpty,
							})
						} else {
							pageRow = append(pageRow, tele.Btn{
								Text: to.S(i),
								Data: genCmd(pp.CmdId, i, pp.CmdArgs...),
							})
						}
					}
					pageRow = append(pageRow, lastBtn)
				}
			} else {
				prevCmd := CmdEmpty
				if pp.No > 1 {
					prevCmd = genCmd(pp.CmdId, pp.No-1, pp.CmdArgs...)
				}
				prevBtn := tele.Btn{
					Text: emj.Left,
					Data: prevCmd,
				}
				nextCmd := CmdEmpty
				if pp.No < pp.Count {
					nextCmd = genCmd(pp.CmdId, pp.No+1, pp.CmdArgs...)
				}
				nextBtn := tele.Btn{
					Text: emj.Right,
					Data: nextCmd,
				}
				pageRow = tele.Row{
					prevBtn,
					tele.Btn{
						Text: "第" + to.S(pp.No) + "/" + to.S(pp.Count) + "页",
						Data: CmdEmpty,
					},
					nextBtn,
				}
			}
			rows = append(rows, pageRow)
		}
		if p.FootRows != nil {
			rows = append(rows, p.FootRows...)
		}
		markup.Inline(rows...)
	}
	// 构建 content
	var content any
	if p.Pic != "" {
		photo := &tele.Photo{
			Caption: p.Info,
		}
		if strings.HasPrefix(p.Pic, "https://") { // url
			photo.File = tele.FromURL(p.Pic)
		} else if strings.Contains(p.Pic, ".") { // file
			photo.File = tele.FromDisk(p.Pic)
		} else { // fileId
			photo.FileID = p.Pic
		}
		content = photo
	} else if p.Vod != "" {
		video := &tele.Video{
			Caption: p.Info,
		}
		if strings.HasPrefix(p.Pic, "https://") { // url
			video.File = tele.FromURL(p.Vod)
		} else if strings.Contains(p.Pic, ".") { // file
			video.File = tele.FromDisk(p.Vod)
		} else { // fileId
			video.FileID = p.Vod
		}
		content = video
	} else {
		content = p.Info
	}
	opts2 := []any{
		markup,
	}
	if len(opts) > 0 {
		opts2 = append(opts2, opts...)
	}
	return content, opts2
}

func buildAlbum(info, medias string) tele.Album {
	var as tele.Album
	ms := strings.Split(medias, ",")
	for i, m := range ms {
		if strings.HasPrefix(m, "_") {
			m = strings.TrimPrefix(m, "_")
			a := &tele.Video{
				File: tele.File{FileID: m},
			}
			if i == 0 {
				a.Caption = info
			}
			as = append(as, a)
		} else {
			a := &tele.Photo{
				File: tele.File{FileID: m},
			}
			if i == 0 {
				a.Caption = info
			}
			as = append(as, a)
		}
	}
	return as
}

// Send 建议 sendX 系列函数的参数顺序: sendX(c tele.Context, isEdit bool[, page int, ...]) error

func Send(c tele.Context, p SendParams, opts ...any) error {
	content, opts2 := buildSend(&p, opts...)
	// opts2 = append(opts2, tele.IgnoreThread) // maybe panic
	if p.IsEdit {
		return c.Edit(content, opts2...)
	} else if p.IsReply {
		return c.Reply(content, opts2...)
	} else {
		return c.Send(content, opts2...)
	}
}

func SendText(c tele.Context, text string) error {
	return Send(c, SendParams{
		Info: text,
	})
}

func ReplyText(c tele.Context, text string) error {
	return Send(c, SendParams{
		IsReply: true,
		Info:    text,
	})
}

func sendTo(bot tele.API, cid int64, p SendParams, opts ...any) (*tele.Message, error) {
	content, opts2 := buildSend(&p, opts...)
	return bot.Send(&tele.Chat{
		ID: cid,
	}, content, opts2...)
}

func SendTo(c tele.Context, cid int64, p SendParams, opts ...any) (*tele.Message, error) {
	return sendTo(c.Bot(), cid, p, opts...)
}

func sendToTopic(bot tele.API, cid int64, topicId int, p SendParams, opts ...any) (*tele.Message, error) {
	content, opts2 := buildSend(&p, opts...)
	opts2 = append(opts2, tele.SendOptions{
		ThreadID: topicId,
	})
	return bot.Send(&tele.Chat{
		ID: cid,
	}, content, opts2...)
}

func SendToTopic(c tele.Context, cid int64, topicId int, p SendParams, opts ...any) (*tele.Message, error) {
	return sendToTopic(c.Bot(), cid, topicId, p, opts...)
}

func sendToUsername(bot tele.API, username string, p SendParams, opts ...any) (*tele.Message, error) {
	chat, err := bot.ChatByUsername(username)
	if err != nil {
		return nil, err
	}
	content, opts2 := buildSend(&p, opts...)
	return bot.Send(&tele.Chat{
		ID: chat.ID,
	}, content, opts2...)
}

func SendToUsername(c tele.Context, username string, p SendParams, opts ...any) (*tele.Message, error) {
	return sendToUsername(c.Bot(), username, p, opts...)
}

func sendAlbumTo(bot tele.API, cid int64, info, medias string, opts ...any) ([]tele.Message, error) {
	as := buildAlbum(info, medias)
	return bot.SendAlbum(&tele.Chat{
		ID: cid,
	}, as, opts...)
}

func SendAlbum(c tele.Context, info, medias string, opts ...any) ([]tele.Message, error) {
	return sendAlbumTo(c.Bot(), c.Sender().ID, info, medias, opts...)
}

func SendAlbumTo(c tele.Context, cid int64, info, medias string, opts ...any) ([]tele.Message, error) {
	return sendAlbumTo(c.Bot(), cid, info, medias, opts...)
}

func SendErr(c tele.Context, err error) error {
	log.Err(err)
	return c.Send(emj.Bell + " 操作异常: " + err.Error())
}

func Alert(c tele.Context, text string) error {
	return c.RespondAlert(text)
	//if c.Callback() == nil {
	//	return nil
	//}
	//return c.Respond(&tele.CallbackResponse{
	//	CallbackID: c.Callback().ID,
	//	Text:       text,
	//	ShowAlert:  true,
	//})
}

func Toast(c tele.Context, text string) error {
	return c.RespondText(text)
}

func DelMessage(c tele.Context) error {
	return c.Bot().Delete(c.Message())
}

func DelMessageFrom(c tele.Context, chatId int64, msgId int) error {
	_, err := c.Bot().Raw("deleteMessage", Args{
		"chat_id":    to.S(chatId),
		"message_id": to.S(msgId),
	})
	return err
}
