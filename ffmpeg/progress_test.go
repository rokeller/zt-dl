package ffmpeg

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_downloadProgressTracker_showDownloadProgress(t *testing.T) {
	tests := []struct {
		name             string // description of this test case
		pipeIn           string
		expectedProgress []DownloadProgress
	}{
		{
			name:             "No valid progress updates",
			pipeIn:           "blah\nblotz\nfoo\ntime=00:60:00.12",
			expectedProgress: []DownloadProgress{},
		},
		{
			name:   "Zero progress",
			pipeIn: "time=00:00:00.00",
			expectedProgress: []DownloadProgress{
				{RelCompleted: 0, Elapsed: time.Second * 10, Remaining: 24 * 999 * time.Hour},
			},
		},
		{
			name:   "Single update with time",
			pipeIn: "blah\ntime=n/a\nfoo=x time=00:01:40.00 bar=123\n",
			expectedProgress: []DownloadProgress{
				{RelCompleted: .1, Elapsed: time.Second * 10, Remaining: time.Minute + time.Second*30},
			},
		},
		{
			name:   "Multiple updates with time",
			pipeIn: "blah\nfoo=x time=00:01:40.00 bar=123\n\nqwer time=00:15:00.00 asdf\n",
			expectedProgress: []DownloadProgress{
				{RelCompleted: .1, Elapsed: time.Second * 10, Remaining: time.Minute + time.Second*30},
				{RelCompleted: .9, Elapsed: time.Second * 10, Remaining: time.Second},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &testProgressHandler{progressUpdates: []DownloadProgress{}}
			d := downloadProgressTracker{
				handler: h,
				source:  strings.NewReader(tt.pipeIn),

				start:        time.Now().Add(-10 * time.Second),
				durationMsec: (1000 * time.Second).Milliseconds(),
			}

			d.trackProgress()
			actualProgress := h.progressUpdates
			if !reflect.DeepEqual(actualProgress, tt.expectedProgress) {
				t.Errorf("showDownloadProgress() produced output %v, expected %v", actualProgress, tt.expectedProgress)
			}
		})
	}
}

func Test_parseTimeMsec(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		strH    string
		strM    string
		strS    string
		strMS   string
		want    int64
		wantErr bool
	}{
		{
			name:    "Good (msec-3)",
			strH:    "01",
			strM:    "23",
			strS:    "45",
			strMS:   "678",
			want:    1*60*60*100 + 23*60*1000 + 45*1000 + 678,
			wantErr: false,
		},
		{
			name:    "Good (msec-2)",
			strH:    "01",
			strM:    "23",
			strS:    "45",
			strMS:   "67",
			want:    1*60*60*100 + 23*60*1000 + 45*1000 + 670,
			wantErr: false,
		},
		{
			name:    "Good (msec-1)",
			strH:    "01",
			strM:    "23",
			strS:    "45",
			strMS:   "6",
			want:    1*60*60*100 + 23*60*1000 + 45*1000 + 600,
			wantErr: false,
		},
		{
			name:    "Bad Hour",
			strH:    "ab",
			wantErr: true,
		},
		{
			name:    "Negative Hour",
			strH:    "-1",
			wantErr: true,
		},
		{
			name:    "Bad Minute",
			strH:    "23",
			strM:    "cd",
			wantErr: true,
		},
		{
			name:    "Negative Minute",
			strH:    "23",
			strM:    "-1",
			wantErr: true,
		},
		{
			name:    "Minute >59",
			strH:    "23",
			strM:    "60",
			wantErr: true,
		},
		{
			name:    "Bad Second",
			strH:    "23",
			strM:    "59",
			strS:    "ef",
			wantErr: true,
		},
		{
			name:    "Negative Second",
			strH:    "23",
			strM:    "59",
			strS:    "-1",
			wantErr: true,
		},
		{
			name:    "Second >59",
			strH:    "23",
			strM:    "59",
			strS:    "60",
			wantErr: true,
		},
		{
			name:    "Bad Millisecond",
			strH:    "23",
			strM:    "59",
			strS:    "00",
			strMS:   "g",
			wantErr: true,
		},
		{
			name:    "Bad Millisecond (4 digits)",
			strH:    "23",
			strM:    "59",
			strS:    "00",
			strMS:   "1234",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseTimeMsec(tt.strH, tt.strM, tt.strS, tt.strMS)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseTimeMsec() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseTimeMsec() succeeded unexpectedly")
			}
			if got == tt.want {
				t.Errorf("parseTimeMsec() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testProgressHandler struct {
	started         bool
	progressUpdates []DownloadProgress
	err             error
	finished        bool
}

func (t *testProgressHandler) Start() {
	t.started = true
}

func (t *testProgressHandler) UpdateProgress(p DownloadProgress) {
	p.Elapsed = p.Elapsed.Truncate(time.Second)
	t.progressUpdates = append(t.progressUpdates, p)
}

func (t *testProgressHandler) Error(err error) {
	t.err = err
}

func (t *testProgressHandler) Finished() {
	t.finished = true
}

func Test_consoleProgressHandler_UpdateProgress(t *testing.T) {
	tests := []struct {
		name           string // description of this test case
		p              DownloadProgress
		expectedOutput []byte
	}{
		{
			name: "Update",
			p: DownloadProgress{
				RelCompleted: 0.1234,
				Elapsed:      time.Minute*2 + time.Second*3 + time.Millisecond*4,
				Remaining:    time.Minute*5 + time.Second*6,
			},
			expectedOutput: []byte("Download progress:  12.3% | Elapsed:       2m3s | Remaining:       5m6s\r"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			h := consoleProgressHandler{
				target: buf,
			}
			h.UpdateProgress(tt.p)
			actualOutput := buf.Bytes()

			if !reflect.DeepEqual(actualOutput, tt.expectedOutput) {
				t.Errorf("got output = %q, want %q", buf, tt.expectedOutput)
			}
		})
	}
}
