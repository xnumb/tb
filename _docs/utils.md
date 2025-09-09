## cron
```go
package main
import "github.com/robfig/cron/v3" 

func main() {
	c := cron.New()
	_, _ = c.AddFunc("@every 5s", func() {
		
	})
	c.Start()
}
```

## tronGrid
```go
// app/tron.go
package app

import "github.com/xnumb/tb/tronGrid"

var Tron *tronGrid.Client

func init() {
	Tron = tronGrid.New(Conf.TronGridKey, Conf.Debug)
}

```

## snowId
```go
// app/snow_id.go
package app
import "github.com/bwmarrin/snowflake"

var node *snowflake.Node
func init() {
	n, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatal(err)
	}
	node = n
}
func GetSnowId() string {
	return node.Generate().String()
}
```

## google auth
```go
// app/google_auth.go
package app

import (
	"github.com/pquerna/otp/totp"
)

func GenGoogleSecret(issuer, username string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), err
}

func VerifyGoogleCode(code, secret string) bool {
	return totp.Validate(code, secret)
}
```