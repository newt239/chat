package reaction

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type stubMessageRepo struct {
	domainrepository.MessageRepository
}

func (stubMessageRepo) AddReaction(_ context.Context, r *entity.MessageReaction) error {
	r.CreatedAt = time.Now()
	return nil
}

func (stubMessageRepo) RemoveReaction(_ context.Context, _, _, _ string) error {
	return nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: "Alice"}, nil
}

type stubChannelAccess struct {
	service.ChannelAccessService
}

func (stubChannelAccess) EnsureMessageAccess(_ context.Context, messageID, _ string) (*entity.Message, *entity.Channel, error) {
	return &entity.Message{ID: messageID, ChannelID: "ch1"}, &entity.Channel{ID: "ch1"}, nil
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
	uc := New(stubMessageRepo{}, stubUserRepo{}, notifier, stubChannelAccess{})
	input := ReactionInput{MessageID: "m1", UserID: "u1", Emoji: "👍"}

	if err := uc.AddReaction(context.Background(), input); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if err := uc.RemoveReaction(context.Background(), input); err != nil {
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
