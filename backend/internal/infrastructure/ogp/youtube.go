package ogp

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
)

var (
	youTubeIDPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	isoDurationPattern    = regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$`)
	youTubeHosts          = map[string]bool{"youtube.com": true, "www.youtube.com": true, "m.youtube.com": true, "music.youtube.com": true, "www.youtube-nocookie.com": true}
	youTubePathIDPrefixes = []string{"/shorts/", "/embed/", "/live/", "/v/"}
)

// youTubeVideoID は YouTube の動画 URL から動画 ID を取り出します
func youTubeVideoID(u *url.URL) (string, bool) {
	var id string
	switch host := strings.ToLower(u.Hostname()); {
	case host == "youtu.be":
		id = strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)[0]
	case youTubeHosts[host] && u.Path == "/watch":
		id = u.Query().Get("v")
	case youTubeHosts[host]:
		for _, prefix := range youTubePathIDPrefixes {
			if rest, ok := strings.CutPrefix(u.Path, prefix); ok {
				id = strings.SplitN(rest, "/", 2)[0]
			}
		}
	}
	return id, youTubeIDPattern.MatchString(id)
}

// fetchYouTube は動画ページの OGP と microdata を読み、取れなければ oEmbed で補います。外部 API キーは使いません
func (s *OGPService) fetchYouTube(ctx context.Context, videoID string) *entity.OGPData {
	watchURL := &url.URL{Scheme: "https", Host: "www.youtube.com", Path: "/watch", RawQuery: "v=" + videoID}

	meta, err := s.fetchMeta(ctx, watchURL)
	if err != nil {
		meta = map[string]string{}
	}
	data := buildOGPData(meta, watchURL)
	data.YouTube = &entity.YouTubeVideo{
		VideoID:         videoID,
		ChannelName:     firstOf(meta, "author:name"),
		DurationSeconds: parseISODuration(meta["itemprop:duration"]),
	}

	// 同意画面などで動画ページのメタデータが取れない場合に備える
	if data.Title == nil {
		s.applyOEmbed(ctx, watchURL, data)
	}
	if data.ImageURL == nil {
		data.ImageURL = nonEmpty("https://i.ytimg.com/vi/" + videoID + "/hqdefault.jpg")
		data.ImageWidth, data.ImageHeight = parseInt32("480"), parseInt32("360")
	}
	data.SiteName = nonEmpty("YouTube")
	return data
}

type oEmbedResponse struct {
	Title           string `json:"title"`
	AuthorName      string `json:"author_name"`
	ThumbnailURL    string `json:"thumbnail_url"`
	ThumbnailWidth  int32  `json:"thumbnail_width"`
	ThumbnailHeight int32  `json:"thumbnail_height"`
}

func (s *OGPService) applyOEmbed(ctx context.Context, watchURL *url.URL, data *entity.OGPData) {
	endpoint := "https://www.youtube.com/oembed?" + url.Values{"format": {"json"}, "url": {watchURL.String()}}.Encode()
	resp, err := s.get(ctx, endpoint)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var oembed oEmbedResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxOGPBodyBytes)).Decode(&oembed); err != nil {
		return
	}
	data.Title = nonEmpty(oembed.Title)
	data.YouTube.ChannelName = nonEmpty(oembed.AuthorName)
	if data.ImageURL = nonEmpty(oembed.ThumbnailURL); data.ImageURL != nil {
		data.ImageWidth, data.ImageHeight = &oembed.ThumbnailWidth, &oembed.ThumbnailHeight
	}
}

// parseISODuration は PT1H2M3S 形式の再生時間を秒にします。ライブ配信の P0D などは nil を返します
func parseISODuration(s string) *int32 {
	m := isoDurationPattern.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	total := 0
	for i, unit := range []int{3600, 60, 1} {
		n, _ := strconv.Atoi(m[i+1])
		total += n * unit
	}
	if total == 0 {
		return nil
	}
	seconds := int32(total)
	return &seconds
}
