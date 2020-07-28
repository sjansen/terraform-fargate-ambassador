package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

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
		logger = NewLogger(3)
		for _, kv := range os.Environ() {
			fmt.Println(kv)
		}
	} else {
		logger = NewLogger(2)
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
		logger.Infow("Shutdown signal received.",
			"signal", sig.String(),
		)
		cancel()
		sigs = nil
	}()

	a, err := NewAmbassador(ctx, queue)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger.Info("Startup complete.")
	for sigs != nil {
		msgs, err := a.ReceiveMessages()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else if len(msgs) > 0 {
			logger.Infow("New message(s) received.",
				"count", len(msgs),
			)
			for _, msg := range msgs {
				fmt.Println(msg.Body)
				a.DeleteMessage(msg.Handle)
				time.Sleep(1 * time.Second)
			}
		}
	}
	logger.Info("Shutdown complete.")
}

// NewLogger returns a logger
//
// Valid levels are:
//   0 = errors only,
//   1 = include warnings,
//   2 = include informational messages,
//   3 = include debug messages.
func NewLogger(verbosity int) *zap.SugaredLogger {
	var level zapcore.Level
	switch {
	case verbosity >= 3:
		level = zapcore.DebugLevel
	case verbosity == 2:
		level = zapcore.InfoLevel
	case verbosity == 1:
		level = zapcore.WarnLevel
	default:
		level = zapcore.ErrorLevel
	}

	var stdout io.Writer = os.Stdout
	encoder := zapcore.CapitalLevelEncoder
	if x, ok := stdout.(interface{ Fd() uintptr }); ok {
		if isatty.IsTerminal(x.Fd()) {
			encoder = zapcore.CapitalColorLevelEncoder
		}
	}
	cfg := zapcore.EncoderConfig{
		LevelKey:       "level",
		MessageKey:     "msg",
		NameKey:        "logger",
		TimeKey:        "timestamp",
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeLevel:    encoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(stdout),
		level,
	)

	return zap.New(core).Sugar()
}
