package message

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id}, nil
}

func (stubAccess) AccessibleDescendants(_ context.Context, _ *entity.Channel, _ string) ([]*entity.Channel, error) {
	return []*entity.Channel{{ID: "child"}}, nil
}

type recordingMessageRepo struct {
	domainrepository.MessageRepository
	channelIDs []string
}

func (r *recordingMessageRepo) FindByChannelIDs(_ context.Context, channelIDs []string, _ int, _ *time.Time, _ *time.Time) ([]*entity.Message, error) {
	r.channelIDs = channelIDs
	return nil, nil
}

type recordingSystemMessageRepo struct {
	domainrepository.SystemMessageRepository
	channelIDs []string
}

func (r *recordingSystemMessageRepo) FindByChannelIDs(_ context.Context, channelIDs []string, _ int, _ *time.Time, _ *time.Time) ([]*entity.SystemMessage, error) {
	r.channelIDs = channelIDs
	return nil, nil
}

func TestListMessagesIncludeDescendants(t *testing.T) {
	tests := []struct {
		name               string
		includeDescendants bool
		want               []string
	}{
		{name: "子孫を含める", includeDescendants: true, want: []string{"parent", "child"}},
		{name: "子孫を含めない", includeDescendants: false, want: []string{"parent"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := &recordingMessageRepo{}
			systemMessages := &recordingSystemMessageRepo{}
			lister := &MessageLister{
				messageRepo:      messages,
				systemMsgRepo:    systemMessages,
				channelAccessSvc: stubAccess{},
				outputBuilder:    &MessageOutputBuilder{},
			}

			if _, err := lister.ListMessages(context.Background(), ListMessagesInput{ChannelID: "parent", UserID: "u", IncludeDescendants: tt.includeDescendants}); err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(messages.channelIDs, tt.want) || !slices.Equal(systemMessages.channelIDs, tt.want) {
				t.Fatalf("取得対象のチャンネルが正しくありません: messages=%v system=%v", messages.channelIDs, systemMessages.channelIDs)
			}
		})
	}
}
