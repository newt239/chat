package rpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

// fromProto は presenter の対応表を逆に引いてリクエストの列挙値を変換します。見つからなければゼロ値を返します
func fromProto[K, V comparable](m map[K]V, value V) K {
	for k, v := range m {
		if v == value {
			return k
		}
	}
	var zero K
	return zero
}

func optionalTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	return new(t.AsTime())
}

func locationInput(l *chatv1.MessageLocation) *entity.MessageLocation {
	if l == nil {
		return nil
	}
	return &entity.MessageLocation{Latitude: l.Latitude, Longitude: l.Longitude, AccuracyMeters: l.AccuracyMeters, Label: l.Label}
}
