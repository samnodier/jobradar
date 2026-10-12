package matcher

import "testing"

func TestFeedbackLearning(t *testing.T) {
	good := FeedbackExample{Title: "Go Developer", Skills: []string{"Go"}, Sentiment: 1}
	bad := good
	bad.Sentiment = -1
	if got := FeedbackAdjustment("Go Developer", []string{"Go"}, nil); got != 0 {
		t.Fatal("cold start", got)
	}
	if got := FeedbackAdjustment("Go Developer", []string{"Go"}, []FeedbackExample{good}); got <= 0 || got > 3 {
		t.Fatal("single vote must be shrunk", got)
	}
	if got := FeedbackAdjustment("Go Developer", []string{"Go"}, []FeedbackExample{good, bad}); got != 0 {
		t.Fatal("conflicting votes", got)
	}
	if got := FeedbackAdjustment("Nurse", nil, []FeedbackExample{good}); got != 0 {
		t.Fatal("unrelated role", got)
	}
	examples := make([]FeedbackExample, 200)
	for i := range examples {
		examples[i] = bad
	}
	if got := FeedbackAdjustment("Go Developer", []string{"Go"}, examples); got < -10 || got >= 0 {
		t.Fatal("bounded negative adjustment", got)
	}
}
