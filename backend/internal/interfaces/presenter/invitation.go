package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
)

func Invitation(i invitationuc.InvitationOutput) *chatv1.Invitation {
	return &chatv1.Invitation{
		Id:            i.ID,
		Email:         i.Email,
		Role:          workspaceRoles[i.Role],
		InvitedByName: i.InvitedByName,
		ExpiresAt:     timestamppb.New(i.ExpiresAt),
		CreatedAt:     timestamppb.New(i.CreatedAt),
	}
}
