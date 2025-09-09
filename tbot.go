package tb

import (
	"net/http"
	"net/url"
	"time"

	"github.com/xnumb/tb/log"
	tele "gopkg.in/telebot.v4"
)

// allowUpdates: message,edited_message,channel_post,edited_channel_post,business_connection,business_message,edited_business_message,deleted_business_messages,message_reaction,message_reaction_count,inline_query,chosen_inline_result,callback_query,shipping_query,pre_checkout_query,purchased_paid_media,poll,poll_answer,my_chat_member,chat_member,chat_join_request,chat_boost,removed_chat_boost
var (
	AllowUpdatesLow    = []string{"message", "callback_query"}
	AllowUpdatesNormal = []string{"message", "channel_post", "inline_query", "callback_query"}
	AllowUpdatesHigh   = []string{"message", "channel_post", "inline_query", "callback_query", "chosen_inline_result", "my_chat_member", "chat_member", "chat_join_request"}
)

// Asker 用作获取/设置/完成Ask的接口
type Asker interface {
	Get(senderId int64) string
	Set(senderId int64, cmd string, messageId int) error
	Done(senderId int64) (int, error)
}

type Tbot struct {
	client *tele.Bot
}

type InitParams struct {
	Token        string
	Proxy        string
	AllowUpdates []string
	Asker        Asker
	// 按钮过期时间 0为不过期
	BtnExpireMin int64
	Btns         Btns
}

func New(p InitParams) (*Tbot, error) {
	var httpClient *http.Client
	if p.Proxy != "" {
		proxyUrl, err := url.Parse(p.Proxy)
		if err != nil {
			log.Fatal(err, "初始化tBot失败(ParseProxy) token", p.Token)
			return nil, err
		}
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
		}
	}
	b, err := tele.NewBot(tele.Settings{
		Token: p.Token,
		Poller: &tele.LongPoller{
			Timeout:        10 * time.Second,
			AllowedUpdates: p.AllowUpdates,
		},
		ParseMode: tele.ModeHTML,
		Client:    httpClient, //使用代理
	})
	if err != nil {
		log.Err(err, "初始化tBot失败(NewBot) token:", p.Token)
		return nil, err
	}
	bot := Tbot{
		client: b,
	}
	bot.registerCallback(p.Btns, p.Asker, p.BtnExpireMin)

	return &bot, nil
}
