```go
package main

var CloneAlert chan uint
var tbots *tb.Tbots

func init() {
	CloneAlert = make(chan uint)
}

func main() {
	run()
}

func run() {
	mod.Reset()
	log.Info("Start client bot.")
	tbots = tb.NewTbots(client.Gen, func(id uint, u *tele.User) {
		if u == nil {
			return
		}
		b := mod.Bot{}
		if err := b.Get(id); err != nil {
			log.Err(err)
			return
		}
		b.Tid = u.ID
		b.Name = strings.TrimSpace(u.FirstName + " " + u.LastName)
		b.Username = u.Username
		b.Started = true
		_ = b.Save()
	})
	bots := mod.Bots{}
	if err := bots.Get(); err != nil {
		log.Fatal(err)
		return
	}
	for _, b := range bots {
		if tbots.Start(b.ID, b.Token) {
			log.Info("Start bot:", "name", b.Name, "id", b.ID)
		}
	}
	log.Info("Listen clone")
	for {
		botId := <-CloneAlert
		b := mod.Bot{}
		if err := b.Get(botId); err != nil {
			log.Err(err)
			continue
		}
		tbots.Start(botId, b.Token)
	}
}


func Clone(token string, userTid int64) error {
	b := mod.Bot{
		Token:    token,
		AdminIds: to.S(userTid),
		Welcome:  "hello",
	}
	if err := b.Add(); err != nil {
		return err
	}
	CloneAlert <- b.ID
	return nil
}

```