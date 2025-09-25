package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type downloadProgressTracker struct {
	outType string
	source  io.Reader
	target  io.Writer

	start        time.Time
	durationMsec int64
}

func (t *downloadProgressTracker) showDownloadProgress() {
	r := regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2}).(\d{1,3})`)
	scanner := bufio.NewScanner(t.source)

	for scanner.Scan() {
		line := scanner.Text()
		pos := strings.Index(line, "time=")
		if pos >= 0 {
			m := r.FindStringSubmatch(line)
			if nil == m || len(m) <= 0 {
				continue
			}

			strH, strM, strS, strMS := m[1], m[2], m[3], m[4]
			posMsec, err := parseTimeMsec(strH, strM, strS, strMS)
			if nil != err {
				continue
			}
			relPos := float32(posMsec) / float32(t.durationMsec)
			elapsed := time.Now().UTC().Sub(t.start)
			if relPos > 0 {
				estimatedTotal := int64(float32(elapsed.Milliseconds()) / relPos)
				remaining :=
					(time.Millisecond * time.Duration(estimatedTotal-elapsed.Milliseconds())).
						Truncate(time.Second)
				fmt.Fprintf(t.target,
					"Download progress: %5.1f%% | Elapsed: %10s | Remaining: %10s\r",
					relPos*100, elapsed.Truncate(time.Second), remaining)
			} else {
				fmt.Fprintf(t.target,
					"Download progress: %5.1f%% | Elapsed: %10s\r",
					relPos*100, elapsed.Truncate(time.Second))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(t.target)
}

func parseTimeMsec(strH, strM, strS, strMS string) (int64, error) {
	h, err := strconv.ParseInt(strH, 10, 8)
	if nil != err {
		return 0, err
	} else if h < 0 {
		return 0, fmt.Errorf("hours must not be negative: %d", h)
	}

	m, err := strconv.ParseInt(strM, 10, 8)
	if nil != err {
		return 0, err
	} else if m < 0 || m > 59 {
		return 0, fmt.Errorf("minutes must be between 0 and 59 inclusive: %d", m)
	}

	s, err := strconv.ParseInt(strS, 10, 8)
	if nil != err {
		return 0, err
	} else if s < 0 || s > 59 {
		return 0, fmt.Errorf("seconds must be between 0 and 59 inclusive: %d", s)
	}

	msVal, err := strconv.ParseInt(strMS, 10, 16)
	if nil != err {
		return 0, err
	}

	ms := int64(0)
	switch len(strMS) {
	case 1:
		ms = msVal * 100

	case 2:
		ms = msVal * 10

	case 3:
		ms = msVal

	default:
		return 0, fmt.Errorf("milliseconds string '%s' has unsupported length", strMS)
	}

	return (h * 60 * 60 * 1000) + (m * 60 * 1000) + (s * 1000) + ms, nil
}
