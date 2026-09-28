package message

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubAttachmentRepo struct {
	domainrepository.AttachmentRepository
	attachments []*entity.Attachment
}

func (r *stubAttachmentRepo) FindPendingByIDsForUser(
	_ context.Context,
	_ string,
	_ []string,
) ([]*entity.Attachment, error) {
	return r.attachments, nil
}

func TestVerifyAttachments(t *testing.T) {
	tests := []struct {
		name        string
		attachments []*entity.Attachment
		wantErr     error
	}{
		{
			name:        "本人のものでない添付は拒否する",
			attachments: []*entity.Attachment{},
			wantErr:     ErrAttachmentNotFound,
		},
		{
			name:        "別チャンネル宛の添付は拒否する",
			attachments: []*entity.Attachment{{ID: "a1", ChannelID: "other"}},
			wantErr:     ErrAttachmentNotFound,
		},
		{
			name:        "同じチャンネル宛の自分の添付は許可する",
			attachments: []*entity.Attachment{{ID: "a1", ChannelID: "ch1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &MessageCreator{attachmentRepo: &stubAttachmentRepo{attachments: tt.attachments}}
			input := CreateMessageInput{UserID: "u1", AttachmentIDs: []string{"a1"}}

			err := creator.verifyAttachments(context.Background(), input, "ch1")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
		})
	}
}
