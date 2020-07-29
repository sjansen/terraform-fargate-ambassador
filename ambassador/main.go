package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

type Message struct {
	Body   string
	Handle string
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

	logger.Info("Startup initiated.")
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	go HandleSignals(ctx, cancel, wg)
	wg.Add(1)

	go MonitorDiskUsage(ctx, wg)
	wg.Add(1)

	srv := NewServer()
	go WaitForShutdown(ctx, srv, wg)
	wg.Add(1)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("ListenAndServe: %v", err)
			cancel()
		}
		wg.Done()
	}()
	wg.Add(1)

	a, err := NewAmbassador(ctx, cfg)
	if err != nil {
		logger.Fatalw("Failed to create Ambassador.",
			"error", err,
		)
	}
	logger.Info("Startup complete.")

OuterLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutdown initiated.")
			break OuterLoop
		default:
			msgs, err := a.ReceiveMessages()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else if len(msgs) > 0 {
				logger.Infow("Message(s) received.",
					"count", len(msgs),
				)
				for _, msg := range msgs {
					http.PostForm(cfg.LinkURL+"/echo",
						url.Values{"msg": {msg.Body}},
					)
					a.DeleteMessage(msg.Handle)
					time.Sleep(1 * time.Second)
				}
			}
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete.")
}
