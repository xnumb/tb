package fetch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/xnumb/tb/log"
)

var ErrRequestNotFound = errors.New("请求不存在")

type Params map[string]any
type Header map[string]string
type Body []byte

func (b *Body) String() string {
	return string(*b)
}

func (b *Body) Parse(obj any) error {
	err := json.Unmarshal(*b, obj)
	if err != nil {
		log.Err(err, "obj", obj)
	}
	return err
}

func Post(url string, p Params, h Header) (Body, error) {
	bd, err := do("POST", url, p, h)
	if err != nil {
		log.Err(err, "url", url, "params", p, "header", h)
		return nil, err
	}
	return bd, nil
}

func Get(url string, h Header) (Body, error) {
	bd, err := do("GET", url, nil, h)
	if err != nil {
		log.Err(err, "url", url, "header", h)
		return nil, err
	}
	return bd, nil
}

func do(method, url string, p Params, h Header) (Body, error) {
	var req *http.Request
	var err error
	if method == "POST" {
		if p != nil {
			argsStr, err := json.Marshal(p)
			if err != nil {
				return nil, err
			}
			req, err = http.NewRequest(method, url, strings.NewReader(string(argsStr)))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, ErrRequestNotFound
	}
	req.Header.Set("Content-Type", "application/json")
	for hk, hv := range h {
		req.Header.Add(hk, hv)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
