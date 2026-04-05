package ffmpeg

type StreamsSelector interface {
	SelectStreams(streams []SourceStream) ([]SourceStream, error)
}

type bestStreamsSelector struct {
	languagePreference []string
}

var _ StreamsSelector = bestStreamsSelector{}

func NewBestStreamsSelectorWithSubtitles(languagePreference ...string) bestStreamsSelector {
	return bestStreamsSelector{languagePreference: languagePreference}
}

// SelectStreams implements [StreamsSelector].
func (b bestStreamsSelector) SelectStreams(streams []SourceStream) ([]SourceStream, error) {
	res := []SourceStream{}

	if as := bestAudioStream(FilterStreams(streams, IsAudioStream)); nil != as {
		res = append(res, as)
	}
	if vs := bestVideoStream(FilterStreams(streams, IsVideoStream)); nil != vs {
		res = append(res, vs)
	}

	for _, s := range FilterStreams(streams, IsSubtitleStream) {
		sts := s.(*SubtitleStream)
		for _, lang := range b.languagePreference {
			if lang == sts.Language {
				res = append(res, sts)
			}
		}
	}
	return res, nil
}

func bestAudioStream(streams []SourceStream) SourceStream {
	if len(streams) < 1 {
		return nil
	} else if len(streams) == 1 {
		return streams[0]
	}

	var best *AudioStream
	for _, s := range streams {
		as := s.(*AudioStream)
		if nil == best || as.SampleRate > best.SampleRate {
			best = as
		}
	}
	return best
}

func bestVideoStream(streams []SourceStream) SourceStream {
	if len(streams) < 1 {
		return nil
	} else if len(streams) == 1 {
		return streams[0]
	}

	var best *VideoStream
	for _, s := range streams {
		vs := s.(*VideoStream)
		if nil == best ||
			(vs.Width > best.Width && vs.Height > best.Height && vs.AvgFrameRate > best.AvgFrameRate) {
			best = vs
		}
	}
	return best
}
