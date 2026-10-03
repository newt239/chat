package httpapi

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/newt239/chat/internal/domain/service"
	imageuc "github.com/newt239/chat/internal/usecase/image"
)

// imageHandler はアイコン画像の期限のない URL を、ストレージの期限付き URL へリダイレクトします
func imageHandler(storage service.StorageService) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Param("*")
		if path == "" || strings.Contains(path, "..") {
			return c.NoContent(http.StatusNotFound)
		}
		url, err := storage.GenerateDownloadURL(c.Request().Context(), imageuc.KeyPrefix+path, service.DownloadURLExpires)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		// 期限付き URL より先に切れるよう、リダイレクトのキャッシュは短くする
		c.Response().Header().Set("Cache-Control", "private, max-age=60")
		return c.Redirect(http.StatusFound, url)
	}
}
