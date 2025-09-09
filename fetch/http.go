package fetch

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"tb2/log"
	"tb2/tberr"
)

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
		return err
	}
	return nil
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
		return nil, tberr.ErrRequestNotFound
	}
	req.Header.Set("Content-Type", "application/json")
	for hk, hv := range h {
		req.Header.Add(hk, hv)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Err(err, "http.do close error")
		}
	}(res.Body)

	body, _ := io.ReadAll(res.Body)

	return body, nil
}
