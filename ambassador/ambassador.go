package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
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

func NewAmbassador(ctx aws.Context, cfg *Config) (*Ambassador, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	queueURL, err := getQueueURL(sess, cfg.Queue)
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
		logger.Debugw("ReceiveMessages failed.",
			"error", err,
		)
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
