package reaction

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubMessageRepo struct {
	domainrepository.MessageRepository
}

func (stubMessageRepo) FindByID(_ context.Context, id string) (*entity.Message, error) {
	return &entity.Message{ID: id, ChannelID: "ch1"}, nil
}

func (stubMessageRepo) AddReaction(_ context.Context, _ *entity.MessageReaction) error {
	return nil
}

func (stubMessageRepo) RemoveReaction(_ context.Context, _, _, _ string) error {
	return nil
}

type stubChannelRepo struct {
	domainrepository.ChannelRepository
}

func (stubChannelRepo) FindByID(_ context.Context, id string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "ws"}, nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: "Alice"}, nil
}

type stubChannelAccess struct{}

func (stubChannelAccess) EnsureChannelAccess(_ context.Context, channelID string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: channelID}, nil
}

type recordingNotifier struct {
	added   []ReactionNotification
	removed []ReactionNotification
}

func (n *recordingNotifier) NotifyReactionAdded(_, _ string, reaction ReactionNotification) {
	n.added = append(n.added, reaction)
}

func (n *recordingNotifier) NotifyReactionRemoved(_, _ string, reaction ReactionNotification) {
	n.removed = append(n.removed, reaction)
}

func TestReactionNotificationsIncludeUserOnAdd(t *testing.T) {
	notifier := &recordingNotifier{}
	uc := NewReactionInteractor(stubMessageRepo{}, stubChannelRepo{}, nil, nil, stubUserRepo{}, notifier, stubChannelAccess{})
	input := AddReactionInput{MessageID: "m1", UserID: "u1", Emoji: "👍"}

	if err := uc.AddReaction(context.Background(), input); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if err := uc.RemoveReaction(context.Background(), RemoveReactionInput(input)); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	added := notifier.added[0]
	if added.User == nil || added.User.DisplayName != "Alice" || added.CreatedAt.IsZero() {
		t.Errorf("追加の通知に押したユーザーと日時が含まれていません: %+v", added)
	}
	if removed := notifier.removed[0]; removed.User != nil || removed.UserID != "u1" {
		t.Errorf("削除の通知が期待と異なります: %+v", removed)
	}
}
