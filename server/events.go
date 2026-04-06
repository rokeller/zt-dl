package server

// serverEvent defines the root object for an event sent by the server.
type serverEvent struct {
	Correlation              string                         `json:"correlation,omitempty"`
	QueueUpdated             *eventQueueUpdated             `json:"queueUpdated,omitempty"`
	DownloadStarted          *eventDownloadStarted          `json:"downloadStarted,omitempty"`
	ProgressUpdated          *eventProgressUpdated          `json:"progressUpdated,omitempty"`
	DownloadErrored          *eventDownloadErrored          `json:"downloadErrored,omitempty"`
	StateUpdated             *eventStateUpdated             `json:"stateUpdated,omitempty"`
	StreamSelectionRequested *eventStreamSelectionRequested `json:"selectStreams,omitempty"`
}

type clientEvent struct {
	Correlation     string                `json:"correlation,omitempty"`
	StreamsSelected *eventStreamsSelected `json:"streamsSelected,omitempty"`
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

type eventStreamSelectionRequested struct {
	SourceStreams []sourceStream `json:"streams"`
}

type eventStreamsSelected struct {
	SelectedStreams []sourceStream `json:"streams"`
}

type sourceStream struct {
	Index       int    `json:"index"`
	Type        string `json:"type"`
	Description string `json:"desc"`
}
