package ffmpeg

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func Test_downloadProgressTracker_showDownloadProgress(t *testing.T) {
	tests := []struct {
		name        string // description of this test case
		pipeIn      string
		expectedOut string
	}{
		{
			name:        "No valid progress updates",
			pipeIn:      "blah\nblotz\nfoo\ntime=00:60:00.12",
			expectedOut: "\n",
		},
		{
			name:        "Zero progress",
			pipeIn:      "time=00:00:00.00",
			expectedOut: "Download progress:   0.0% | Elapsed:        10s\r\n",
		},
		{
			name:        "Single update with time",
			pipeIn:      "blah\ntime=n/a\nfoo=x time=00:01:40.00 bar=123\n",
			expectedOut: "Download progress:  10.0% | Elapsed:        10s | Remaining:      1m30s\r\n",
		},
		{
			name:   "Multiple updates with time",
			pipeIn: "blah\nfoo=x time=00:01:40.00 bar=123\n\nqwer time=00:15:00.00 asdf\n",
			expectedOut: "" +
				"Download progress:  10.0% | Elapsed:        10s | Remaining:      1m30s\r" +
				"Download progress:  90.0% | Elapsed:        10s | Remaining:         1s\r" +
				"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := bytes.Buffer{}
			d := downloadProgressTracker{
				source: strings.NewReader(tt.pipeIn),
				target: &target,

				start:        time.Now().Add(-10 * time.Second),
				durationMsec: (1000 * time.Second).Milliseconds(),
			}

			d.showDownloadProgress()
			actualOut := target.String()
			if actualOut != tt.expectedOut {
				t.Errorf("showDownloadProgress() produced output %q, expected %q", actualOut, tt.expectedOut)
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
