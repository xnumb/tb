package utils

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
