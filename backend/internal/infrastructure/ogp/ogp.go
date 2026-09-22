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
	"strings"
	"time"

	"golang.org/x/net/html"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/domain/service"
)

type OGPData struct {
	Title       *string
	Description *string
	ImageURL    *string
	SiteName    *string
	CardType    *string
}

type OGPService struct {
	httpClient *http.Client
}

// ErrBlockedAddress は内部ネットワーク宛のリクエストを拒否したことを表します
var ErrBlockedAddress = errors.New("内部ネットワーク宛の URL は取得できません")

const (
	maxOGPRedirects = 3
	maxOGPBodyBytes = 1024 * 1024
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

func (s *OGPService) FetchOGP(ctx context.Context, urlStr string) (*service.OGPData, error) {
	// URLの検証
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// HTTPリクエストの作成
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// User-Agentを設定（一部サイトでブロックされるのを防ぐ）
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ChatApp/1.0; +https://example.com)")

	// リクエスト実行
	resp, err := s.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, ErrBlockedAddress) {
			return nil, fmt.Errorf("%w: %s", domainerrors.ErrValidation, ErrBlockedAddress.Error())
		}
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			_ = err // エラーログは出力しない（既にレスポンスを読み取った後なので）
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Content-Typeの確認
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// HTMLの読み込み（最初の1MBまで）
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxOGPBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// HTMLの解析
	return s.parseHTML(string(body), parsedURL), nil
}

func (s *OGPService) parseHTML(htmlContent string, baseURL *url.URL) *service.OGPData {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return &service.OGPData{}
	}

	ogpData := &service.OGPData{}
	s.extractMetaTags(doc, ogpData, baseURL)
	return ogpData
}

func (s *OGPService) extractMetaTags(n *html.Node, ogpData *service.OGPData, baseURL *url.URL) {
	if n.Type == html.ElementNode && n.Data == "meta" {
		// metaタグの属性を取得
		attrs := make(map[string]string)
		for _, attr := range n.Attr {
			attrs[attr.Key] = attr.Val
		}

		// Twitter Card メタタグの処理
		if name, ok := attrs["name"]; ok && strings.HasPrefix(name, "twitter:") {
			content := attrs["content"]
			switch name {
			case "twitter:title":
				if ogpData.Title == nil {
					ogpData.Title = &content
				}
			case "twitter:description":
				if ogpData.Description == nil {
					ogpData.Description = &content
				}
			case "twitter:image":
				if ogpData.ImageURL == nil {
					ogpData.ImageURL = s.resolveURL(content, baseURL)
				}
			case "twitter:card":
				ogpData.CardType = &content
			}
		}

		// Open Graph メタタグの処理
		if property, ok := attrs["property"]; ok && strings.HasPrefix(property, "og:") {
			content := attrs["content"]
			switch property {
			case "og:title":
				if ogpData.Title == nil {
					ogpData.Title = &content
				}
			case "og:description":
				if ogpData.Description == nil {
					ogpData.Description = &content
				}
			case "og:image":
				if ogpData.ImageURL == nil {
					ogpData.ImageURL = s.resolveURL(content, baseURL)
				}
			case "og:site_name":
				ogpData.SiteName = &content
			}
		}

		// 標準HTMLメタタグの処理
		if name, ok := attrs["name"]; ok {
			content := attrs["content"]
			switch name {
			case "description":
				if ogpData.Description == nil {
					ogpData.Description = &content
				}
			}
		}
	}

	// titleタグの処理
	if n.Type == html.ElementNode && n.Data == "title" && ogpData.Title == nil {
		if n.FirstChild != nil {
			title := strings.TrimSpace(n.FirstChild.Data)
			ogpData.Title = &title
		}
	}

	// 子ノードを再帰的に処理
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.extractMetaTags(c, ogpData, baseURL)
	}
}

func (s *OGPService) resolveURL(urlStr string, baseURL *url.URL) *string {
	if urlStr == "" {
		return nil
	}

	// 絶対URLの場合はそのまま返す
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		return &urlStr
	}

	// 相対URLの場合はbaseURLと結合
	resolvedURL, err := baseURL.Parse(urlStr)
	if err != nil {
		return nil
	}

	resolvedStr := resolvedURL.String()
	return &resolvedStr
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
