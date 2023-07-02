package ntfy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type NtfyClient interface {
	SendNotification(msg string) error
}

type ntfyClient struct {
	httpClient http.Client
	url        string
}

func New(url string, timeout time.Duration) NtfyClient {
	return &ntfyClient{
		url:        url,
		httpClient: http.Client{Timeout: timeout},
	}
}

func (n *ntfyClient) SendNotification(msg string) error {
	res, err := n.httpClient.Post(n.url, "text/plain", strings.NewReader(msg))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read body: %v", err)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("status: %d, body: %s", res.StatusCode, string(body))
	}
	return nil

}

type Mock struct {
	msgs []string
}

func (n *Mock) SendNotification(msg string) error {

	n.msgs = append(n.msgs, msg)
	return nil
}

func (n *Mock) Messages() []string {
	return n.msgs
}
