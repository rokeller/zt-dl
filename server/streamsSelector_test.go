package server

import (
	"reflect"
	"testing"

	"github.com/rokeller/zt-dl/ffmpeg"
)

type unknownSourceStream struct {
	index int
	name  string
}

var _ ffmpeg.SourceStream = unknownSourceStream{}

// Index implements [ffmpeg.SourceStream].
func (u unknownSourceStream) Index() int {
	return u.index
}

// String implements [ffmpeg.SourceStream].
func (u unknownSourceStream) String() string {
	return u.name
}

func Test_newInteractiveStreamsSelector(t *testing.T) {
	hub := &wsHub{
		addHandler:    make(chan clientEventHandler),
		removeHandler: make(chan clientEventHandler),
		outbox:        make(chan serverEvent),
	}
	want := &interactiveStreamsSelector{
		add:    hub.addHandler,
		remove: hub.removeHandler,
		outbox: hub.outbox,
	}
	if got := newInteractiveStreamsSelector(hub); !reflect.DeepEqual(got, want) {
		t.Errorf("newInteractiveStreamsSelector() = %v, want %v", got, want)
	}
}

func Test_interactiveStreamsSelector_SelectStreams(t *testing.T) {
	s := &interactiveStreamsSelector{
		add:    make(chan clientEventHandler),
		remove: make(chan clientEventHandler),
		outbox: make(chan serverEvent),
	}

	ss := []ffmpeg.SourceStream{
		&ffmpeg.AudioStream{
			Stream: ffmpeg.Stream{
				Index:     0,
				CodecName: "aac",
			},
			SampleRate: 22000,
		},
		&ffmpeg.AudioStream{
			Stream: ffmpeg.Stream{
				Index:     1,
				CodecName: "aac",
			},
			SampleRate: 44000,
		},
		&ffmpeg.SubtitleStream{
			Stream: ffmpeg.Stream{
				Index:     2,
				CodecName: "srt",
			},
			Language: "tst",
		},
		&ffmpeg.VideoStream{
			Stream: ffmpeg.Stream{
				Index:     3,
				CodecName: "h264",
			},
			Width:        1280,
			Height:       720,
			AvgFrameRate: 25,
			BitRate:      12000000,
		},
		&ffmpeg.VideoStream{
			Stream: ffmpeg.Stream{
				Index:     4,
				CodecName: "h264",
			},
			Width:        1280,
			Height:       720,
			AvgFrameRate: 50,
			BitRate:      24000000,
		},
	}

	go func() {
		t.Helper()
		handler := <-s.add
		serverEvent := <-s.outbox
		correlation := serverEvent.Correlation
		if nil == serverEvent.StreamSelectionRequested {
			t.Errorf("Outgoing event got %#v, want StreamSelectionRequested", serverEvent)
		}

		// Send client event with stream selection to handler
		clientEvent := clientEvent{
			Correlation: correlation,
			StreamsSelected: &eventStreamsSelected{
				SelectedStreams: []sourceStream{
					{Index: 0},
					{Index: 2},
					{Index: 3},
				},
			},
		}
		go handler.Handle(sourcedClientEvent{event: clientEvent})

		handlerRemoved := <-s.remove
		if handler != handlerRemoved {
			t.Errorf("Got removed handler %v, want %v", handlerRemoved, handler)
		}
	}()

	got, err := s.SelectStreams(ss)
	if nil != err {
		t.Errorf("SelectStreams() got error %v, want nil", err)
	}
	want := []ffmpeg.SourceStream{ss[0], ss[2], ss[3]}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SelectStreams() got streams %#v, want %#v", got, want)
	}
}

