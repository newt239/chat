package http

import (
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

const specPath = "../../../../../openapi/bundled.yaml"

// パラメータ名の違いを無視して比較するため、プレースホルダを {} に正規化する
var (
	openapiParam = regexp.MustCompile(`\{[^/]+\}`)
	echoParam    = regexp.MustCompile(`:[^/]+`)
)

// TestRoutesMatchOpenAPI は登録済みルートと OpenAPI 定義の過不足を検出します
func TestRoutesMatchOpenAPI(t *testing.T) {
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("OpenAPI スキーマを読み込めませんでした: %v", err)
	}

	spec := make(map[string]struct{})
	for path, item := range doc.Paths.Map() {
		for method := range item.Operations() {
			spec[method+" "+openapiParam.ReplaceAllString(path, "{}")] = struct{}{}
		}
	}

	registered := make(map[string]struct{})
	for _, route := range NewRouter(RouterConfig{}).Routes() {
		if route.Path == "/ws" || strings.Contains(route.Path, "*") || strings.HasPrefix(route.Method, "echo_route_") {
			continue
		}
		registered[route.Method+" "+echoParam.ReplaceAllString(route.Path, "{}")] = struct{}{}
	}

	for key := range spec {
		if _, ok := registered[key]; !ok {
			t.Errorf("OpenAPI に定義されているがルート未登録: %s", key)
		}
	}
	for key := range registered {
		if _, ok := spec[key]; !ok {
			t.Errorf("ルート登録されているが OpenAPI に定義なし: %s", key)
		}
	}
}
