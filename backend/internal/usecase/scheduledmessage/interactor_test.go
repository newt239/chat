package scheduledmessage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type fakeScheduledRepo struct {
	domainrepository.ScheduledMessageRepository
	due    []*entity.ScheduledMessage
	sent   map[string]string
	failed map[string]string
}

func (r *fakeScheduledRepo) ClaimDue(_ context.Context, _, _ time.Time, _ int) ([]*entity.ScheduledMessage, error) {
	return r.due, nil
}

func (r *fakeScheduledRepo) MarkSent(_ context.Context, id, messageID string) error {
	r.sent[id] = messageID
	return nil
}

func (r *fakeScheduledRepo) MarkFailed(_ context.Context, id, reason string) error {
	r.failed[id] = reason
	return nil
}

// fakeMessageUC は "lost" チャンネルへの投稿を権限なし、"broken" を内部エラーにする
type fakeMessageUC struct {
	created []messageuc.CreateMessageInput
}

func (m *fakeMessageUC) CreateMessage(_ context.Context, input messageuc.CreateMessageInput) (*messageuc.MessageOutput, error) {
	switch input.ChannelID {
	case "lost":
		return nil, domerr.ErrUnauthorized
	case "broken":
		return nil, errors.New("connection reset")
	}
	m.created = append(m.created, input)
	return &messageuc.MessageOutput{ID: "msg-" + input.ChannelID}, nil
}

type nopLogger struct{ service.Logger }

func (nopLogger) Error(string, ...service.LogField) {}

func TestDispatchDue(t *testing.T) {
	repo := &fakeScheduledRepo{
		due: []*entity.ScheduledMessage{
			{ID: "s1", UserID: "u1", ChannelID: "ch1", Body: "おはよう", AttachmentIDs: []string{"a1"}},
			{ID: "s2", UserID: "u1", ChannelID: "lost", Body: "x"},
			{ID: "s3", UserID: "u1", ChannelID: "broken", Body: "x"},
		},
		sent:   map[string]string{},
		failed: map[string]string{},
	}
	messages := &fakeMessageUC{}
	uc := NewInteractor(repo, nil, nil, nil, nil, messages, nopLogger{})

	count, err := uc.DispatchDue(context.Background())
	if err != nil || count != 3 {
		t.Fatalf("取り出した件数が期待と異なります: %d %v", count, err)
	}
	if repo.sent["s1"] != "msg-ch1" || len(messages.created) != 1 || messages.created[0].AttachmentIDs[0] != "a1" {
		t.Errorf("通常の投稿経路で送信されていません: %+v %+v", repo.sent, messages.created)
	}
	if repo.failed["s2"] != domerr.ErrUnauthorized.Error() {
		t.Errorf("権限を失った予約の失敗理由が期待と異なります: %q", repo.failed["s2"])
	}
	if repo.failed["s3"] != errSendFailed.Error() {
		t.Errorf("内部エラーの詳細を失敗理由に出しています: %q", repo.failed["s3"])
	}
}

func TestScheduleRejectsPastAndEmpty(t *testing.T) {
	uc := NewInteractor(nil, nil, nil, nil, nil, nil, nopLogger{})
	ctx := context.Background()

	_, err := uc.Schedule(ctx, ScheduleInput{UserID: "u1", ChannelID: "ch1", Body: "x", ScheduledAt: time.Now().Add(-time.Minute)})
	if !errors.Is(err, ErrScheduleInPast) {
		t.Errorf("過去の日時を拒否していません: %v", err)
	}
	_, err = uc.Schedule(ctx, ScheduleInput{UserID: "u1", ChannelID: "ch1", Body: " ", ScheduledAt: time.Now().Add(time.Hour)})
	if !errors.Is(err, messageuc.ErrEmptyMessage) {
		t.Errorf("空の予約を拒否していません: %v", err)
	}
}
