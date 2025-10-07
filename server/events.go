package server

type event struct {
	QueueUpdated    *eventQueueUpdated    `json:"queueUpdated,omitempty"`
	DownloadStarted *eventDownloadStarted `json:"downloadStarted,omitempty"`
	ProgressUpdated *eventProgressUpdated `json:"progressUpdated,omitempty"`
	StateUpdated    *eventStateUpdated    `json:"stateUpdated,omitempty"`
	DownloadErrored *eventDownloadErrored `json:"downloadErrored,omitempty"`
}

type eventQueueUpdated struct {
	Queue []toDownload `json:"queue"`
}

type eventDownloadStarted struct {
	Filename string `json:"filename"`
}

type eventProgressUpdated struct {
	RelCompleted float32 `json:"completed"`
	Elapsed      string  `json:"elapsed"`
	Remaining    string  `json:"remaining"`
}

type eventDownloadErrored struct {
	Filename string `json:"filename"`
	Reason   string `json:"reason"`
}

type eventStateUpdated struct {
	State  string `json:"state"`
	Reason string `json:"reason"`
}