func Test_sourceStreamSelectedHandler_Handle(t *testing.T) {
	type fields struct {
		correlation string
		done        chan []sourceStream
	}
	tests := []struct {
		name              string
		fields            fields
		e                 sourcedClientEvent
		wantDoneChanLen   int
		wantDone          []sourceStream
		wantRemoveChanLen int
	}{
		{
			name: "NonMatch/MissingEventData",
			fields: fields{
				correlation: "non-match",
				done:        make(chan []sourceStream, 1),
			},
			e: sourcedClientEvent{
				event: clientEvent{Correlation: "non-match"},
			},
			wantDoneChanLen:   0, // We didn't get a result, no selection
			wantRemoveChanLen: 0, // We didn't get a result, handler not removed
		},
		{
			name: "NonMatch/CorrelationMismatch",
			fields: fields{
				correlation: "corr1",
				done:        make(chan []sourceStream, 1),
			},
			e: sourcedClientEvent{
				event: clientEvent{Correlation: "corr2", StreamsSelected: &eventStreamsSelected{}},
			},
			wantDoneChanLen:   0, // We didn't get a result, no selection
			wantRemoveChanLen: 0, // We didn't get a result, handler not removed
		},
		{
			name: "Match/NullSelection",
			fields: fields{
				correlation: "match-1",
				done:        make(chan []sourceStream, 1),
			},
			e: sourcedClientEvent{
				event: clientEvent{
					Correlation:     "match-1",
					StreamsSelected: &eventStreamsSelected{},
				},
			},
			wantDoneChanLen:   1, // We got a result, it's in done
			wantDone:          nil,
			wantRemoveChanLen: 1, // We got a result, handler remove requested
		},
		{
			name: "Match/EmptySelection",
			fields: fields{
				correlation: "match-2",
				done:        make(chan []sourceStream, 1),
			},
			e: sourcedClientEvent{
				event: clientEvent{
					Correlation: "match-2",
					StreamsSelected: &eventStreamsSelected{
						SelectedStreams: []sourceStream{},
					},
				},
			},
			wantDoneChanLen:   1, // We got a result, it's in done
			wantDone:          []sourceStream{},
			wantRemoveChanLen: 1, // We got a result, handler remove requested
		},
		{
			name: "Match/WithSelection",
			fields: fields{
				correlation: "match-3",
				done:        make(chan []sourceStream, 1),
			},
			e: sourcedClientEvent{
				event: clientEvent{
					Correlation: "match-3",
					StreamsSelected: &eventStreamsSelected{
						SelectedStreams: []sourceStream{
							{Index: 123},
						},
					},
				},
			},
			wantDoneChanLen: 1, // We got a result, it's in done
			wantDone: []sourceStream{
				{Index: 123},
			},
			wantRemoveChanLen: 1, // We got a result, handler remove requested
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &interactiveStreamsSelector{
				// add:    make(chan clientEventHandler),
				remove: make(chan clientEventHandler, 1),
				// outbox: make(chan serverEvent),
			}
			h := &sourceStreamSelectedHandler{
				owner:       ss,
				correlation: tt.fields.correlation,
				done:        tt.fields.done,
			}
			h.Handle(tt.e)

			gotLen := len(tt.fields.done)
			if gotLen != tt.wantDoneChanLen {
				t.Errorf("done channel: got length %d, want %d", gotLen, tt.wantDoneChanLen)
			}

			if tt.wantDoneChanLen > 0 {
				gotDone := <-tt.fields.done
				if !reflect.DeepEqual(gotDone, tt.wantDone) {
					t.Errorf("selected streams: got %#v, want %#v", gotDone, tt.wantDone)
				}
			}

			gotLen = len(ss.remove)
			if gotLen != tt.wantRemoveChanLen {
				t.Errorf("remove channel: got length %d, want %d", gotLen, tt.wantRemoveChanLen)
			}
		})
	}
}

func Test_sourceStreamToSourceStreamDesc(t *testing.T) {
	type args struct {
		s ffmpeg.SourceStream
	}
	tests := []struct {
		name string
		s    ffmpeg.SourceStream
		want sourceStream
	}{
		{
			name: "AudioStream",
			s: &ffmpeg.AudioStream{
				Stream: ffmpeg.Stream{
					Index:     1,
					CodecName: "TestAudio",
				},
				SampleRate:    22000,
				Channels:      3,
				ChannelLayout: "TestStereo",
				Language:      "ut",
			},
			want: sourceStream{
				Index:       1,
				Type:        "Audio",
				Description: "TestAudio, sample rate 22000Hz, 3 channels (TestStereo), language \"ut\" (stream #1)",
			},
		},
		{
			name: "SubtitleStream",
			s: &ffmpeg.SubtitleStream{
				Stream: ffmpeg.Stream{
					Index:     2,
					CodecName: "TestText",
				},
				Language: "TestLang",
			},
			want: sourceStream{
				Index:       2,
				Type:        "Subtitle",
				Description: "TestText, language \"TestLang\" (stream #2)",
			},
		},
		{
			name: "VideoStream",
			s: &ffmpeg.VideoStream{
				Stream: ffmpeg.Stream{
					Index:     3,
					CodecName: "TestVideo",
				},
				Width:        1280,
				Height:       720,
				BitRate:      1200000,
				AvgFrameRate: 25,
			},
			want: sourceStream{
				Index:       3,
				Type:        "Video",
				Description: "TestVideo, dimensions 1280x720, bit rate 1200000bps, avg frame rate 25fps (stream #3)",
			},
		},
		{
			name: "UnknownStream",
			s:    unknownSourceStream{4, "TestStream"},
			want: sourceStream{
				Index:       4,
				Type:        "Unknown",
				Description: "TestStream",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourceStreamToSourceStreamDesc(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sourceStreamToSourceStreamDesc() = %v, want %v", got, tt.want)
			}
		})
	}
}
