package app

import (
	"context"
	"errors"
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
	workspaceID = "ws"
	creatorID   = "creator"
	otherID     = "other"
	adminID     = "admin"
	// チャンネルの ID は投稿時に UUID として検証される
	joinedID  = "11111111-1111-4111-8111-111111111111"
	publicID  = "22222222-2222-4222-8222-222222222222"
	privateID = "33333333-3333-4333-8333-333333333333"
	parentID  = "44444444-4444-4444-8444-444444444444"
)

var channels = map[string]*entity.Channel{
	joinedID:  {ID: joinedID, WorkspaceID: workspaceID, Name: "joined", Type: entity.ChannelTypePublic},
	publicID:  {ID: publicID, WorkspaceID: workspaceID, Name: "public", Type: entity.ChannelTypePublic},
	privateID: {ID: privateID, WorkspaceID: workspaceID, Name: "private", Type: entity.ChannelTypePrivate},
}

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	if ch := channels[id]; ch != nil {
		return ch, nil
	}
	return nil, domerr.ErrChannelNotFound
}

func (s stubAccess) EnsureChannelMember(ctx context.Context, id string, userID string) (*entity.Channel, error) {
	return s.EnsureChannelAccess(ctx, id, userID)
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

func (stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	return &entity.Workspace{ID: id, CreatedBy: adminID}, nil
}

type stubChannelRepo struct {
	domainrepository.ChannelRepository
}

func (stubChannelRepo) FindByID(_ context.Context, id string) (*entity.Channel, error) {
	return channels[id], nil
}

type fakeMemberRepo struct {
	domainrepository.ChannelMemberRepository
	members map[string]bool
}

func (r *fakeMemberRepo) IsMember(_ context.Context, channelID, userID string) (bool, error) {
	return r.members[channelID+"/"+userID], nil
}

func (r *fakeMemberRepo) AddMember(_ context.Context, m *entity.ChannelMember) error {
	r.members[m.ChannelID+"/"+m.UserID] = true
	return nil
}

type stubMessageRepo struct {
	domainrepository.MessageRepository
}

func (stubMessageRepo) FindByID(_ context.Context, id string) (*entity.Message, error) {
	if id == parentID {
		return &entity.Message{ID: id, ChannelID: joinedID}, nil
	}
	return nil, nil
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

func (r *fakeUserRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.User, error) {
	found := []*entity.User{}
	for _, id := range ids {
		if u := r.users[id]; u != nil {
			found = append(found, u)
		}
	}
	return found, nil
}

type fakeAppRepo struct {
	domainrepository.AppRepository
	apps map[string]*entity.App
}

func (r *fakeAppRepo) FindByID(_ context.Context, id string) (*entity.App, error) {
	if a, ok := r.apps[id]; ok {
		copied := *a
		return &copied, nil
	}
	return nil, nil
}

func (r *fakeAppRepo) FindOfficial(_ context.Context, _ string) (*entity.App, error) {
	for _, a := range r.apps {
		if a.IsOfficial {
			return a, nil
		}
	}
	return nil, nil
}

func (r *fakeAppRepo) Create(_ context.Context, a *entity.App) error {
	a.ID = "00000000-0000-4000-8000-00000000000" + string(rune('0'+len(r.apps)))
	r.apps[a.ID] = a
	return nil
}

func (r *fakeAppRepo) MarkUsed(_ context.Context, id string, usedAt time.Time) error {
	r.apps[id].LastUsedAt = &usedAt
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
	apps     *fakeAppRepo
	users    *fakeUserRepo
	members  *fakeMemberRepo
	poster   *fakePoster
	recorder *audittest.Recorder
}

type nopLogger struct{ domainservice.Logger }

func (nopLogger) Warn(string, ...domainservice.LogField) {}

// post は着信 Webhook と同じくトークンを確かめてから投稿します
func (f *fixture) post(token string, out *CreateOutput, input PostInput) (*messageuc.MessageOutput, error) {
	app, err := f.uc.Authenticate(context.Background(), out.App.ID, token)
	if err != nil {
		return nil, err
	}
	return f.uc.Post(context.Background(), app, input)
}

func newFixture() *fixture {
	f := &fixture{
		apps:     &fakeAppRepo{apps: map[string]*entity.App{}},
		users:    &fakeUserRepo{users: map[string]*entity.User{creatorID: {ID: creatorID, DisplayName: "Alice"}}},
		members:  &fakeMemberRepo{members: map[string]bool{}},
		poster:   &fakePoster{},
		recorder: &audittest.Recorder{},
	}
	f.uc = NewInteractor(f.apps, f.users, stubWorkspaceRepo{}, stubChannelRepo{}, f.members, stubMessageRepo{}, stubAccess{}, f.poster, stubTxManager{}, f.recorder, nopLogger{})
	return f
}

func (f *fixture) create(t *testing.T, permissions ...entity.AppPermission) *CreateOutput {
	t.Helper()
	defaultChannel := joinedID
	out, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, UserID: creatorID, Settings: SettingsInput{
		Name:             " Deploy Bot ",
		Permissions:      permissions,
		DefaultChannelID: &defaultChannel,
	}})
	if err != nil {
		t.Fatalf("アプリを作成できません: %v", err)
	}
	return out
}

