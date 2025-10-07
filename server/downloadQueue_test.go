package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	e "github.com/rokeller/zt-dl/exec"
	"github.com/rokeller/zt-dl/ffmpeg"
	"github.com/rokeller/zt-dl/test"
	"github.com/rokeller/zt-dl/zattoo"
)

func Test_downloadQueue_Run(t *testing.T) {
	tests := []struct {
		name       string
		q          []toDownload
		ctxFactory func(parent context.Context) (context.Context, context.CancelFunc)
		wantEvents []event
	}{
		{
			name: "ContextCancelled/RightAway",
			ctxFactory: func(parent context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(parent)
				cancel()
				return ctx, func() {}
			},
		},
		{
			name: "ContextCancelled/After>1sec",
			ctxFactory: func(parent context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(parent, time.Second+time.Millisecond*10)
				return ctx, cancel
			},
		},
		{
			name: "QueueItemHandling/WithTimeout",
			q: []toDownload{
				{RecordingId: 1234, OutputPath: "/tmp/foo/bar"},
			},
			ctxFactory: func(parent context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(parent, time.Second+time.Millisecond*10)
				return ctx, cancel
			},
			wantEvents: []event{
				{DownloadStarted: &eventDownloadStarted{Filename: "/tmp/foo/bar"}},
				{QueueUpdated: &eventQueueUpdated{Queue: []toDownload{}}},
				{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}},
				{StateUpdated: &eventStateUpdated{State: "detect_streams", Reason: "detecting recording audio and video streams ..."}},
				{DownloadErrored: &eventDownloadErrored{Filename: "/tmp/foo/bar", Reason: "failed to run ffprobe: exit status 1"}},
			},
		},
		{
			name: "QueueItemHandling/WithLongTimeout",
			q: []toDownload{
				{RecordingId: 2345, OutputPath: "/dev/null"},
			},
			ctxFactory: func(parent context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(parent, time.Second*2)
				return ctx, cancel
			},
			wantEvents: []event{
				{DownloadStarted: &eventDownloadStarted{Filename: "/dev/null"}},
				{QueueUpdated: &eventQueueUpdated{Queue: []toDownload{}}},
				{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}},
				{StateUpdated: &eventStateUpdated{State: "detect_streams", Reason: "detecting recording audio and video streams ..."}},
				{DownloadErrored: &eventDownloadErrored{Filename: "/dev/null", Reason: "failed to run ffprobe: exit status 1"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				if len(tt.q) > 0 &&
					r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", tt.q[0].RecordingId) &&
					r.Method == http.MethodPost {
					test.HttpResponse{
						StatusCode: 200,
						Body:       []byte(`{"success":true,"stream":{"url":"http://localhost:8888/blah/blotz"}}`),
					}.Respond(w)
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := zattoo.NewAccountWithSession(t, host, client)
			s := &server{
				a:   a,
				hub: newHub(),
			}
			q := &downloadQueue{
				server: s,
				mu:     sync.Mutex{},
				q:      tt.q,
			}
			ctx, cancel := tt.ctxFactory(t.Context())
			defer cancel()
			q.Run(ctx)

			consumeEvents(t, s.hub.outbox, tt.wantEvents)
			ensureNoMoreEvents(t, s.hub.outbox)
		})
	}
}

func Test_downloadQueue_checkForDownloads(t *testing.T) {
	tests := []struct {
		name       string
		q          []toDownload
		resp       test.HttpResponse
		want       bool
		wantCh     bool
		wantEvents []event
	}{
		{
			name:   "EmptyQueue",
			q:      []toDownload{},
			want:   false,
			wantCh: false,
		},
		{
			name: "QueueWith/1-item",
			q: []toDownload{
				{RecordingId: 1234, OutputPath: "/tmp/blah.mp4"},
			},
			resp:   test.HttpResponse{StatusCode: 404},
			want:   true,
			wantCh: true,
			wantEvents: []event{
				{DownloadStarted: &eventDownloadStarted{Filename: "/tmp/blah.mp4"}},
				{QueueUpdated: &eventQueueUpdated{Queue: []toDownload{}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				if len(tt.q) > 0 &&
					r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", tt.q[0].RecordingId) &&
					r.Method == http.MethodPost {
					tt.resp.Respond(w)
					return
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := zattoo.NewAccountWithSession(t, host, client)
			s := &server{
				a:   a,
				hub: newHub(),
			}
			q := &downloadQueue{
				server: s,
				mu:     sync.Mutex{},
				q:      tt.q,
			}
			got, gotCh := q.checkForDownloads()
			if got != tt.want {
				t.Errorf("downloadQueue.checkForDownloads() got = %v, want %v", got, tt.want)
			}
			if gotCh != nil {
				defer func() {
					<-gotCh
				}()
				if !tt.wantCh {
					t.Errorf("downloadQueue.checkForDownloads() gotCh = %v, but want none", gotCh)
				}
			} else if tt.wantCh {
				t.Error("downloadQueue.checkForDownloads() wantCh, but got nil")
			}

			consumeEvents(t, s.hub.outbox, tt.wantEvents)
			ensureNoMoreEvents(t, s.hub.outbox)
		})
	}
}

func Test_downloadQueue_downloadRecording(t *testing.T) {
	tests := []struct {
		name          string
		r             toDownload
		respRecording test.HttpResponse
		wantEvents    []event
	}{
		{
			name:          "GetRecordingStreamUrlFails",
			r:             toDownload{RecordingId: 1234, OutputPath: "/tmp/GetRecordingStreamUrlFails"},
			respRecording: test.HttpResponse{StatusCode: 404},
			wantEvents: []event{
				{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}},
				{DownloadErrored: &eventDownloadErrored{Filename: "/tmp/GetRecordingStreamUrlFails", Reason: "failed to get recording with status 404"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", tt.r.RecordingId) &&
					r.Method == http.MethodPost {
					tt.respRecording.Respond(w)
					return
				}
				w.Header().Add("x-reason", "unsupported-uri")
				w.WriteHeader(404)
			})
			defer ts.Close()
			a := zattoo.NewAccountWithSession(t, host, client)
			s := &server{
				a:   a,
				hub: newHub(),
			}
			q := &downloadQueue{
				server: s,
				mu:     sync.Mutex{},
				q:      []toDownload{},
			}
			done := make(chan struct{}, 1)
			q.downloadRecording(tt.r, done)
			<-done

			consumeEvents(t, s.hub.outbox, tt.wantEvents)
			ensureNoMoreEvents(t, s.hub.outbox)
		})
	}
}

func Test_downloadQueue_downloadRecording_DetectStreamsFails(t *testing.T) {
	if test.IsTestCall() {
		test.AssertArgs(
			"ffprobe",
			"-protocol_whitelist", "https,tls,tcp",
			"-print_format", "json",
			"-show_format",
			"-show_streams",
			"-i", "https://Test_downloadQueue_downloadRecording_DetectStreamsFails",
		)
		os.Exit(2)
		return
	}

	me := test.CallerFuncName(0)
	e.CmdFactory = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return test.TestCommandContext(t, me, ctx, name, arg...)
	}

	ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", 1111) &&
			r.Method == http.MethodPost {
			test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"stream":{"url":"https://Test_downloadQueue_downloadRecording_DetectStreamsFails"}}`),
			}.Respond(w)
			return
		}
		w.Header().Add("x-reason", "unsupported-uri")
		w.WriteHeader(404)
	})
	defer ts.Close()
	a := zattoo.NewAccountWithSession(t, host, client)
	s := &server{
		a:   a,
		hub: newHub(),
	}
	q := &downloadQueue{
		server: s,
		mu:     sync.Mutex{},
		q:      []toDownload{},
	}
	done := make(chan struct{}, 1)
	q.downloadRecording(toDownload{RecordingId: 1111}, done)
	<-done

	consumeEvents(t, s.hub.outbox, []event{
		{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}},
		{StateUpdated: &eventStateUpdated{State: "detect_streams", Reason: "detecting recording audio and video streams ..."}},
		{DownloadErrored: &eventDownloadErrored{Reason: "failed to run ffprobe: exit status 2"}},
	})
	ensureNoMoreEvents(t, s.hub.outbox)
}

func Test_downloadQueue_downloadRecording_DownloadFails(t *testing.T) {
	if test.IsTestCall() {
		switch test.GetArgs()[0] {
		case "ffprobe":
			test.AssertArgs(
				"ffprobe",
				"-protocol_whitelist", "https,tls,tcp",
				"-print_format", "json",
				"-show_format",
				"-show_streams",
				"-i", "https://Test_downloadQueue_downloadRecording_DownloadFails",
			)
			fmt.Println(`{
	"format": { "duration": "123456.789" },
	"streams": [
	{
		"index": 0,
		"codec_type": "audio",
		"sample_rate": "44000"
	},
	{
		"index": 1,
		"codec_type": "audio",
		"sample_rate": "88000"
	},
	{
		"index": 2,
		"codec_type": "video",
		"width": 600,
		"height": 400,
		"avg_frame_rate": "50/1",
		"bit_rate": "1200"
	},
	{
		"index": 5,
		"codec_type": "video",
		"width": 1200,
		"height": 800,
		"avg_frame_rate": "30/1",
		"bit_rate": "3456000"
	}
]}`)
			os.Exit(0)
		case "ffmpeg":
			os.Exit(1)
		}
		return
	}

	me := test.CallerFuncName(0)
	e.CmdFactory = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return test.TestCommandContext(t, me, ctx, name, arg...)
	}

	ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", 1111) &&
			r.Method == http.MethodPost {
			test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"stream":{"url":"https://Test_downloadQueue_downloadRecording_DownloadFails"}}`),
			}.Respond(w)
			return
		}
		w.Header().Add("x-reason", "unsupported-uri")
		w.WriteHeader(404)
	})
	defer ts.Close()
	a := zattoo.NewAccountWithSession(t, host, client)
	s := &server{
		a:   a,
		hub: newHub(),
	}
	q := &downloadQueue{
		server: s,
		mu:     sync.Mutex{},
		q:      []toDownload{},
	}
	done := make(chan struct{}, 1)
	q.downloadRecording(toDownload{RecordingId: 1111}, done)
	<-done

	consumeEvents(t, s.hub.outbox, []event{
		{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}},
		{StateUpdated: &eventStateUpdated{State: "detect_streams", Reason: "detecting recording audio and video streams ..."}},
		{StateUpdated: &eventStateUpdated{State: "download", Reason: "starting download ..."}},
		{DownloadErrored: &eventDownloadErrored{Reason: "ffmpeg failed: exit status 1"}},
	})
	ensureNoMoreEvents(t, s.hub.outbox)
}

