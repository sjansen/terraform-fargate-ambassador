package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var startTime time.Time

var CLI struct {
	CheckHealth CheckHealth `cmd help:"Request status from ambassador server."`
	Server      Server      `cmd help:"Run ambassador in server mode."`
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
	startTime = time.Now()

	cfg, err := NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if cfg.Debug {
		logger = NewLogger(3)
		for _, kv := range os.Environ() {
			fmt.Println(kv)
		}
	} else {
		logger = NewLogger(2)
	}

	cli := kong.Parse(&CLI)
	err = cli.Run(cfg)
	cli.FatalIfErrorf(err)
}