func TestCreateStoresHashBotUserAndJoinsDefaultChannel(t *testing.T) {
	f := newFixture()
	out := f.create(t, entity.AppPermissionPostJoinedChannels)

	stored := f.apps.apps[out.App.ID]
	if out.Token == "" || *stored.TokenHash == out.Token || !stored.VerifyToken(out.Token) {
		t.Fatalf("トークンがハッシュで保存されていません: %+v", stored)
	}
	bot := f.users.users[stored.BotUserID]
	if bot == nil || !bot.IsApp || bot.IsOfficial || bot.DisplayName != "Deploy Bot" {
		t.Fatalf("ボットユーザーが作成されていません: %+v", bot)
	}
	if !f.members.members[joinedID+"/"+stored.BotUserID] {
		t.Fatal("既定のチャンネルに参加していません")
	}
	if got := f.recorder.Actions(); len(got) != 1 || got[0] != entity.AuditActionAppCreated {
		t.Fatalf("作成が監査ログに記録されていません: %v", got)
	}
}

func TestCreateValidatesOutgoingWebhook(t *testing.T) {
	f := newFixture()
	_, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, UserID: creatorID, Settings: SettingsInput{
		Name:        "x",
		Permissions: []entity.AppPermission{entity.AppPermissionOutgoingWebhook},
	}})
	if !errors.Is(err, ErrOutgoingURLMissing) {
		t.Fatalf("送信先のない送信 Webhook を拒否していません: %v", err)
	}

	url := "https://example.com/hook"
	out, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, UserID: creatorID, Settings: SettingsInput{
		Name:        "x",
		Permissions: []entity.AppPermission{entity.AppPermissionOutgoingWebhook},
		OutgoingURL: &url,
	}})
	if err != nil || out.App.OutgoingSecret == nil || *out.App.OutgoingSecret == "" {
		t.Fatalf("署名の秘密鍵が発行されていません: %v %+v", err, out)
	}
}

func TestPostChecksPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []entity.AppPermission
		channelID   string
		parentID    *string
		wantErr     error
	}{
		{name: "参加中のチャンネルに投稿できる", permissions: []entity.AppPermission{entity.AppPermissionPostJoinedChannels}, channelID: joinedID},
		{name: "権限がなければ参加中でも投稿できない", channelID: joinedID, wantErr: ErrForbiddenChannel},
		{name: "参加していない公開チャンネルには公開チャンネルの権限が要る", permissions: []entity.AppPermission{entity.AppPermissionPostJoinedChannels}, channelID: publicID, wantErr: ErrForbiddenChannel},
		{name: "公開チャンネルの権限で参加していない公開チャンネルに投稿できる", permissions: []entity.AppPermission{entity.AppPermissionPostPublicChannels}, channelID: publicID},
		{name: "非公開チャンネルは参加が要る", permissions: []entity.AppPermission{entity.AppPermissionPostPublicChannels}, channelID: privateID, wantErr: ErrForbiddenChannel},
		{name: "スレッドへの返信には権限が要る", permissions: []entity.AppPermission{entity.AppPermissionPostJoinedChannels}, channelID: joinedID, parentID: new(parentID), wantErr: ErrForbiddenThread},
		{name: "スレッドに返信できる", permissions: []entity.AppPermission{entity.AppPermissionPostJoinedChannels, entity.AppPermissionPostThreadReplies}, channelID: joinedID, parentID: new(parentID)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			out := f.create(t, tt.permissions...)
			_, err := f.post(out.Token, out, PostInput{Text: "hi", ChannelID: &tt.channelID, ParentID: tt.parentID})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (len(f.poster.posted) != 1 || f.apps.apps[out.App.ID].LastUsedAt == nil) {
				t.Fatal("投稿されていません")
			}
		})
	}
}

func TestPostUsesDefaultChannelAndRejectsWrongToken(t *testing.T) {
	f := newFixture()
	out := f.create(t, entity.AppPermissionPostJoinedChannels)

	if _, err := f.post("wrong", out, PostInput{Text: "hi"}); !errors.Is(err, ErrAppNotFound) {
		t.Fatalf("誤ったトークンを拒否していません: %v", err)
	}
	if _, err := f.post(out.Token, out, PostInput{Text: "hi"}); err != nil {
		t.Fatalf("既定のチャンネルに投稿できません: %v", err)
	}
	if got := f.poster.posted[0].ChannelID; got != joinedID {
		t.Fatalf("既定のチャンネルに投稿されていません: %s", got)
	}
}

func TestOfficialAppIsCreatedOnceAndCannotBeManaged(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	official, err := f.uc.EnsureOfficial(ctx, workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.uc.EnsureOfficial(ctx, workspaceID)
	if err != nil || again.ID != official.ID {
		t.Fatalf("公式アプリが 2 つ作られました: %v", err)
	}
	if bot := f.users.users[official.BotUserID]; bot == nil || !bot.IsOfficial {
		t.Fatalf("公式アプリのボットユーザーに印がありません: %+v", bot)
	}
	if err := f.uc.Delete(ctx, TargetInput{AppID: official.ID, UserID: adminID}); !errors.Is(err, ErrOfficialApp) {
		t.Fatalf("管理者でも公式アプリは削除できないはず: %v", err)
	}
	// 公式アプリは参加していないチャンネルにも投稿できる
	if _, err := f.uc.PostAsOfficial(ctx, workspaceID, privateID, nil, "reminder"); err != nil {
		t.Fatalf("公式アプリで投稿できません: %v", err)
	}
}

func TestDeleteRequiresCreatorOrAdmin(t *testing.T) {
	f := newFixture()
	out := f.create(t)
	ctx := context.Background()

	if err := f.uc.Delete(ctx, TargetInput{AppID: out.App.ID, UserID: otherID}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("作成者と管理者以外の削除を拒否していません: %v", err)
	}
}