func Test_downloadQueue_downloadRecording_DownloadSucceeds(t *testing.T) {
	if test.IsTestCall() {
		switch test.GetArgs()[0] {
		case "ffprobe":
			test.AssertArgs(
				"ffprobe",
				"-protocol_whitelist", "https,tls,tcp",
				"-print_format", "json",
				"-show_format",
				"-show_streams",
				"-i", "https://Test_downloadQueue_downloadRecording_DownloadSucceeds",
			)
			fmt.Println(`{
	"format": { "duration": "123456.789" },
	"streams": [
	{
		"index": 0,
		"codec_type": "audio",
		"sample_rate": "44000"
	},
	{
		"index": 1,
		"codec_type": "audio",
		"sample_rate": "88000"
	},
	{
		"index": 2,
		"codec_type": "video",
		"width": 600,
		"height": 400,
		"avg_frame_rate": "50/1",
		"bit_rate": "1200"
	},
	{
		"index": 5,
		"codec_type": "video",
		"width": 1200,
		"height": 800,
		"avg_frame_rate": "30/1",
		"bit_rate": "3456000"
	}
]}`)
			os.Exit(0)
		case "ffmpeg":
			os.Exit(0)
		}
		return
	}

	me := test.CallerFuncName(0)
	e.CmdFactory = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return test.TestCommandContext(t, me, ctx, name, arg...)
	}

	ts, client, host := test.NewHttpTestSetup(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == fmt.Sprintf("/zapi/watch/recording/%d", 1111) &&
			r.Method == http.MethodPost {
			test.HttpResponse{
				StatusCode: 200,
				Body:       []byte(`{"success":true,"stream":{"url":"https://Test_downloadQueue_downloadRecording_DownloadSucceeds"}}`),
			}.Respond(w)
			return
		}
		w.Header().Add("x-reason", "unsupported-uri")
		w.WriteHeader(404)
	})
	defer ts.Close()
	a := zattoo.NewAccountWithSession(t, host, client)
	s := &server{
		a:   a,
		hub: newHub(),
	}
	q := &downloadQueue{
		server: s,
		mu:     sync.Mutex{},
		q:      []toDownload{},
	}
	done := make(chan struct{}, 1)
	q.downloadRecording(toDownload{RecordingId: 1111}, done)
	<-done

	consumeEvents(t, s.hub.outbox, []event{
		{StateUpdated: &eventStateUpdated{State: "get_stream_url", Reason: "getting recording stream URL ..."}},
		{StateUpdated: &eventStateUpdated{State: "detect_streams", Reason: "detecting recording audio and video streams ..."}},
		{StateUpdated: &eventStateUpdated{State: "download", Reason: "starting download ..."}},
	})
	ensureNoMoreEvents(t, s.hub.outbox)
}

