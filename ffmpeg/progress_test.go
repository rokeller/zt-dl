package ffmpeg

import "testing"

func Test_parseTimeMsec(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
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
			name:    "Bad Minute",
			strH:    "23",
			strM:    "cd",
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
