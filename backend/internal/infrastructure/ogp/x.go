package ogp

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
)

var (
	xHosts        = map[string]bool{"x.com": true, "www.x.com": true, "twitter.com": true, "www.twitter.com": true, "mobile.twitter.com": true, "mobile.x.com": true}
	xStatusPath   = regexp.MustCompile(`^/[A-Za-z0-9_]{1,15}/status/\d+`)
	xTitlePattern = regexp.MustCompile(`^(.*) \(@([A-Za-z0-9_]{1,15})\) on X$`)
	xSiteName     = "X"
)

func isXPostURL(u *url.URL) bool {
	return xHosts[strings.ToLower(u.Hostname())] && xStatusPath.MatchString(u.Path)
}

// applyXPost は X が返す「名前 (@ID) on X」形式のタイトルから投稿者を取り出します。削除済みなどで形式が違えば通常の OGP のままにします
func applyXPost(data *entity.OGPData) {
	if data.Title == nil {
		return
	}
	m := xTitlePattern.FindStringSubmatch(*data.Title)
	if m == nil {
		return
	}
	data.XPost = &entity.XPost{AuthorName: m[1], AuthorHandle: m[2]}
	data.SiteName = nonEmpty(xSiteName)
}
