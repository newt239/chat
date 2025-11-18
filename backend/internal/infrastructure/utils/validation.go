package utils

import (
	"github.com/labstack/echo/v4"
)

// GetParamFromContext はパラメータを取得し、空の場合はエラーを返します
func GetParamFromContext(c echo.Context, paramName string) (string, error) {
	param := c.Param(paramName)
	if param == "" {
		return "", HandleParamError(paramName)
	}
	return param, nil
}
