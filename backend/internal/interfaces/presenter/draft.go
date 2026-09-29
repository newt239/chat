package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

func Draft(d *entity.Draft) *chatv1.Draft {
	return &chatv1.Draft{Id: d.ID, ChannelId: d.ChannelID, ParentId: d.ParentID, Body: d.Body, UpdatedAt: timestamppb.New(d.UpdatedAt)}
}

func OptionalDraft(d *entity.Draft) *chatv1.Draft {
	if d == nil {
		return nil
	}
	return Draft(d)
}
