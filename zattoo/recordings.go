package zattoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type playlistResponse struct {
	Recordings []recording `json:"recordings"`
	Success    bool        `json:"success"`
}

type recording struct {
	Id           int64     `json:"id"`
	ProgramId    int64     `json:"program_id"`
	ChannelId    string    `json:"cid"`
	Level        string    `json:"level"`
	Title        string    `json:"title"`
	EpisodeTitle string    `json:"episode_title"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
}

type watchRecordingResponse struct {
	Csid            string `json:"csid"`
	Stream          stream `json:"stream"`
	Success         bool   `json:"success"`
	DrmLimitApplied bool   `json:"drm_limit_applied"`
}

type stream struct {
	WatchUrls      []watchUrl `json:"watch_urls"`
	Url            string     `json:"url"`
	Quality        string     `json:"quality"`
	ForwardSeeking bool       `json:"forward_seeking"`
}

type watchUrl struct {
	Url         string `json:"url"`
	MaxRate     int64  `json:"maxrate"`
	AudiChannel string `json:"audio_channel"`
}

type programDetailsResponse struct {
	Success  bool             `json:"success"`
	Programs []programDetails `json:"programs"`
}

type programDetails struct {
	ChannelName string `json:"channel_name"`
	ChannelId   string `json:"cid"`
	Title       string `json:"t"`
	Description string `json:"d"`
	Year        int    `json:"year"`
	Start       int64  `json:"s"`
	End         int64  `json:"e"`
}

func (s *session) getPlaylist(a Account) ([]recording, error) {
	resp, err := s.client.Get(fmt.Sprintf("https://%s/zapi/v2/playlist", a.domain))
	if nil != err {
		return nil, err
	}

	defer resp.Body.Close()
	var res playlistResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); nil != err {
		return nil, err
	}

	if !res.Success {
		return nil, errors.New("failed to fetch playlist")
	}

	return res.Recordings, nil
}

func (s *session) getRecording(a Account, id int64) (stream, error) {
	data := url.Values{}
	data.Set("with_schedule", "false")
	data.Set("stream_type", "hls7")
	data.Set("https_watch_urls", "true")
	data.Set("sdh_subtitles", "true")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://%s/zapi/watch/recording/%d", a.domain, id),
		strings.NewReader(data.Encode()))
	if nil != err {
		return stream{}, err
	}
	resp, err := s.client.Do(req)
	if nil != err {
		return stream{}, err
	}

	defer resp.Body.Close()
	var res watchRecordingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); nil != err {
		return stream{}, err
	}

	if !res.Success {
		return stream{}, errors.New("failed to get recording details")
	}

	return res.Stream, nil
}

func (s *session) getProgramDetails(a Account, id int64) (programDetails, error) {
	params := url.Values{}
	params.Set("program_ids", strconv.FormatInt(id, 10))
	u := fmt.Sprintf("https://%s/zapi/v2/cached/program/power_details/%s?%s",
		a.domain, s.powerGuideHash, params.Encode())
	resp, err := s.client.Get(u)
	if nil != err {
		return programDetails{}, err
	}

	defer resp.Body.Close()
	var res programDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); nil != err {
		return programDetails{}, err
	}

	if !res.Success {
		return programDetails{}, fmt.Errorf("failed to get program details (status %d)", resp.StatusCode)
	}

	return res.Programs[0], nil
}