func Test_downloadQueue_Enqueue(t *testing.T) {
	type args struct {
		recordingId int64
		outputPath  string
	}
	tests := []struct {
		name         string
		q            []toDownload
		args         args
		wantQueueLen int
	}{
		{
			name:         "EmptyQueue",
			q:            []toDownload{},
			args:         args{recordingId: 1234, outputPath: "test"},
			wantQueueLen: 1,
		},
		{
			name: "QueueWith/1-entry",
			q: []toDownload{
				{RecordingId: 11, OutputPath: "foo"},
			},
			args:         args{recordingId: 22, outputPath: "bar"},
			wantQueueLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				hub: newHub(),
			}
			q := &downloadQueue{
				server: s,
				mu:     sync.Mutex{},
				q:      tt.q,
			}

			q.Enqueue(tt.args.recordingId, tt.args.outputPath)
			if len(q.q) != tt.wantQueueLen {
				t.Errorf("queue length is %d, but want %d", len(q.q), tt.wantQueueLen)
			}
			consumeEvent(t, s.hub.outbox, event{QueueUpdated: &eventQueueUpdated{
				Queue: append(tt.q, toDownload{tt.args.recordingId, tt.args.outputPath}),
			}})
			ensureNoMoreEvents(t, s.hub.outbox)
		})
	}
}

