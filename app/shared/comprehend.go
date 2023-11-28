package shared

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
)

var ComprehendService comprehendService

type comprehendService struct {
	client *comprehend.Client
}

func (c *comprehendService) Init() {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("failed to load configs error, " + err.Error())
	}

	ComprehendService.client = comprehend.NewFromConfig(cfg)
}

func (c *comprehendService) DetectToxicContent(text string) (*comprehend.DetectToxicContentOutput, error) {

	langInput := &comprehend.DetectDominantLanguageInput{
		Text: aws.String(text),
	}

	langResult, err := c.client.DetectDominantLanguage(context.Background(), langInput)
	if err != nil {
		return nil, fmt.Errorf("error detecting dominant language: %w", err)
	}

	if len(langResult.Languages) == 0 || langResult.Languages[0].LanguageCode == nil {
		return nil, fmt.Errorf("no dominant language detected")
	}

	dominantLanguage := *langResult.Languages[0].LanguageCode

	toxicContentInput := &comprehend.DetectToxicContentInput{
		TextSegments: []types.TextSegment{{
			Text: aws.String(text),
		},
		},
		LanguageCode: types.LanguageCode(dominantLanguage),
	}

	detectToxicContentOutput, err := c.client.DetectToxicContent(context.Background(), toxicContentInput)
	if err != nil {
		return nil, fmt.Errorf("error detecting toxic content: %w", err)
	}

	return detectToxicContentOutput, nil
}
