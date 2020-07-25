package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog/log"
)

type Ambassador struct {
	ctx   aws.Context
	queue *string
	sqs   *sqs.SQS
}

type Message struct {
	Body   string
	Handle string
}

func NewAmbassador(ctx aws.Context, queue string) (*Ambassador, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	queueURL, err := getQueueURL(sess, queue)
	if err != nil {
		return nil, err
	}

	return &Ambassador{
		ctx:   ctx,
		queue: queueURL,
		sqs:   sqs.New(sess),
	}, nil
}

func (a *Ambassador) DeleteMessage(handle string) error {
	_, err := a.sqs.DeleteMessageWithContext(a.ctx, &sqs.DeleteMessageInput{
		QueueUrl:      a.queue,
		ReceiptHandle: aws.String(handle),
	})
	return err
}

func (a *Ambassador) ReceiveMessages() ([]Message, error) {
	result, err := a.sqs.ReceiveMessageWithContext(a.ctx, &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            a.queue,
		MaxNumberOfMessages: aws.Int64(1),
	})
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			if err.Code() == request.CanceledErrorCode {
				return nil, nil
			}
		}
		return nil, err
	}

	messages := make([]Message, 0, len(result.Messages))
	for _, msg := range result.Messages {
		messages = append(messages, Message{
			Body:   aws.StringValue(msg.Body),
			Handle: aws.StringValue(msg.ReceiptHandle),
		})
	}
	return messages, nil
}

func getQueueURL(sess *session.Session, queue string) (*string, error) {
	svc := sqs.New(sess)

	urlResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queue),
	})
	if err != nil {
		return nil, err
	}

	return urlResult.QueueUrl, nil
}

func main() {
	debug := os.Getenv("DEBUG") != ""
	if debug {
		for _, kv := range os.Environ() {
			fmt.Println(kv)
		}
	}

	queue := os.Getenv("QUEUE")
	if queue == "" {
		fmt.Fprintln(os.Stderr, "Required setting missing: $QUEUE")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received.")
		cancel()
		sigs = nil
	}()

	a, err := NewAmbassador(ctx, queue)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	log.Info().Msg("Startup complete.")
	for sigs != nil {
		msgs, err := a.ReceiveMessages()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else if len(msgs) > 0 {
			log.Info().
				Int("count", len(msgs)).
				Msg("New message(s) received.")
			for _, msg := range msgs {
				fmt.Println(msg.Body)
				a.DeleteMessage(msg.Handle)
			}
		}
	}
	log.Info().Msg("Shutdown complete.")
}
