package ffmpeg

type DownloadableOption func(*downloadable)

func WithOverwrite(overwrite bool) DownloadableOption {
	return func(d *downloadable) {
		d.overwrite = overwrite
	}
}
