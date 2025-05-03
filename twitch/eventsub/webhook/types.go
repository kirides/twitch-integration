package webhook

type callbackVerificationData struct {
	Message struct {
		ID        string
		Retry     int
		Type      string
		Signature string
		Timestamp string
	}
	Subscription struct {
		Type    string
		Version string
	}
}
