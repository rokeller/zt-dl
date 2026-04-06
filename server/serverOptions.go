package server

import (
	"github.com/rokeller/zt-dl/ffmpeg"
	"github.com/rokeller/zt-dl/zattoo"
)

func WithZattooAccount(a *zattoo.Account) ServeOption {
	return func(s *server) {
		s.a = a
	}
}

func WithPort(port uint16) ServeOption {
	return func(s *server) {
		s.port = port
	}
}

func WithOutputDir(outdir string) ServeOption {
	return func(s *server) {
		s.outdir = outdir
	}
}

func WithOverwrite(overwrite bool) ServeOption {
	return func(s *server) {
		s.overwrite = overwrite
	}
}

func WithOpenWebUI(openWebUI bool) ServeOption {
	return func(s *server) {
		s.openWebUI = openWebUI
	}
}

func WithBestStreamsSelection() ServeOption {
	return func(s *server) {
		s.streamsSelectorFactory = bestStreamsSelectorFactory
	}
}

func WithInteractiveStreamsSelection() ServeOption {
	return func(s *server) {
		s.streamsSelectorFactory = func() ffmpeg.StreamsSelector {
			return newInteractiveStreamsSelector(s.hub)
		}
	}
}

func bestStreamsSelectorFactory() ffmpeg.StreamsSelector {
	return ffmpeg.NewBestStreamsSelector()
}
