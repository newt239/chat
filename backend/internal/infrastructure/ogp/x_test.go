package ogp

import (
	"net/url"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestIsXPostURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{url: "https://x.com/jack/status/20", want: true},
		{url: "https://twitter.com/jack/status/20?s=20", want: true},
		{url: "https://mobile.twitter.com/jack/status/20/photo/1", want: true},
		{url: "https://x.com/jack", want: false},
		{url: "https://example.com/jack/status/20", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			u, _ := url.Parse(tt.url)
			if got := isXPostURL(u); got != tt.want {
				t.Errorf("got=%t want=%t", got, tt.want)
			}
		})
	}
}

func TestApplyXPost(t *testing.T) {
	data := &entity.OGPData{Title: nonEmpty("jack (@jack) on X")}
	applyXPost(data)
	if data.XPost == nil || data.XPost.AuthorName != "jack" || data.XPost.AuthorHandle != "jack" {
		t.Fatalf("投稿者を取り出せていません: %+v", data.XPost)
	}

	notFound := &entity.OGPData{Title: nonEmpty("Post Not Found - X | 404 Error")}
	applyXPost(notFound)
	if notFound.XPost != nil {
		t.Fatal("見つからない投稿を X の投稿として扱っています")
	}
}
