package executor

import (
	"context"
	"fmt"
	"github.com/GLCharge/distributed-scheduler/model"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"net/http"
	"strings"
)

type hTTPExecutor struct {
	Client HttpClient
}

type aMQPExecutor struct{}

// HttpClient interface
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPSPrefix and HTTPPrefix are prefixes for HTTP and HTTPS protocols
const (
	HTTPSPrefix = "https://"
	HTTPPrefix  = "http://"
)

func (he *hTTPExecutor) Execute(ctx context.Context, j *model.Job) error {
	// Create the HTTP request
	req, err := he.createHTTPRequest(ctx, j)
	if err != nil {
		return err
	}

	// Send the request and get the response
	resp, err := he.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if status code is one of the valid response codes
	if !he.validResponseCode(resp.StatusCode, j.HTTPJob.ValidResponseCodes) {
		return model.ErrInvalidResponseCode
	}

	return nil
}

func (he *hTTPExecutor) validResponseCode(code int, validCodes []int) bool {
	// If no valid response codes are defined, 200 is the default
	if len(validCodes) == 0 {
		return code == http.StatusOK
	}

	// Check if the response code is one of the valid response codes
	for _, c := range validCodes {
		if c == code {
			return true
		}
	}

	return false
}

func (he *hTTPExecutor) createHTTPRequest(ctx context.Context, j *model.Job) (*http.Request, error) {
	// Create the request body
	body := he.createHTTPRequestBody(j.HTTPJob.Body.String)

	// Create the request URL
	url := he.createHTTPRequestURL(j.HTTPJob.URL)

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, j.HTTPJob.Method, url, body)
	if err != nil {
		return nil, err
	}

	// Set the headers
	he.setHTTPRequestHeaders(req, j.HTTPJob.Headers)

	// Set the auth
	he.setHTTPRequestAuth(req, j.HTTPJob.Auth)

	return req, nil
}

func (he *hTTPExecutor) createHTTPRequestBody(body string) io.Reader {
	if body == "" {
		return nil
	}

	return strings.NewReader(body)
}

func (he *hTTPExecutor) createHTTPRequestURL(url string) string {
	if strings.HasPrefix(url, HTTPPrefix) || strings.HasPrefix(url, HTTPSPrefix) {
		return url
	}

	return HTTPSPrefix + url
}

func (he *hTTPExecutor) setHTTPRequestHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func (he *hTTPExecutor) setHTTPRequestAuth(req *http.Request, auth model.Auth) {
	switch auth.Type {
	case model.AuthTypeBasic:
		req.SetBasicAuth(auth.Username.String, auth.Password.String)
	case model.AuthTypeBearer:
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.BearerToken.String))
	}
}

func (ae *aMQPExecutor) Execute(ctx context.Context, j *model.Job) error {
	// Create a new AMQP connection
	conn, err := amqp.Dial(j.AMQPJob.Connection)
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP: %w", err)
	}
	defer conn.Close()

	// Create a new AMQP channel
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Declare an exchange
	err = ch.ExchangeDeclare(
		j.AMQPJob.Exchange,     // name
		j.AMQPJob.ExchangeType, // type
		j.AMQPJob.Durable,      // durable
		j.AMQPJob.AutoDelete,   // auto-deleted
		j.AMQPJob.Internal,     // internal
		j.AMQPJob.NoWait,       // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %w", err)
	}

	// Publish a message to the exchange
	err = ch.PublishWithContext(
		ctx,
		j.AMQPJob.Exchange,   // exchange
		j.AMQPJob.RoutingKey, // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: j.AMQPJob.ContentType,
			Body:        []byte(j.AMQPJob.Body),
			Headers:     j.AMQPJob.Headers,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
