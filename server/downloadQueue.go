package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rokeller/zt-dl/ffmpeg"
	"github.com/rokeller/zt-dl/zattoo"
)

type toDownload struct {
	RecordingId int64  `json:"recordingId"`
	OutputPath  string `json:"filename"`
}

type downloadQueue struct {
	a  *zattoo.Account
	mu sync.Mutex
	q  []toDownload
}

func NewDownloadQueue(a *zattoo.Account) *downloadQueue {
	return &downloadQueue{
		a:  a,
		mu: sync.Mutex{},
		q:  []toDownload{},
	}
}

func (q *downloadQueue) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	downloading := false
	var downloadDone chan struct{}

	for {
		if !downloading {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				downloading, downloadDone = q.checkForDownloads()
			}
		} else {
			select {
			case <-ctx.Done():
				return
			case <-downloadDone:
				ticker.Reset(time.Second)
				downloading = false
			}
		}
	}
}

func (q *downloadQueue) checkForDownloads() (bool, chan struct{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.q) > 0 {
		dl, remaining := q.q[0], q.q[1:]
		q.q = remaining
		events <- event{DownloadStarted: &eventDownloadStarted{Filename: dl.OutputPath}}
		events <- event{QueueUpdated: &eventQueueUpdated{Queue: q.q}}

		downloadDone := make(chan struct{})
		go q.downloadRecording(dl, downloadDone)

		return true, downloadDone
	}

	return false, nil
}

func (q *downloadQueue) downloadRecording(r toDownload, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	events <- event{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}}
	url, err := q.a.GetRecordingStreamUrl(r.RecordingId)
	if nil != err {
		events <- event{DownloadErrored: &eventDownloadErrored{Reason: err.Error()}}
		fmt.Fprintf(os.Stderr, "failed to get recording stream: %v\n", err)
		return
	}

	d := ffmpeg.NewDownloadable(url, r.OutputPath)
	events <- event{StateUpdated: &eventStateUpdated{State: "detect_streams", Reason: "detecting recording audio and video streams ..."}}
	fmt.Println("Detecting streams ...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := d.DetectStreams(ctx); nil != err {
		events <- event{DownloadErrored: &eventDownloadErrored{Reason: err.Error()}}
		fmt.Fprintf(os.Stderr, "failed to get detect recording streams: %v\n", err)
		return
	}

	events <- event{StateUpdated: &eventStateUpdated{State: "download", Reason: "starting download ..."}}
	if err := d.Download(context.Background(), &broadcastDownloadProgressHandler{
		eventQueueUpdated: eventQueueUpdated{
			Queue: q.q,
		},
	}); nil != err {
		events <- event{DownloadErrored: &eventDownloadErrored{Reason: err.Error()}}
		fmt.Fprintf(os.Stderr, "failed to download recording: %v\n", err)
	}
}

func (q *downloadQueue) Enqueue(recordingId int64, outputPath string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.q = append(q.q, toDownload{recordingId, outputPath})
	events <- event{QueueUpdated: &eventQueueUpdated{Queue: q.q}}
}

type broadcastDownloadProgressHandler struct {
	eventQueueUpdated
}

// Start implements ffmpeg.DownloadProgressHandler.
func (b *broadcastDownloadProgressHandler) Start() {
	fmt.Printf("web download started\n")
}

// Error implements ffmpeg.DownloadProgressHandler.
func (b *broadcastDownloadProgressHandler) Error(err error) {
	fmt.Fprintf(os.Stderr, "web download failed: %v\n", err)
}

// Finished implements ffmpeg.DownloadProgressHandler.
func (b *broadcastDownloadProgressHandler) Finished() {
	fmt.Printf("web download finished\n")
}

// UpdateProgress implements ffmpeg.DownloadProgressHandler.
func (b *broadcastDownloadProgressHandler) UpdateProgress(p ffmpeg.DownloadProgress) {
	fmt.Printf("web download progress: %5.1f%% | Elapsed: %10s | Remaining: %10s\r",
		p.RelCompleted*100, p.Elapsed.Truncate(time.Second), p.Remaining)
	events <- event{ProgressUpdated: &eventProgressUpdated{
		RelCompleted: p.RelCompleted,
		Elapsed:      p.Elapsed.Truncate(time.Second).String(),
		Remaining:    p.Remaining.String(),
	}}
}
