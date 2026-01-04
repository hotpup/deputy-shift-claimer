package main

import (
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func connectIMAP(server, email, password string) (*client.Client, error) {
	c, err := client.DialTLS(server, nil)
	if err != nil {
		return nil, err
	}

	if err := c.Login(email, password); err != nil {
		return nil, err
	}

	log.Println("Connected to IMAP")
	return c, nil
}

func searchDeputyEmails(c *client.Client) ([]uint32, error) {
	_, err := c.Select("INBOX", false)
	if err != nil {
		return nil, err
	}

	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("From", "deputy")

	return c.Search(criteria)
}
