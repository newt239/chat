package webhook

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const (
	channelID  = "channel"
	dmID       = "dm"
	archivedID = "archived"
	creatorID  = "creator"
	otherID    = "other"
	adminID    = "admin"
	outsiderID = "outsider"
)

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, userID string) (*entity.Channel, error) {
	if userID == outsiderID {
		return nil, domerr.ErrUnauthorized
	}
	ch := &entity.Channel{ID: id, WorkspaceID: "ws", Name: "general", Type: entity.ChannelTypePublic}
	switch id {
	case dmID:
		ch.Type = entity.ChannelTypeDM
	case archivedID:
		now := time.Now()
		ch.ArchivedAt = &now
	}
	return ch, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	role := entity.WorkspaceRoleMember
	if userID == adminID {
		role = entity.WorkspaceRoleAdmin
	}
	return &entity.WorkspaceMember{UserID: userID, Role: role}, nil
}

type fakeUserRepo struct {
	domainrepository.UserRepository
	users map[string]*entity.User
}

func (r *fakeUserRepo) Create(_ context.Context, u *entity.User) error {
	u.ID = "bot-" + u.DisplayName
	r.users[u.ID] = u
	return nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return r.users[id], nil
}

func (r *fakeUserRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.User, error) {
	found := []*entity.User{}
	for _, id := range ids {
		if u := r.users[id]; u != nil {
			found = append(found, u)
		}
	}
	return found, nil
}

func (r *fakeUserRepo) Update(_ context.Context, u *entity.User) error {
	r.users[u.ID] = u
	return nil
}

type fakeWebhookRepo struct {
	domainrepository.WebhookRepository
	webhooks map[string]*entity.Webhook
}

func (r *fakeWebhookRepo) FindByID(_ context.Context, id string) (*entity.Webhook, error) {
	if w, ok := r.webhooks[id]; ok {
		copied := *w
		return &copied, nil
	}
	return nil, nil
}

func (r *fakeWebhookRepo) FindByChannelID(_ context.Context, id string) ([]*entity.Webhook, error) {
	found := []*entity.Webhook{}
	for _, w := range r.webhooks {
		if w.ChannelID == id {
			found = append(found, w)
		}
	}
	return found, nil
}

func (r *fakeWebhookRepo) Create(_ context.Context, w *entity.Webhook) error {
	w.ID = "00000000-0000-0000-0000-00000000000" + string(rune('0'+len(r.webhooks)))
	r.webhooks[w.ID] = w
	return nil
}

func (r *fakeWebhookRepo) Update(_ context.Context, w *entity.Webhook) error {
	r.webhooks[w.ID] = w
	return nil
}

func (r *fakeWebhookRepo) MarkUsed(_ context.Context, id string, usedAt time.Time) error {
	r.webhooks[id].LastUsedAt = &usedAt
	return nil
}

func (r *fakeWebhookRepo) Delete(_ context.Context, id string) error {
	delete(r.webhooks, id)
	return nil
}

type fakePoster struct {
	posted []*entity.Message
}

func (p *fakePoster) CreateBotMessage(_ context.Context, ch *entity.Channel, m *entity.Message) (*messageuc.MessageOutput, error) {
	m.ChannelID = ch.ID
	p.posted = append(p.posted, m)
	return &messageuc.MessageOutput{ChannelID: ch.ID, UserID: m.UserID, Body: m.Body, CreatedAt: time.Now()}, nil
}

type stubTxManager struct{}

func (stubTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fixture struct {
	uc       *Interactor
	webhooks *fakeWebhookRepo
	users    *fakeUserRepo
	poster   *fakePoster
	recorder *audittest.Recorder
}

func newFixture() *fixture {
	f := &fixture{
		webhooks: &fakeWebhookRepo{webhooks: map[string]*entity.Webhook{}},
		users:    &fakeUserRepo{users: map[string]*entity.User{creatorID: {ID: creatorID, DisplayName: "Alice"}}},
		poster:   &fakePoster{},
		recorder: &audittest.Recorder{},
	}
	f.uc = NewInteractor(f.webhooks, f.users, stubWorkspaceRepo{}, stubAccess{}, f.poster, stubTxManager{}, f.recorder)
	return f
}

func (f *fixture) create(t *testing.T) *CreateOutput {
	t.Helper()
	out, err := f.uc.Create(context.Background(), CreateInput{ChannelID: channelID, UserID: creatorID, Name: " Deploy Bot "})
	if err != nil {
		t.Fatalf("Webhook を作成できません: %v", err)
	}
	return out
}

func TestCreateStoresHashAndBotUser(t *testing.T) {
	f := newFixture()
	out := f.create(t)

	stored := f.webhooks.webhooks[out.Webhook.ID]
	if out.Token == "" || stored.TokenHash == out.Token || !stored.VerifyToken(out.Token) {
		t.Fatalf("トークンがハッシュで保存されていません: %+v", stored)
	}
	bot := f.users.users[stored.BotUserID]
	if bot == nil || !bot.IsBot || bot.DisplayName != "Deploy Bot" {
		t.Fatalf("ボットユーザーが作成されていません: %+v", bot)
	}
	if !out.Webhook.CanManage || out.Webhook.CreatedBy.DisplayName != "Alice" {
		t.Fatalf("発行者の情報が正しくありません: %+v", out.Webhook)
	}
	if got := f.recorder.Actions(); len(got) != 1 || got[0] != entity.AuditActionWebhookCreated {
		t.Fatalf("作成が監査ログに記録されていません: %v", got)
	}
}

func TestCreateRejectsDMAndArchived(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	if _, err := f.uc.Create(ctx, CreateInput{ChannelID: dmID, UserID: creatorID, Name: "x"}); !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("DM への作成が拒否されていません: %v", err)
	}
	if _, err := f.uc.Create(ctx, CreateInput{ChannelID: archivedID, UserID: creatorID, Name: "x"}); !errors.Is(err, domerr.ErrChannelArchived) {
		t.Fatalf("アーカイブ済みチャンネルへの作成が拒否されていません: %v", err)
	}
	if _, err := f.uc.Create(ctx, CreateInput{ChannelID: channelID, UserID: outsiderID, Name: "x"}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("閲覧できないチャンネルへの作成が拒否されていません: %v", err)
	}
}

