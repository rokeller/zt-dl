package ffmpeg

import (
	"reflect"
	"testing"
)

func TestFilterStreams(t *testing.T) {
	sourceStreams := []SourceStream{
		&AudioStream{
			Stream: Stream{
				Index: 0,
			},
		},
		&VideoStream{
			Stream: Stream{
				Index: 1,
			},
		},
		&SubtitleStream{
			Stream: Stream{
				Index: 2,
			},
		},
	}
	tests := []struct {
		name      string
		predicate SourceStreamPredicate
		want      []SourceStream
	}{
		{
			name:      "AudioOnly",
			predicate: IsAudioStream,
			want: []SourceStream{
				&AudioStream{
					Stream: Stream{
						Index: 0,
					},
				},
			},
		},
		{
			name:      "VideoOnly",
			predicate: IsVideoStream,
			want: []SourceStream{
				&VideoStream{
					Stream: Stream{
						Index: 1,
					},
				},
			},
		},
		{
			name:      "SubtitleOnly",
			predicate: IsSubtitleStream,
			want: []SourceStream{
				&SubtitleStream{
					Stream: Stream{
						Index: 2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterStreams(sourceStreams, tt.predicate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterStreams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransformStreams(t *testing.T) {
	sourceStreams := []SourceStream{
		&AudioStream{
			Stream: Stream{
				Index:     0,
				CodecName: "audio_codec",
			},
			SampleRate:    111,
			Channels:      2,
			ChannelLayout: "test_audio",
			Language:      "audio_lang",
		},
		&VideoStream{
			Stream: Stream{
				Index:     1,
				CodecName: "video_codec",
			},
			Width:        123,
			Height:       456,
			AvgFrameRate: 78,
			BitRate:      9000,
		},
		&SubtitleStream{
			Stream: Stream{
				Index:     2,
				CodecName: "subtitle_codec",
			},
			Language: "srt_lang",
		},
	}
	tests := []struct {
		name        string
		transformer SourceStreamTransformer[string]
		want        []string
	}{
		{
			name: "Stringer",
			transformer: func(ss SourceStream) string {
				return ss.String()
			},
			want: []string{
				"audio_codec, sample rate 111Hz, 2 channels (test_audio), language \"audio_lang\" (stream #0)",
				"video_codec, width/height 123/456, bit rate 9000bps, avg frame rate 78fps (stream #1)",
				"subtitle_codec, language \"srt_lang\" (stream #2)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TransformStreams(sourceStreams, tt.transformer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransformStreams() = %v, want %v", got, tt.want)
			}
		})
	}
}
