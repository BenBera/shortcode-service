package library

import (
	"fmt"
	goutils "github.com/mudphilo/go-utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func HTTPGet(remoteURL string, headers map[string]string, payload map[string]string) (string, int) {

	var fields []string

	if payload != nil {

		for key, value := range payload {

			val := fmt.Sprintf("%s=%v", key, url.QueryEscape(value))

			fields = append(fields, val)
		}
	}

	if len(fields) > 0 {

		params := strings.Join(fields, "&")
		remoteURL = fmt.Sprintf("%s?%s", remoteURL, params)

	}

	req, err := http.NewRequest("GET", remoteURL, nil)
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return "", 0
	}

	if headers != nil {

		for key, value := range headers {

			req.Header.Set(key, value)
		}
	}

	resp, err := goutils.NewNetClient().Do(req)
	if resp != nil {

		defer resp.Body.Close()
	}

	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return "", 0
	}

	st := resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Printf("got error making http request %s", err.Error())
		return "", 0
	}

	return string(body), st
}
