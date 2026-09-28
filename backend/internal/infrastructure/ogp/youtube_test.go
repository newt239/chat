package ogp

import (
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestYouTubeVideoID(t *testing.T) {
	tests := []struct {
		url    string
		want   string
		wantOK bool
	}{
		{url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=42", want: "dQw4w9WgXcQ", wantOK: true},
		{url: "https://youtu.be/dQw4w9WgXcQ?si=abc", want: "dQw4w9WgXcQ", wantOK: true},
		{url: "https://m.youtube.com/shorts/dQw4w9WgXcQ", want: "dQw4w9WgXcQ", wantOK: true},
		{url: "https://www.youtube.com/embed/dQw4w9WgXcQ", want: "dQw4w9WgXcQ", wantOK: true},
		{url: "https://www.youtube.com/results?search_query=react", want: ""},
		{url: "https://www.youtube.com/watch?v=short", want: "short"},
		{url: "https://example.com/watch?v=dQw4w9WgXcQ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			u, _ := url.Parse(tt.url)
			got, ok := youTubeVideoID(u)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("got=(%q, %t) want=(%q, %t)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestParseISODuration(t *testing.T) {
	tests := map[string]int32{"PT3M34S": 214, "PT1H2M3S": 3723, "PT45S": 45}
	for input, want := range tests {
		if got := parseISODuration(input); got == nil || *got != want {
			t.Errorf("%s: got=%v want=%d", input, got, want)
		}
	}
	for _, input := range []string{"P0D", "", "3:34"} {
		if got := parseISODuration(input); got != nil {
			t.Errorf("%s は nil を返すべきです: %d", input, *got)
		}
	}
}

func TestBuildOGPDataFromYouTubePage(t *testing.T) {
	page := `<html><head><title>fallback - YouTube</title>
<meta property="og:site_name" content="YouTube">
<meta property="og:title" content="動画のタイトル">
<meta property="og:image" content="https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg">
<meta property="og:image:width" content="1280"><meta property="og:image:height" content="720">
<meta name="twitter:card" content="player">
</head><body><div itemscope itemtype="http://schema.org/VideoObject">
<meta itemprop="name" content="動画のタイトル"><meta itemprop="duration" content="PT24M18S">
<span itemprop="author" itemscope><link itemprop="url" href="http://www.youtube.com/@channel"><link itemprop="name" content="チャンネル名"></span>
<link itemprop="name" content="投稿者ではない名前">
</div></body></html>`
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	meta := map[string]string{}
	collectMeta(doc, meta, false)
	base, _ := url.Parse("https://www.youtube.com/watch?v=dQw4w9WgXcQ")

	data := buildOGPData(meta, base)

	if *data.Title != "動画のタイトル" || *data.CardType != "player" || *data.ImageWidth != 1280 || *data.ImageHeight != 720 {
		t.Errorf("OGP が期待と異なります: title=%s card=%s", *data.Title, *data.CardType)
	}
	if meta["author:name"] != "チャンネル名" {
		t.Errorf("投稿者名が期待と異なります: %q", meta["author:name"])
	}
	if got := parseISODuration(meta["itemprop:duration"]); got == nil || *got != 24*60+18 {
		t.Errorf("再生時間が期待と異なります: %v", got)
	}
}

func TestBuildOGPDataFallsBackToTitleAndResolvesRelativeImage(t *testing.T) {
	doc, _ := html.Parse(strings.NewReader(`<html><head><title> ページ </title><meta name="description" content="説明"><meta property="og:image" content="/ogp.png"></head></html>`))
	meta := map[string]string{}
	collectMeta(doc, meta, false)
	base, _ := url.Parse("https://example.com/blog/post")

	data := buildOGPData(meta, base)

	if *data.Title != "ページ" || *data.Description != "説明" || *data.ImageURL != "https://example.com/ogp.png" {
		t.Errorf("OGP が期待と異なります: %s %s %s", *data.Title, *data.Description, *data.ImageURL)
	}
	if data.YouTube != nil || data.ImageWidth != nil {
		t.Error("YouTube 以外のページに動画情報や寸法が入っています")
	}
}
