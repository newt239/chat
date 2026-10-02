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

// reverseLookup は値から map のキーを引きます。見つからなければゼロ値を返します
func reverseLookup[K, V comparable](m map[K]V, value V) K {
	for k, v := range m {
		if v == value {
			return k
		}
	}
	var zero K
	return zero
}

// ConvertAll はスライスの各要素を変換します
func ConvertAll[T, U any](items []T, convert func(T) U) []U {
	converted := make([]U, 0, len(items))
	for _, item := range items {
		converted = append(converted, convert(item))
	}
	return converted
}
