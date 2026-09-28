// Package presenter はユースケースの出力を API のメッセージに変換します
package presenter

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func optionalTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// ConvertAll はスライスの各要素を変換します
func ConvertAll[T, U any](items []T, convert func(T) U) []U {
	converted := make([]U, 0, len(items))
	for _, item := range items {
		converted = append(converted, convert(item))
	}
	return converted
}
