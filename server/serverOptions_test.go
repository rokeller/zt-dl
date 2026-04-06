package server

import (
	"reflect"
	"testing"
)

func TestWithBestStreamsSelection(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		verify func(t *testing.T, s *server)
	}{
		{
			name: "BestStreamsSelection",
			verify: func(t *testing.T, s *server) {
				sel := s.streamsSelectorFactory()
				tt := reflect.TypeOf(sel)
				if tt.String() != "ffmpeg.bestStreamsSelector" {
					t.Errorf(
						"StreamsSelector: got type %q, want \"ffmpeg.bestStreamsSelector\"",
						tt.String())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{}
			opt := WithBestStreamsSelection()
			opt(s)
			tt.verify(t, s)
		})
	}
}

func TestWithInteractiveStreamsSelection(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		verify func(t *testing.T, s *server)
	}{
		{
			name: "InteractiveStreamsSelector",
			verify: func(t *testing.T, s *server) {
				sel := s.streamsSelectorFactory()
				tt := reflect.TypeOf(sel)
				if tt.String() != "*server.interactiveStreamsSelector" {
					t.Errorf(
						"StreamsSelector: got type %q, want \"server.interactiveStreamsSelector\"",
						tt.String())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				hub: &wsHub{},
			}
			opt := WithInteractiveStreamsSelection()
			opt(s)
			tt.verify(t, s)
		})
	}
}
