package ogp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
)

type OGPService struct {
	httpClient *http.Client
}

// ErrBlockedAddress は内部ネットワーク宛のリクエストを拒否したことを表します
var ErrBlockedAddress = errors.New("内部ネットワーク宛の URL は取得できません")

const (
	maxOGPRedirects = 3
	// YouTube の動画ページはメタデータが 700KB 付近にあるため余裕を持たせる
	maxOGPBodyBytes = 2 * 1024 * 1024
	userAgent       = "Mozilla/5.0 (compatible; ChatApp/1.0; +https://example.com)"
)

// isBlockedIP はループバック・プライベート・リンクローカル（クラウドのメタデータ含む）を弾きます
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

func NewOGPService() *OGPService {
	dialer := &net.Dialer{Timeout: 5 * time.Second}

	return &OGPService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxOGPRedirects {
					return fmt.Errorf("リダイレクトが多すぎます")
				}
				if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
					return fmt.Errorf("unsupported URL scheme: %s", req.URL.Scheme)
				}
				return nil
			},
			// 名前解決後のアドレスを検証し、DNS リバインディングによる迂回も防ぐ
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					host, _, err := net.SplitHostPort(addr)
					if err != nil {
						return nil, err
					}
					ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
					if err != nil {
						return nil, err
					}
					for _, ip := range ips {
						if isBlockedIP(ip) {
							return nil, ErrBlockedAddress
						}
					}
					return dialer.DialContext(ctx, network, addr)
				},
			},
		},
	}
}

func (s *OGPService) FetchOGP(ctx context.Context, urlStr string) (*entity.OGPData, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	if videoID, ok := youTubeVideoID(parsedURL); ok {
		return s.fetchYouTube(ctx, videoID), nil
	}

	meta, err := s.fetchMeta(ctx, parsedURL)
	if err != nil {
		return nil, err
	}
	return buildOGPData(meta, parsedURL), nil
}

// get は GET リクエストを送り、200 以外や内部ネットワーク宛をエラーにします
func (s *OGPService) get(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, ErrBlockedAddress) {
			return nil, fmt.Errorf("%w: %s", domainerrors.ErrValidation, ErrBlockedAddress.Error())
		}
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}
	return resp, nil
}

func (s *OGPService) fetchMeta(ctx context.Context, pageURL *url.URL) (map[string]string, error) {
	resp, err := s.get(ctx, pageURL.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	doc, err := html.Parse(io.LimitReader(resp.Body, maxOGPBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	meta := map[string]string{}
	collectMeta(doc, meta, false)
	return meta, nil
}

// collectMeta は <meta>・<title>・投稿者の microdata を最初に出現した値だけ集めます
func collectMeta(n *html.Node, meta map[string]string, inAuthor bool) {
	if n.Type == html.ElementNode {
		attrs := make(map[string]string, len(n.Attr))
		for _, attr := range n.Attr {
			attrs[attr.Key] = attr.Val
		}
		switch n.Data {
		case "meta":
			key := attrs["property"]
			if key == "" {
				key = attrs["name"]
			}
			if key == "" && attrs["itemprop"] != "" {
				key = "itemprop:" + attrs["itemprop"]
			}
			setOnce(meta, key, attrs["content"])
		case "link":
			if inAuthor && attrs["itemprop"] == "name" {
				setOnce(meta, "author:name", attrs["content"])
			}
		case "title":
			if n.FirstChild != nil {
				setOnce(meta, "title", strings.TrimSpace(n.FirstChild.Data))
			}
		}
		inAuthor = inAuthor || attrs["itemprop"] == "author"
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectMeta(c, meta, inAuthor)
	}
}

func setOnce(meta map[string]string, key, value string) {
	if _, exists := meta[key]; key != "" && value != "" && !exists {
		meta[key] = value
	}
}

func buildOGPData(meta map[string]string, baseURL *url.URL) *entity.OGPData {
	return &entity.OGPData{
		Title:       firstOf(meta, "og:title", "twitter:title", "title"),
		Description: firstOf(meta, "og:description", "twitter:description", "description"),
		ImageURL:    resolveURL(firstOf(meta, "og:image", "og:image:url", "twitter:image"), baseURL),
		SiteName:    firstOf(meta, "og:site_name"),
		CardType:    firstOf(meta, "twitter:card"),
		ImageWidth:  parseInt32(meta["og:image:width"]),
		ImageHeight: parseInt32(meta["og:image:height"]),
	}
}

func firstOf(meta map[string]string, keys ...string) *string {
	for _, key := range keys {
		if value, ok := meta[key]; ok {
			return &value
		}
	}
	return nil
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseInt32(s string) *int32 {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil || v <= 0 {
		return nil
	}
	n := int32(v)
	return &n
}

func resolveURL(urlStr *string, baseURL *url.URL) *string {
	if urlStr == nil {
		return nil
	}
	resolvedURL, err := baseURL.Parse(*urlStr)
	if err != nil {
		return nil
	}
	resolved := resolvedURL.String()
	return &resolved
}

// ExtractURLs はテキストからURLを抽出します
func (s *OGPService) ExtractURLs(text string) []string {
	return ExtractURLs(text)
}

// URLを抽出する正規表現
var urlRegex = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

func ExtractURLs(text string) []string {
	matches := urlRegex.FindAllString(text, -1)

	// 重複を除去
	urlSet := make(map[string]bool)
	var uniqueURLs []string

	for _, match := range matches {
		if !urlSet[match] {
			urlSet[match] = true
			uniqueURLs = append(uniqueURLs, match)
		}
	}

	return uniqueURLs
}