func Test_broadcastDownloadProgressHandler_Start(t *testing.T) {
	type fields struct {
		eventQueueUpdated eventQueueUpdated
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &broadcastDownloadProgressHandler{
				eventQueueUpdated: tt.fields.eventQueueUpdated,
			}
			b.Start()
		})
	}
}

func Test_broadcastDownloadProgressHandler_Error(t *testing.T) {
	type fields struct {
		eventQueueUpdated eventQueueUpdated
	}
	tests := []struct {
		name   string
		fields fields
		err    error
	}{
		{
			name: "Success",
			err:  errors.New("injected"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &broadcastDownloadProgressHandler{
				eventQueueUpdated: tt.fields.eventQueueUpdated,
			}
			b.Error(tt.err)
		})
	}
}

func Test_broadcastDownloadProgressHandler_Finished(t *testing.T) {
	type fields struct {
		eventQueueUpdated eventQueueUpdated
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &broadcastDownloadProgressHandler{
				eventQueueUpdated: tt.fields.eventQueueUpdated,
			}
			b.Finished()
		})
	}
}

func Test_broadcastDownloadProgressHandler_UpdateProgress(t *testing.T) {
	type fields struct {
		eventQueueUpdated eventQueueUpdated
	}
	tests := []struct {
		name   string
		fields fields
		p      ffmpeg.DownloadProgress
	}{
		{
			name: "Success",
			p: ffmpeg.DownloadProgress{
				RelCompleted: 0.123,
				Elapsed:      time.Millisecond * 12345,
				Remaining:    time.Millisecond * 98765,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				hub: newHub(),
			}
			b := &broadcastDownloadProgressHandler{
				server:            s,
				eventQueueUpdated: tt.fields.eventQueueUpdated,
			}
			b.UpdateProgress(tt.p)
			consumeEvent(t, b.hub.outbox, event{ProgressUpdated: &eventProgressUpdated{
				RelCompleted: tt.p.RelCompleted,
				Elapsed:      tt.p.Elapsed.Truncate(time.Second).String(),
				Remaining:    tt.p.Remaining.String(),
			}})
			ensureNoMoreEvents(t, s.hub.outbox)
		})
	}
}
