package shared

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var SnsService snsService

type snsService struct {
	client   *sns.Client
	topicArn string
}

func (s *snsService) Init() {

	topicArn := os.Getenv("CONTENT_MODERATION_TOPIC_ARN")
	if topicArn == "" {
		panic("topic arn is empty")
	}

	SnsService.topicArn = topicArn

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("failed to load configs error, " + err.Error())
	}

	SnsService.client = sns.NewFromConfig(cfg)
}

func (s *snsService) Publish(message string) error {

	input := &sns.PublishInput{
		Message:  &message,
		TopicArn: &s.topicArn,
	}

	_, err := s.client.Publish(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}
