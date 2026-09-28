package channel

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

const (
	workspaceID = "general"
	adminID     = "11111111-1111-1111-1111-111111111111"
	memberID    = "22222222-2222-2222-2222-222222222222"
	guestID     = "33333333-3333-3333-3333-333333333333"
)

type fakeChannelRepo struct {
	domainrepository.ChannelRepository
	channels map[string]*entity.Channel
	members  *fakeMemberRepo
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

type stubTxManager struct{}

func (stubTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fixture struct {
	uc       ChannelUseCase
	channels *fakeChannelRepo
	members  *fakeMemberRepo
	stars    *fakeStarRepo
}

func newFixture() *fixture {
	members := &fakeMemberRepo{joined: map[string]map[string]bool{}}
	channels := &fakeChannelRepo{channels: map[string]*entity.Channel{}, members: members}
	stars := &fakeStarRepo{starred: map[string]bool{}}
	workspaces := stubWorkspaceRepo{}
	access := domainservice.NewChannelAccessService(channels, members, workspaces)
	uc := NewChannelInteractor(channels, members, stars, workspaces, stubReadStateRepo{}, stubTxManager{}, nil, access)
	return &fixture{uc: uc, channels: channels, members: members, stars: stars}
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

	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("閲覧できない非公開チャンネルの下に作成できています: %v", err)
	}
}

func TestCreateChannelPermission(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	if _, err := f.uc.CreateChannel(ctx, CreateChannelInput{WorkspaceID: workspaceID, UserID: memberID, Name: "public"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("一般メンバーが公開チャンネルを作成できています: %v", err)
	}
	if _, err := f.uc.CreateChannel(ctx, CreateChannelInput{WorkspaceID: workspaceID, UserID: guestID, Name: "private", IsPrivate: true}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("ゲストが非公開チャンネルを作成できています: %v", err)
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
