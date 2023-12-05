package types

type Content struct {
	ID          string   `json:"id"`
	Text        string   `json:"text"`
	Status      Status   `json:"status"`
	ToxicLabels []string `json:"toxicLabels,omitempty"`
}

type Status string

const (
	StatusNone           Status = "none"
	StatusDetected       Status = "detected"
	StatusReviewRequired Status = "review_required"
)
