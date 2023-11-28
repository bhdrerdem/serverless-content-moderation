package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	types "github.com/bhdrerdem/serverless-content-moderation"
	"github.com/bhdrerdem/serverless-content-moderation/shared"
	"github.com/rs/zerolog/log"
)

func main() {
	shared.DBService.Init()
	shared.ComprehendService.Init()
	lambda.Start(consume)
}

func consume(_ context.Context, sqsEvent events.SQSEvent) error {
	log.Debug().Int("length", len(sqsEvent.Records)).Msg("Got sqs event")

	if len(sqsEvent.Records) == 0 {
		log.Error().Interface("sqsEvent", sqsEvent).Msg("No SQS message passed to function")
		return nil
	}

	var wg sync.WaitGroup
	ch := make(chan error, len(sqsEvent.Records))
	for _, msg := range sqsEvent.Records {
		wg.Add(1)
		go processMessage(&wg, ch, msg)
	}

	go func() {
		log.Debug().Msg("Waiting for process messages routine")
		wg.Wait()
		close(ch)
		log.Debug().Msg("Process messages routine is done")
	}()

	for e := range ch {
		log.Error().Err(e).Msg("Processing error")
	}

	return nil
}

func processMessage(
	wg *sync.WaitGroup,
	ch chan error,
	msg events.SQSMessage) {

	defer wg.Done()
	log.Debug().
		Str("messageId", msg.MessageId).
		Str("body", msg.Body).
		Msg("Got SQS message")

	var body struct {
		Message string `json:"Message"`
	}

	if err := json.Unmarshal([]byte(msg.Body), &body); err != nil {
		log.Error().Err(err).
			Interface("body", msg.Body).
			Msg("Failed to unmarshal body json")
		ch <- err
		return
	}

	content := &types.Content{}
	if err := json.Unmarshal([]byte(body.Message), &content); err != nil {
		log.Error().Err(err).
			Interface("content", body.Message).
			Msg("Failed to unmarshal content json")
		ch <- err
		return
	}

	if content.ID == "" || content.Text == "" {
		log.Error().Interface("content", content).Msg("One of the required field is empty")
		ch <- errors.New("one of reqired field is empty")
		return
	}

	log.Info().
		Interface("content", content).
		Msg("Process content")

	handleContentAnalysis(content)
}

func handleContentAnalysis(content *types.Content) {

	toxicContentOutput, err := shared.ComprehendService.DetectToxicContent(content.Text)
	if err != nil {
		content.Status = types.StatusReviewRequired
		log.Error().Err(err).
			Interface("content", content).
			Msg("Failed to detect toxic content")
	} else {

		result := toxicContentOutput.ResultList[0]
		if *result.Toxicity >= 0.3 {
			content.Status = types.StatusDetected

			for _, label := range result.Labels {
				content.ToxicLabels = append(content.ToxicLabels, string(label.Name))
			}
		} else {
			content.Status = types.StatusNone
		}
	}

	err = shared.DBService.PutItem(content)
	if err != nil {
		log.Error().Err(err).
			Interface("content", content).
			Msg("Failed to persist content data")
	}
}
