package goinsta

import (
	"net/http"
	"net/http/cookiejar"

	response "github.com/tducasse/goinsta/response"
)

type Informations struct {
	Username  string
	Password  string
	DeviceID  string
	UUID      string
	RankToken string
	Token     string
	PhoneID   string
}

type Instagram struct {
	Cookiejar *cookiejar.Jar
	InstaType
	Transport        http.Transport
	challengePath    string
	challengeOptions *Challenge
	reqOptions       *reqOptions
}

func (insta *Instagram) canChallenge() bool {
	if insta.challengeOptions == nil {
		return false
	}
	return true
}

type ChallengeDelivery string

type Challenge struct {
	Delivery ChallengeDelivery
	Code     string
}

func (c *Challenge) Choice() ChallengeDelivery {
	switch c.Delivery {
	// use the delivery method provided
	case GOINSTA_CHALLENGE_SMS, GOINSTA_CHALLENGE_EMAIL:
		return c.Delivery
	}
	// by default we will use email
	return GOINSTA_CHALLENGE_EMAIL
}

type InstaType struct {
	IsLoggedIn   bool
	Informations Informations
	LoggedInUser response.User

	Proxy string
}

type BackupType struct {
	Cookies []http.Cookie
	InstaType
}
