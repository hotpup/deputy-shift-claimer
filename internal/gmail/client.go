package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Client wraps the Gmail API client
type Client struct {
	service *gmail.Service
}

// Message represents a simplified email message
type Message struct {
	ID      string
	Subject string
	Body    string
	From    string
	Date    string
}

// NewClient creates a new Gmail client
func NewClient() (*Client, error) {
	ctx := context.Background()

	// Read credentials from credentials.json
	credBytes, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(credBytes, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	client, err := getClient(config)
	if err != nil {
		return nil, err
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %w", err)
	}

	return &Client{service: srv}, nil
}

// getClient retrieves a token, saves the token, then returns the generated client
func getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokFile, tok); err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
}

// getTokenFromWeb requests a token from the web
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("unable to encode token: %w", err)
	}
	return nil
}

// FetchDeputyEmails fetches emails from the Deputy label
func (c *Client) FetchDeputyEmails() ([]*Message, error) {
	user := "me"

	// Search for messages with the Deputy label
	query := "label:Deputy"

	listCall := c.service.Users.Messages.List(user).Q(query).MaxResults(50)
	msgList, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %w", err)
	}

	var messages []*Message
	for _, m := range msgList.Messages {
		msg, err := c.service.Users.Messages.Get(user, m.Id).Format("full").Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve message %s: %w", m.Id, err)
		}

		message := &Message{
			ID: msg.Id,
		}

		// Extract headers
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "Subject":
				message.Subject = header.Value
			case "From":
				message.From = header.Value
			case "Date":
				message.Date = header.Value
			}
		}

		// Extract body
		body := getMessageBody(msg.Payload)
		message.Body = body

		messages = append(messages, message)
	}

	return messages, nil
}

// getMessageBody extracts the body from a message payload
func getMessageBody(payload *gmail.MessagePart) string {
	if payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	// Check for multipart messages
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			if part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
		// Recursively check nested parts
		if len(part.Parts) > 0 {
			body := getMessageBody(part)
			if body != "" {
				return body
			}
		}
	}

	return ""
}
