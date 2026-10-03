package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const (
	userID    = "11111111-1111-4111-8111-111111111111"
	otherID   = "22222222-2222-4222-8222-222222222222"
	channelID = "33333333-3333-4333-8333-333333333333"
	botID     = "bot"
	dmID      = "dm"
)

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelMember(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "ws"}, nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, Preferences: entity.UserPreferences{Timezone: "Asia/Tokyo"}}, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, id string) (*entity.WorkspaceMember, error) {
	return &entity.WorkspaceMember{UserID: id}, nil
}

type stubChannelRepo struct {
	domainrepository.ChannelRepository
	dmWith []string
}

func (r *stubChannelRepo) FindOrCreateDM(_ context.Context, _ string, a, b string) (*entity.Channel, error) {
	r.dmWith = []string{a, b}
	return &entity.Channel{ID: dmID}, nil
}

type fakeReminderRepo struct {
	domainrepository.ReminderRepository
	created []*entity.Reminder
	sent    []string
}

func (r *fakeReminderRepo) Create(_ context.Context, rem *entity.Reminder) error {
	rem.ID = "r1"
	r.created = append(r.created, rem)
	return nil
}

func (r *fakeReminderRepo) ClaimDue(context.Context, time.Time, time.Time, int) ([]*entity.Reminder, error) {
	return r.created, nil
}

func (r *fakeReminderRepo) MarkSent(_ context.Context, id string) error {
	r.sent = append(r.sent, id)
	return nil
}

type post struct {
	channelID string
	body      string
}

type fakePoster struct {
	posts []post
}

func (*fakePoster) EnsureOfficial(context.Context, string) (*entity.App, error) {
	return &entity.App{BotUserID: botID, IsOfficial: true}, nil
}

func (p *fakePoster) PostAsOfficial(_ context.Context, _ string, channelID string, _ *string, body string) (*messageuc.MessageOutput, error) {
	p.posts = append(p.posts, post{channelID: channelID, body: body})
	return &messageuc.MessageOutput{ChannelID: channelID, Body: body}, nil
}

func newInteractor() (*Interactor, *fakeReminderRepo, *stubChannelRepo, *fakePoster) {
	reminders, channels, poster := &fakeReminderRepo{}, &stubChannelRepo{}, &fakePoster{}
	i := New(reminders, stubUserRepo{}, stubWorkspaceRepo{}, channels, stubAccess{}, poster)
	i.now = func() time.Time { return time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC) }
	return i, reminders, channels, poster
}

func TestRemindCreatesReminderAndDeliversToDM(t *testing.T) {
	i, reminders, channels, poster := newInteractor()
	ctx := context.Background()

	out, err := i.Execute(ctx, ExecuteInput{UserID: userID, ChannelID: channelID, Text: "/remind <@" + otherID + "> 資料を送る 明日 9:00"})
	if err != nil {
		t.Fatal(err)
	}
	// 実行者のタイムゾーン（JST）で解釈する
	if got := reminders.created[0]; got.Text != "資料を送る" || !got.RemindAt.Equal(time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)) || *got.TargetUserID != otherID {
		t.Fatalf("リマインダーが期待と異なります: %+v", got)
	}
	if out.ChannelID != channelID || !strings.Contains(out.Body, "2026/10/02 09:00 JST") {
		t.Fatalf("受付の応答がチャンネルに投稿されていません: %+v", out)
	}

	if n, err := i.DispatchDue(ctx); err != nil || n != 1 {
		t.Fatalf("届けられません: %d %v", n, err)
	}
	delivered := poster.posts[1]
	if delivered.channelID != dmID || channels.dmWith[0] != botID || channels.dmWith[1] != otherID || !strings.Contains(delivered.body, "<@"+otherID+"> リマインダー: 資料を送る") {
		t.Fatalf("宛先との DM に届いていません: %+v %v", delivered, channels.dmWith)
	}
	if len(reminders.sent) != 1 {
		t.Fatal("送信済みになっていません")
	}
}

func TestUnknownCommand(t *testing.T) {
	i, _, _, _ := newInteractor()
	if _, err := i.Execute(context.Background(), ExecuteInput{UserID: userID, ChannelID: channelID, Text: "/nope"}); !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("不明なコマンドを拒否していません: %v", err)
	}
}
