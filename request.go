package goinsta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type reqOptions struct {
	Endpoint     string
	PostData     string
	IsLoggedIn   bool
	IgnoreStatus bool
	Query        map[string]string
}

func (insta *Instagram) OptionalRequest(endpoint string, a ...interface{}) (body []byte, err error) {
	return insta.sendRequest(&reqOptions{
		Endpoint: fmt.Sprintf(endpoint, a...),
	})
}

func (insta *Instagram) sendSimpleRequest(endpoint string, a ...interface{}) (body []byte, err error) {
	return insta.sendRequest(&reqOptions{
		Endpoint: fmt.Sprintf(endpoint, a...),
	})
}

func (insta *Instagram) sendRequest(o *reqOptions) (body []byte, err error) {

	if !insta.IsLoggedIn && !o.IsLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	method := "GET"
	if len(o.PostData) > 0 {
		method = "POST"
	}

	u, err := url.Parse(GOINSTA_API_URL + o.Endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for k, v := range o.Query {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	var req *http.Request
	req, err = http.NewRequest(method, u.String(), bytes.NewBuffer([]byte(o.PostData)))
	if err != nil {
		return
	}

	req.Header.Set("Connection", "close")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie2", "$Version=1")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("User-Agent", GOINSTA_USER_AGENT)

	client := &http.Client{
		Jar:     insta.Cookiejar,
		Timeout: time.Minute * 3,
	}

	if insta.Proxy != "" {
		proxy, err := url.Parse(insta.Proxy)
		if err != nil {
			return body, err
		}
		insta.Transport.Proxy = http.ProxyURL(proxy)

		client.Transport = &insta.Transport
	} else {
		// Remove proxy if insta.Proxy was removed
		insta.Transport.Proxy = nil
		client.Transport = &insta.Transport
	}

	resp, err := client.Do(req)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()

	u, _ = url.Parse(GOINSTA_API_URL)
	for _, value := range insta.Cookiejar.Cookies(u) {
		if strings.Contains(value.Name, "csrftoken") {
			insta.Informations.Token = value.Value
		}
	}

	// if insta.Proxy != "" {
	// 	if sz := respSize(resp); sz > 15000 {
	// 		log.Println("Large hit:", GOINSTA_API_URL+o.Endpoint, sz)
	// 	}
	// }

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 && !o.IgnoreStatus {
		e := fmt.Errorf("Invalid status code %s %d", string(body), resp.StatusCode)
		switch resp.StatusCode {
		case 400:
			var load ErrorLoad
			json.Unmarshal(body, &load)
			if load.ErrorType == "bad_password" || load.ErrorType == "invalid_user" || load.ErrorType == "unusable_password" {
				e = ErrBadPassword
			} else if load.ErrorType == "checkpoint_challenge_required" {
				e = ErrChallenge
			} else if load.Message == "Not authorized to view user" {
				e = ErrPrivate
			} else {
				e = ErrLoggedOut
				log.Println("Logged out!", load.ErrorType, string(body))
			}
		case 403:
			if strings.Contains(strings.ToLower(string(body)), "login_required") {
				e = ErrLoggedOut
			}
		case 404:
			e = ErrNotFound
		}
		return body, e
	}

	return body, err
}

func respSize(resp *http.Response) int {
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(&buf)
	return buf.Len()
}
