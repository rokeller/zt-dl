package server

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rokeller/zt-dl/ffmpeg"
)

type interactiveStreamsSelector struct {
	add    chan clientEventHandler
	remove chan clientEventHandler
	outbox chan serverEvent
}

func newInteractiveStreamsSelector(hub *wsHub) ffmpeg.StreamsSelector {
	return &interactiveStreamsSelector{
		add:    hub.addHandler,
		remove: hub.removeHandler,
		outbox: hub.outbox,
	}
}

// SelectStreams implements [ffmpeg.StreamsSelector].
func (s *interactiveStreamsSelector) SelectStreams(
	streams []ffmpeg.SourceStream,
) ([]ffmpeg.SourceStream, error) {
	// Send a 'StreamSelectionRequested' event to the clients and wait for a
	// client to send a selection back to us.
	correlation := uuid.New().String()
	h := &sourceStreamSelectedHandler{
		owner:       s,
		correlation: correlation,
		done:        make(chan []sourceStream),
	}
	// Register the event handler in the hub.
	s.add <- h
	// Notify the hub clients about the requested stream selection.
	s.outbox <- serverEvent{
		Correlation: correlation,
		StreamSelectionRequested: &eventStreamSelectionRequested{
			SourceStreams: ffmpeg.TransformStreams(streams, sourceStreamToSourceStreamDesc),
		},
	}
	fmt.Println("Waiting for source stream selection by user ...")
	selected := <-h.done
	fmt.Println("Stream selection finished.")

	// Now select those streams from the input that match the selected streams.
	return ffmpeg.FilterStreams(streams, func(stream ffmpeg.SourceStream) bool {
		for _, sel := range selected {
			if sel.Index == stream.Index() {
				return true
			}
		}
		return false
	}), nil
}

type sourceStreamSelectedHandler struct {
	owner       *interactiveStreamsSelector
	correlation string
	done        chan []sourceStream
}

var _ clientEventHandler = &sourceStreamSelectedHandler{}

func (h *sourceStreamSelectedHandler) Handle(e sourcedClientEvent) {
	se := e.event
	if nil == se.StreamsSelected || h.correlation != se.Correlation {
		// We don't care about this event - it's not for streams selection or it
		// doesn't match our expected correlation ID.
		return
	}
	// Remove this handler now so it won't be called for other client events.
	h.owner.remove <- h

	selected := se.StreamsSelected.SelectedStreams
	fmt.Printf("User selected %d stream(s) for download.\n", len(selected))
	h.done <- selected
}

func sourceStreamToSourceStreamDesc(s ffmpeg.SourceStream) sourceStream {
	var t string
	switch s.(type) {
	case *ffmpeg.AudioStream:
		t = "Audio"
	case *ffmpeg.SubtitleStream:
		t = "Subtitle"
	case *ffmpeg.VideoStream:
		t = "Video"
	default:
		t = "Unknown"
	}
	return sourceStream{
		Index:       s.Index(),
		Type:        t,
		Description: s.String(),
	}
}
