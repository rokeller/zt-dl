package ffmpeg

type SourceStreamPredicate func(SourceStream) bool

func FilterStreams(streams []SourceStream, predicate SourceStreamPredicate) []SourceStream {
	res := make([]SourceStream, 0, len(streams))
	for _, stream := range streams {
		if predicate(stream) {
			res = append(res, stream)
		}
	}
	return res
}

func IsAudioStream(s SourceStream) bool {
	_, ok := s.(*AudioStream)
	return ok
}

func IsSubtitleStream(s SourceStream) bool {
	_, ok := s.(*SubtitleStream)
	return ok
}

func IsVideoStream(s SourceStream) bool {
	_, ok := s.(*VideoStream)
	return ok
}
