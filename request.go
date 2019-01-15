package goinsta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type reqOptions struct {
	Endpoint     string
	PostData     string
	IsLoggedIn   bool
	IgnoreStatus bool
	Query        map[string]string
	IsChallenge  bool
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

	if !o.IsChallenge {
		// if this is not a challenge, keep track of the original request
		// that way we can attempt it again
		insta.reqOptions = o
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
		Jar: insta.Cookiejar,
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

	if token := insta.GetCSRFToken(); token != NoCSRFToken {
		insta.Informations.Token = token
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// if we have 200 or we can ignore the status
	// then we can safely return the response from Instagram API
	if resp.StatusCode == 200 || o.IgnoreStatus {
		return body, err
	}

	// if we got a 404 return back a NotFound error
	if resp.StatusCode == 404 {
		return body, ErrNotFound
	}

	// try to parse the error message from instagram
	var msg apiResponseMessage
	json.Unmarshal(body, &msg)

	if msg.Challenge != nil && msg.Challenge.Path != "" {
		// set the challenge path if provided
		insta.challengePath = strings.Replace(msg.Challenge.Path, "/challenge/", "challenge/", 1)
	}

	needsChallenge := false
	switch msg.StepName {
	case "select_verify_method", "verify_code", "submit_phone", "verify_email":
		needsChallenge = true
	}
	if msg.Message == "challenge_required" {
		needsChallenge = true
	}

	// if this is not a challenge error
	// return it back to the client
	if !needsChallenge {
		return body, fmt.Errorf("API Response: %v", string(body))
	}

	// if no challenge settings are provided then
	if !insta.canChallenge() {
		return body, ErrChallengeOptionsRequired
	}

	_, e := insta.requestChallengeCode()
	if e != nil {
		return body, fmt.Errorf("API Challenge Request Error: %v", e)
	}

	// if we don't have a challenge code. tell the client
	if insta.challengeOptions.Code == "" {
		return body, ErrChallengeCodeRequired
	}

	// try to submit the challenge code
	_, e = insta.submitChallengeCode()
	if e != nil {
		return body, fmt.Errorf("API Challenge Code Error: %v", e)
	}

	// if we were able to successfully submit the code
	// try the api call again
	return insta.sendRequest(o)

}

type apiResponseMessage struct {
	Message       string             `json:"message"`
	Challenge     *challengeRequired `json:"challenge"`
	StepName      string             `json:"step_name"`
	ChallengeStep *challengeStep     `json:"step_data"`
}

type challengeRequired struct {
	URL  string `json:"url"`
	Path string `json:"api_path"`
}

type challengeStep struct {
	Choice string `json:"choice"` // 0=SMS, 1=Email
}