func TestManagePermission(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	out := f.create(t)

	list, err := f.uc.List(ctx, ListInput{ChannelID: channelID, UserID: otherID})
	if err != nil || len(list) != 1 || list[0].CanManage {
		t.Fatalf("発行者以外には編集不可として返すはず: %+v err=%v", list, err)
	}
	if _, err := f.uc.Update(ctx, UpdateInput{WebhookID: out.Webhook.ID, UserID: otherID, Name: "x"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("発行者以外が編集できてしまいます: %v", err)
	}
	if err := f.uc.Delete(ctx, TargetInput{WebhookID: out.Webhook.ID, UserID: otherID}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("発行者以外が削除できてしまいます: %v", err)
	}

	updated, err := f.uc.Update(ctx, UpdateInput{WebhookID: out.Webhook.ID, UserID: adminID, Name: "CI"})
	if err != nil || updated.Name != "CI" {
		t.Fatalf("管理者が編集できません: %+v err=%v", updated, err)
	}
	if bot := f.users.users[f.webhooks.webhooks[out.Webhook.ID].BotUserID]; bot.DisplayName != "CI" {
		t.Fatalf("ボットユーザーの表示名が更新されていません: %s", bot.DisplayName)
	}
	if err := f.uc.Delete(ctx, TargetInput{WebhookID: out.Webhook.ID, UserID: adminID}); err != nil {
		t.Fatalf("管理者が削除できません: %v", err)
	}
	if got := f.recorder.Actions(); got[len(got)-1] != entity.AuditActionWebhookDeleted {
		t.Fatalf("削除が監査ログに記録されていません: %v", got)
	}
}

func TestRegenerateTokenInvalidatesOldToken(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	out := f.create(t)

	token, err := f.uc.RegenerateToken(ctx, TargetInput{WebhookID: out.Webhook.ID, UserID: creatorID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.uc.Post(ctx, PostInput{WebhookID: out.Webhook.ID, Token: out.Token, Text: "hi"}); !errors.Is(err, ErrWebhookNotFound) {
		t.Fatalf("古いトークンで投稿できてしまいます: %v", err)
	}
	if _, err := f.uc.Post(ctx, PostInput{WebhookID: out.Webhook.ID, Token: token, Text: "hi"}); err != nil {
		t.Fatalf("新しいトークンで投稿できません: %v", err)
	}
}

func TestPost(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	out := f.create(t)
	id := out.Webhook.ID
	name := "  GitHub  "
	avatar := "https://example.com/icon.png"

	if _, err := f.uc.Post(ctx, PostInput{WebhookID: id, Token: out.Token, Text: " デプロイ完了 ", Username: &name, AvatarURL: &avatar}); err != nil {
		t.Fatalf("投稿できません: %v", err)
	}
	posted := f.poster.posted[0]
	stored := f.webhooks.webhooks[id]
	if posted.UserID != stored.BotUserID || posted.Body != "デプロイ完了" || posted.ChannelID != channelID {
		t.Fatalf("ボットユーザー名義で投稿されていません: %+v", posted)
	}
	if *posted.SenderName != "GitHub" || *posted.SenderAvatarURL != avatar {
		t.Fatalf("投稿ごとの表示名とアイコンが反映されていません: %+v", posted)
	}
	if stored.LastUsedAt == nil {
		t.Fatal("最終使用日時が記録されていません")
	}

	badAvatar := "javascript:alert(1)"
	cases := []struct {
		name  string
		input PostInput
		want  error
	}{
		{"ID が UUID でない", PostInput{WebhookID: "x", Token: out.Token, Text: "a"}, ErrWebhookNotFound},
		{"トークンが違う", PostInput{WebhookID: id, Token: "wrong", Text: "a"}, ErrWebhookNotFound},
		{"本文が空", PostInput{WebhookID: id, Token: out.Token, Text: "  "}, domerr.ErrValidation},
		{"本文が長すぎる", PostInput{WebhookID: id, Token: out.Token, Text: strings.Repeat("あ", MaxTextLength+1)}, domerr.ErrValidation},
		{"アイコンが http(s) でない", PostInput{WebhookID: id, Token: out.Token, Text: "a", AvatarURL: &badAvatar}, domerr.ErrValidation},
	}
	for _, tc := range cases {
		if _, err := f.uc.Post(ctx, tc.input); !errors.Is(err, tc.want) {
			t.Errorf("%s: %v を期待しましたが %v でした", tc.name, tc.want, err)
		}
	}
	if len(f.poster.posted) != 1 {
		t.Fatalf("不正な投稿が保存されています: %d", len(f.poster.posted))
	}
}

func TestPostRejectsWhenCreatorLostAccess(t *testing.T) {
	f := newFixture()
	out := f.create(t)
	f.webhooks.webhooks[out.Webhook.ID].CreatedBy = outsiderID

	if _, err := f.uc.Post(context.Background(), PostInput{WebhookID: out.Webhook.ID, Token: out.Token, Text: "a"}); !errors.Is(err, ErrInactive) {
		t.Fatalf("発行者がチャンネルを閲覧できない Webhook で投稿できてしまいます: %v", err)
	}
}
