package messages

type AvailabilityStatus struct {
	AvailableWorkerCount int `json:"available_worker_count"`
}

type InitializationStatus struct {
	Success bool `json:"success"`
}

type Job struct {
	CallbackURL string `json:"callback_url"`
	MediaURL    string `json:"media_url"`
	Metadata    string `json:"metadata"`
}

type Result struct {
	Failure       string `json:"failure"`
	FailureDetail string `json:"failure_detail"`
	Message       string `json:"message"`
	Metadata      string `json:"metadata"`
}
