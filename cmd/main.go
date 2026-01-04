package main

import (
	"io"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
)

func main() {
	imapServer := "imap.gmail.com:993"
	email := "you@gmail.com"
	password := "APP_PASSWORD"

	c, err := connectIMAP(imapServer, email, password)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	ids, err := searchDeputyEmails(c)
	if err != nil {
		log.Fatal(err)
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)

	go func() {
		_ = c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
	}()

	for msg := range messages {
		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		m, err := mail.ReadMessage(r)
		if err != nil {
			continue
		}

		bodyBytes, _ := io.ReadAll(m.Body)
		body := string(bodyBytes)

		shift, ok := parseShift(body)
		if !ok {
			continue
		}

		if shiftMatches(shift) {
			triggerAlert(shift)
		}
	}
}
