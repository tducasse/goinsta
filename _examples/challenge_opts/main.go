package main

import (
	"fmt"
	"log"

	"github.com/abramovic/goinsta"
)

func main() {
	// we will not pass any options
	insta := goinsta.New("USERNAME", "PASSWORD", nil)

	err := insta.Login()
	fatalOnErr(err)

	defer insta.Logout()

	instaPage, err := insta.GetUserByUsername("thedodo")
	fatalOnErr(err)

	fmt.Printf("%+v\n", instaPage)
}

func fatalOnErr(err error) {
	if err == nil {
		// do nothing
		return
	}
	switch err {
	case goinsta.ErrChallengeOptionsRequired:
		log.Fatal("we did not pass any options to goinsta.New() so we could not send a challenge code ")
	case goinsta.ErrChallengeCodeRequired:
		log.Fatal("a challenge is required. please check your phone/email")
	case goinsta.ErrChallengeCodeInvalid:
		log.Fatal("The challenge code provided above is invalid")
	default:
		log.Fatal(err)
	}
}
