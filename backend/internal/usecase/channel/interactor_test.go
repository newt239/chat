package channel

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/usecase/systemmessage"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
)

const (
	workspaceID = "general"
	adminID     = "11111111-1111-1111-1111-111111111111"
	memberID    = "22222222-2222-2222-2222-222222222222"
	guestID     = "33333333-3333-3333-3333-333333333333"
)

type fakeChannelRepo struct {
	domainrepository.ChannelRepository
	channels      map[string]*entity.Channel
	members       *fakeMemberRepo
	lastMessageAt map[string]time.Time
	lastFilter    domainrepository.BrowsableChannelFilter
}

func (r *fakeChannelRepo) SearchBrowsableChannels(ctx context.Context, workspaceID, userID string, filter domainrepository.BrowsableChannelFilter) ([]*entity.Channel, int, error) {
	r.lastFilter = filter
	channels, err := r.FindBrowsableChannels(ctx, workspaceID, userID)
	return channels, len(channels), err
}

func (r *fakeChannelRepo) FindLastMessageAtBatch(_ context.Context, _ []string) (map[string]time.Time, error) {
	return r.lastMessageAt, nil
}

func (r *fakeChannelRepo) FindBrowsableChannels(ctx context.Context, _ string, userID string) ([]*entity.Channel, error) {
	var result []*entity.Channel
	for _, ch := range r.channels {
		if !ch.IsPrivate() || r.members.joined[ch.ID][userID] {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (r *fakeChannelRepo) CountMembersBatch(_ context.Context, ids []string) (map[string]int, error) {
	result := map[string]int{}
	for _, id := range ids {
		result[id] = len(r.members.joined[id])
	}
	return result, nil
}

func (r *fakeChannelRepo) FindByID(_ context.Context, id string) (*entity.Channel, error) {
	return r.channels[id], nil
}

func (r *fakeChannelRepo) FindAccessibleChannels(_ context.Context, _ string, userID string) ([]*entity.Channel, error) {
	var result []*entity.Channel
	for _, ch := range r.channels {
		if r.members.joined[ch.ID][userID] {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (r *fakeChannelRepo) FindByNames(_ context.Context, _ string, names []string) ([]*entity.Channel, error) {
	var result []*entity.Channel
	for _, ch := range r.channels {
		if slices.Contains(names, ch.Name) {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (r *fakeChannelRepo) FindDescendants(_ context.Context, parent *entity.Channel) ([]*entity.Channel, error) {
	var result []*entity.Channel
	for _, ch := range r.channels {
		if strings.HasPrefix(ch.Name, parent.Name+"/") {
			result = append(result, ch)
		}
	}
	return result, nil
}

func (r *fakeChannelRepo) Create(_ context.Context, ch *entity.Channel) error {
	r.channels[ch.ID] = ch
	return nil
}

func (r *fakeChannelRepo) Update(_ context.Context, ch *entity.Channel) error {
	r.channels[ch.ID] = ch
	return nil
}

func (r *fakeChannelRepo) Delete(_ context.Context, id string) error {
	delete(r.channels, id)
	return nil
}

func (r *fakeChannelRepo) byName(name string) *entity.Channel {
	for _, ch := range r.channels {
		if ch.Name == name {
			return ch
		}
	}
	return nil
}

type fakeMemberRepo struct {
	domainrepository.ChannelMemberRepository
	joined map[string]map[string]bool
}

func (r *fakeMemberRepo) AddMember(_ context.Context, m *entity.ChannelMember) error {
	if r.joined[m.ChannelID] == nil {
		r.joined[m.ChannelID] = map[string]bool{}
	}
	r.joined[m.ChannelID][m.UserID] = true
	return nil
}

func (r *fakeMemberRepo) IsMember(_ context.Context, channelID, userID string) (bool, error) {
	return r.joined[channelID][userID], nil
}

func (r *fakeMemberRepo) FindJoinedChannelIDs(_ context.Context, userID string, channelIDs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range channelIDs {
		if r.joined[id][userID] {
			result[id] = true
		}
	}
	return result, nil
}

type fakeStarRepo struct {
	domainrepository.ChannelStarRepository
	starred map[string]bool
}

func (r *fakeStarRepo) SetStarred(_ context.Context, _ string, channelID string, starred bool) error {
	r.starred[channelID] = starred
	return nil
}

func (r *fakeStarRepo) FindStarredChannelIDs(_ context.Context, _ string, channelIDs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range channelIDs {
		if r.starred[id] {
			result[id] = true
		}
	}
	return result, nil
}

type fakeMuteRepo struct {
	domainrepository.ChannelMuteRepository
	muted map[string]bool
}

func (r *fakeMuteRepo) SetMuted(_ context.Context, _ string, channelID string, muted bool) error {
	r.muted[channelID] = muted
	return nil
}

func (r *fakeMuteRepo) FindMutedChannelIDs(_ context.Context, _ string, channelIDs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range channelIDs {
		if r.muted[id] {
			result[id] = true
		}
	}
	return result, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	return &entity.Workspace{ID: id}, nil
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	roles := map[string]entity.WorkspaceRole{adminID: entity.WorkspaceRoleAdmin, memberID: entity.WorkspaceRoleMember, guestID: entity.WorkspaceRoleGuest}
	role, ok := roles[userID]
	if !ok {
		return nil, nil
	}
	return &entity.WorkspaceMember{UserID: userID, Role: role}, nil
}

type stubReadStateRepo struct {
	domainrepository.ReadStateRepository
}

func (stubReadStateRepo) GetUnreadCountBatch(_ context.Context, _ []string, _ string) (map[string]int, error) {
	return map[string]int{}, nil
}

func (stubReadStateRepo) GetUnreadMentionCountBatch(_ context.Context, _ []string, _ string) (map[string]int, error) {
	return map[string]int{}, nil
}

type stubPermissionRepo struct {
	domainrepository.PermissionRepository
	overrides []entity.PermissionOverride
}

func (r *stubPermissionRepo) FindOverrides(context.Context, string) ([]entity.PermissionOverride, error) {
	return r.overrides, nil
}

type stubTxManager struct{}

func (stubTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fixture struct {
	uc          ChannelUseCase
	channels    *fakeChannelRepo
	members     *fakeMemberRepo
	stars       *fakeStarRepo
	mutes       *fakeMuteRepo
	permissions *stubPermissionRepo
	recorder    *audittest.Recorder
	revoker     *stubRevoker
}

type stubSystemMessages struct{}

func (stubSystemMessages) Create(context.Context, systemmessage.CreateInput) (*entity.SystemMessage, error) {
	return &entity.SystemMessage{}, nil
}

type stubRevoker struct{ revoked []string }

func (r *stubRevoker) RevokeChannel(_, channelID, userID string) {
	r.revoked = append(r.revoked, channelID+"/"+userID)
}

type nopLogger struct {
	domainservice.Logger
}

func (nopLogger) Warn(string, ...domainservice.LogField) {}

func newFixture() *fixture {
	members := &fakeMemberRepo{joined: map[string]map[string]bool{}}
	channels := &fakeChannelRepo{channels: map[string]*entity.Channel{}, members: members, lastMessageAt: map[string]time.Time{}}
	stars := &fakeStarRepo{starred: map[string]bool{}}
	mutes := &fakeMuteRepo{muted: map[string]bool{}}
	workspaces := stubWorkspaceRepo{}
	access := domainservice.NewChannelAccessService(channels, members, workspaces)
	permissions := &stubPermissionRepo{}
	recorder := &audittest.Recorder{}
	permissionSvc := domainservice.NewPermissionService(workspaces, permissions)
	revoker := &stubRevoker{}
	uc := NewChannelInteractor(channels, members, stars, mutes, workspaces, stubReadStateRepo{}, stubTxManager{}, stubSystemMessages{}, access, permissionSvc, recorder, revoker, nopLogger{})
	return &fixture{uc: uc, channels: channels, members: members, stars: stars, mutes: mutes, permissions: permissions, recorder: recorder, revoker: revoker}
}

func (f *fixture) create(t *testing.T, userID, name string, isPrivate bool) *ChannelOutput {
	t.Helper()
	out, err := f.uc.CreateChannel(context.Background(), CreateChannelInput{WorkspaceID: workspaceID, UserID: userID, Name: name, IsPrivate: isPrivate})
	if err != nil {
		t.Fatalf("チャンネル %s を作成できません: %v", name, err)
	}
	return out
}

func TestCreateChannelCreatesMissingAncestors(t *testing.T) {
	f := newFixture()

	out := f.create(t, adminID, "Dev/Frontend/Web", false)

	if out.Name != "dev/frontend/web" {
		t.Fatalf("名前が正規化されていません: %s", out.Name)
	}
	dev := f.channels.byName("dev")
	frontend := f.channels.byName("dev/frontend")
	if dev == nil || frontend == nil {
		t.Fatal("存在しない親チャンネルが作成されていません")
	}
	if dev.ParentID != nil || *frontend.ParentID != dev.ID || *out.ParentID != frontend.ID {
		t.Fatal("親子関係が正しく設定されていません")
	}
	if !f.members.joined[dev.ID][adminID] || !f.members.joined[out.ID][adminID] {
		t.Fatal("作成者が作成したチャンネルに参加していません")
	}
}

func TestCreateChannelReusesExistingParent(t *testing.T) {
	f := newFixture()
	dev := f.create(t, adminID, "dev", false)

	out := f.create(t, adminID, "dev/backend", false)

	if *out.ParentID != dev.ID || len(f.channels.channels) != 2 {
		t.Fatal("既存の親チャンネルが使われていません")
	}
}

func TestCreateChannelRejectsDuplicateAndInvalidName(t *testing.T) {
	f := newFixture()
	f.create(t, adminID, "dev", false)

	_, err := f.uc.CreateChannel(context.Background(), CreateChannelInput{WorkspaceID: workspaceID, UserID: adminID, Name: "DEV"})
	if !errors.Is(err, ErrChannelNameExists) {
		t.Fatalf("重複した名前が拒否されていません: %v", err)
	}

	_, err = f.uc.CreateChannel(context.Background(), CreateChannelInput{WorkspaceID: workspaceID, UserID: adminID, Name: "a/b/c/d/e"})
	if !errors.Is(err, domerr.ErrValidation) {
		t.Fatalf("深すぎる階層が拒否されていません: %v", err)
	}
}

func TestCreateChannelUnderInaccessiblePrivateParent(t *testing.T) {
	f := newFixture()
	f.create(t, adminID, "secret", true)

	_, err := f.uc.CreateChannel(context.Background(), CreateChannelInput{WorkspaceID: workspaceID, UserID: memberID, Name: "secret/child", IsPrivate: true})

	if !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("閲覧できない非公開チャンネルの下に作成できています: %v", err)
	}
}

func TestCreateChannelPermission(t *testing.T) {
	denyPrivate := entity.PermissionOverride{Role: entity.WorkspaceRoleMember, Permission: entity.PermissionCreatePrivateChannel, Allowed: false}
	tests := []struct {
		name      string
		userID    string
		isPrivate bool
		overrides []entity.PermissionOverride
		wantErr   error
	}{
		{name: "既定ではメンバーは公開チャンネルを作れる", userID: memberID},
		{name: "既定ではゲストは非公開チャンネルを作れない", userID: guestID, isPrivate: true, wantErr: domerr.ErrUnauthorized},
		{name: "非公開チャンネルの作成を禁止するとメンバーは作れない", userID: memberID, isPrivate: true, overrides: []entity.PermissionOverride{denyPrivate}, wantErr: domerr.ErrUnauthorized},
		{name: "非公開チャンネルを禁止しても公開チャンネルは作れる", userID: memberID, overrides: []entity.PermissionOverride{denyPrivate}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			f.permissions.overrides = tt.overrides
			_, err := f.uc.CreateChannel(context.Background(), CreateChannelInput{WorkspaceID: workspaceID, UserID: tt.userID, Name: "room", IsPrivate: tt.isPrivate})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			var wantActions []entity.AuditAction
			if tt.wantErr == nil {
				wantActions = []entity.AuditAction{entity.AuditActionChannelCreated}
			}
			if !slices.Equal(f.recorder.Actions(), wantActions) {
				t.Errorf("監査ログが期待と異なります: %v", f.recorder.Actions())
			}
		})
	}
}

func TestSetArchived(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		dm      bool
		wantErr error
	}{
		{name: "作成者はアーカイブできる", userID: memberID},
		{name: "管理者はアーカイブできる", userID: adminID},
		{name: "作成者以外のメンバーはアーカイブできない", userID: guestID, wantErr: domerr.ErrUnauthorized},
		{name: "DM はアーカイブできない", userID: memberID, dm: true, wantErr: ErrCannotModifyDM},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			ch := f.create(t, memberID, "room", false)
			if tt.dm {
				f.channels.channels[ch.ID].Type = entity.ChannelTypeDM
			}
			f.recorder.Logs = nil

			out, err := f.uc.SetArchived(context.Background(), SetArchivedInput{ChannelID: ch.ID, UserID: tt.userID, Archived: true})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if len(f.recorder.Logs) != 0 {
					t.Errorf("失敗時に監査ログが記録されました")
				}
				return
			}
			if out.ArchivedAt == nil || !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionChannelArchived}) {
				t.Errorf("アーカイブされていないか記録がありません: %+v %v", out, f.recorder.Actions())
			}

			// 同じ状態への変更は記録しない
			if _, err := f.uc.SetArchived(context.Background(), SetArchivedInput{ChannelID: ch.ID, UserID: tt.userID, Archived: true}); err != nil || len(f.recorder.Logs) != 1 {
				t.Errorf("二重にアーカイブした場合は何もしないはず: %v %v", err, f.recorder.Actions())
			}
		})
	}
}

func TestCreatePrivateChannelWithMembers(t *testing.T) {
	f := newFixture()

	out, err := f.uc.CreateChannel(context.Background(), CreateChannelInput{
		WorkspaceID: workspaceID, UserID: memberID, Name: "project", IsPrivate: true, MemberIDs: []string{adminID, memberID},
	})
	if err != nil {
		t.Fatalf("メンバー指定の非公開チャンネルを作成できません: %v", err)
	}
	if !f.members.joined[out.ID][adminID] || !f.members.joined[out.ID][memberID] {
		t.Fatal("指定したメンバーが参加していません")
	}

	_, err = f.uc.CreateChannel(context.Background(), CreateChannelInput{
		WorkspaceID: workspaceID, UserID: memberID, Name: "project2", IsPrivate: true, MemberIDs: []string{uuid.NewString()},
	})
	if !errors.Is(err, ErrMemberNotInWorkspace) {
		t.Fatalf("ワークスペース外のユーザーを追加できています: %v", err)
	}
}

func TestListChannelsIncludesAncestorsAndStars(t *testing.T) {
	f := newFixture()
	web := f.create(t, adminID, "dev/frontend/web", false)
	dev := f.channels.byName("dev")
	// 一般メンバーは子チャンネルにだけ参加している
	f.members.joined[web.ID][memberID] = true
	f.stars.starred[dev.ID] = true
	f.channels.lastMessageAt[web.ID] = time.Unix(100, 0)

	out, err := f.uc.ListChannels(context.Background(), ListChannelsInput{WorkspaceID: workspaceID, UserID: memberID})
	if err != nil {
		t.Fatalf("一覧を取得できません: %v", err)
	}

	names := make([]string, 0, len(out))
	for _, ch := range out {
		names = append(names, ch.Name)
	}
	if !slices.Equal(names, []string{"dev", "dev/frontend", "dev/frontend/web"}) {
		t.Fatalf("祖先を含めて名前順に返していません: %v", names)
	}
	if out[0].IsMember || !out[0].IsStarred || !out[2].IsMember {
		t.Fatalf("参加状態またはスターが正しくありません: %+v", out)
	}
	if out[0].LastMessageAt != nil || out[2].LastMessageAt == nil || !out[2].LastMessageAt.Equal(time.Unix(100, 0)) {
		t.Fatalf("最後のメッセージの日時が正しくありません: %+v", out)
	}
}

func TestListBrowsableChannels(t *testing.T) {
	f := newFixture()
	public := f.create(t, adminID, "public", false)
	f.create(t, adminID, "secret", true)

	out, err := f.uc.ListBrowsableChannels(context.Background(), ListChannelsInput{WorkspaceID: workspaceID, UserID: memberID})
	if err != nil {
		t.Fatalf("一覧を取得できません: %v", err)
	}
	if len(out) != 1 || out[0].Channel.ID != public.ID || out[0].Channel.IsMember || out[0].MemberCount != 1 {
		t.Fatalf("未参加の公開チャンネルだけが返っていません: %+v", out)
	}
}

func TestSearchBrowsableChannels(t *testing.T) {
	f := newFixture()
	public := f.create(t, adminID, "public", false)

	out, err := f.uc.SearchBrowsableChannels(context.Background(), SearchBrowsableChannelsInput{
		WorkspaceID: workspaceID, UserID: memberID, Query: "pub", Sort: domainrepository.BrowsableChannelSortMemberCount, Page: 3, PerPage: 20,
	})
	if err != nil {
		t.Fatalf("検索できません: %v", err)
	}
	if out.Total != 1 || len(out.Channels) != 1 || out.Channels[0].Channel.ID != public.ID || out.Channels[0].MemberCount != 1 {
		t.Fatalf("検索結果が正しくありません: %+v", out)
	}
	want := domainrepository.BrowsableChannelFilter{Query: "pub", Sort: domainrepository.BrowsableChannelSortMemberCount, Limit: 20, Offset: 40}
	if f.channels.lastFilter != want {
		t.Fatalf("ページから求めた条件が正しくありません: %+v", f.channels.lastFilter)
	}

	if _, err := f.uc.SearchBrowsableChannels(context.Background(), SearchBrowsableChannelsInput{WorkspaceID: workspaceID, UserID: "44444444-4444-4444-4444-444444444444", Page: 1, PerPage: 20}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("ワークスペース外のユーザーが検索できています: %v", err)
	}
}

func TestUpdateChannelRenamesDescendants(t *testing.T) {
	f := newFixture()
	f.create(t, adminID, "dev/frontend/web", false)
	dev := f.channels.byName("dev")

	name := "engineering"
	if _, err := f.uc.UpdateChannel(context.Background(), UpdateChannelInput{ChannelID: dev.ID, UserID: adminID, Name: &name}); err != nil {
		t.Fatalf("名前を変更できません: %v", err)
	}

	if f.channels.byName("engineering/frontend") == nil || f.channels.byName("engineering/frontend/web") == nil {
		t.Fatal("子孫チャンネルのパスが付け替えられていません")
	}
}

func TestDeleteChannelWithChildren(t *testing.T) {
	f := newFixture()
	f.create(t, adminID, "dev/frontend", false)
	dev := f.channels.byName("dev")

	err := f.uc.DeleteChannel(context.Background(), DeleteChannelInput{ChannelID: dev.ID, UserID: adminID})

	if !errors.Is(err, ErrChannelHasChildren) {
		t.Fatalf("子を持つチャンネルを削除できています: %v", err)
	}
}

func TestSetChannelStarredRequiresAccess(t *testing.T) {
	f := newFixture()
	secret := f.create(t, adminID, "secret", true)
	ctx := context.Background()

	if err := f.uc.SetChannelStarred(ctx, SetChannelStarredInput{ChannelID: secret.ID, UserID: memberID, Starred: true}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("閲覧できないチャンネルにスターを付けられています: %v", err)
	}
	if err := f.uc.SetChannelStarred(ctx, SetChannelStarredInput{ChannelID: secret.ID, UserID: adminID, Starred: true}); err != nil || !f.stars.starred[secret.ID] {
		t.Fatalf("スターを付けられません: %v", err)
	}
}

func TestSetChannelMutedRequiresAccess(t *testing.T) {
	f := newFixture()
	secret := f.create(t, adminID, "secret", true)
	ctx := context.Background()

	if err := f.uc.SetChannelMuted(ctx, SetChannelMutedInput{ChannelID: secret.ID, UserID: memberID, Muted: true}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("閲覧できないチャンネルをミュートできています: %v", err)
	}
	if err := f.uc.SetChannelMuted(ctx, SetChannelMutedInput{ChannelID: secret.ID, UserID: adminID, Muted: true}); err != nil || !f.mutes.muted[secret.ID] {
		t.Fatalf("ミュートできません: %v", err)
	}
}

func TestListAndGetChannelIncludeMuted(t *testing.T) {
	f := newFixture()
	general := f.create(t, adminID, "general", false)
	random := f.create(t, adminID, "random", false)
	ctx := context.Background()
	if err := f.uc.SetChannelMuted(ctx, SetChannelMutedInput{ChannelID: random.ID, UserID: adminID, Muted: true}); err != nil {
		t.Fatalf("ミュートできません: %v", err)
	}

	out, err := f.uc.ListChannels(ctx, ListChannelsInput{WorkspaceID: workspaceID, UserID: adminID})
	if err != nil {
		t.Fatalf("一覧を取得できません: %v", err)
	}
	muted := map[string]bool{}
	for _, ch := range out {
		muted[ch.ID] = ch.IsMuted
	}
	if muted[general.ID] || !muted[random.ID] {
		t.Fatalf("ミュート状態が正しくありません: %+v", out)
	}

	got, err := f.uc.GetChannel(ctx, GetChannelInput{ChannelID: random.ID, UserID: adminID})
	if err != nil || !got.IsMuted {
		t.Fatalf("GetChannel がミュート状態を返していません: %+v %v", got, err)
	}
}

func TestUpdateChannelPermissionAndPrivacy(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		dm          bool
		wantErr     error
		wantRevoked bool
	}{
		{name: "作成者は非公開にでき、参加していない接続の配信を止める", userID: memberID, wantRevoked: true},
		{name: "管理者は非公開にできる", userID: adminID, wantRevoked: true},
		{name: "作成者以外のメンバーは変更できない", userID: guestID, wantErr: domerr.ErrUnauthorized},
		{name: "DM は変更できない", userID: memberID, dm: true, wantErr: ErrCannotModifyDM},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			ch := f.create(t, memberID, "room", false)
			if tt.dm {
				f.channels.channels[ch.ID].Type = entity.ChannelTypeDM
			}

			_, err := f.uc.UpdateChannel(context.Background(), UpdateChannelInput{ChannelID: ch.ID, UserID: tt.userID, IsPrivate: new(true)})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if got := len(f.revoker.revoked) == 1; got != tt.wantRevoked {
				t.Errorf("配信の停止が期待と異なります: %v", f.revoker.revoked)
			}
			if tt.wantErr == nil && f.channels.channels[ch.ID].Type != entity.ChannelTypePrivate {
				t.Errorf("非公開になっていません: %s", f.channels.channels[ch.ID].Type)
			}
		})
	}
}

func TestDeleteChannelRevokesSubscriptions(t *testing.T) {
	f := newFixture()
	ch := f.create(t, adminID, "room", false)

	if err := f.uc.DeleteChannel(context.Background(), DeleteChannelInput{ChannelID: ch.ID, UserID: adminID}); err != nil {
		t.Fatalf("削除できません: %v", err)
	}
	if !slices.Equal(f.revoker.revoked, []string{ch.ID + "/"}) {
		t.Errorf("削除したチャンネルの配信を止めていません: %v", f.revoker.revoked)
	}
}
