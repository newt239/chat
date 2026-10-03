package channellink

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

const (
	channelID = "channel"
	memberID  = "member"
	adminID   = "admin"
	guestID   = "guest"
)

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "general"}, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	roles := map[string]entity.WorkspaceRole{memberID: entity.WorkspaceRoleMember, adminID: entity.WorkspaceRoleAdmin, guestID: entity.WorkspaceRoleGuest}
	return &entity.WorkspaceMember{UserID: userID, Role: roles[userID]}, nil
}

type stubPermissionRepo struct {
	domainrepository.PermissionRepository
}

func (stubPermissionRepo) FindOverrides(context.Context, string) ([]entity.PermissionOverride, error) {
	return nil, nil
}

type fakeLinkRepo struct {
	domainrepository.ChannelLinkRepository
	links []*entity.ChannelLink
}

func (r *fakeLinkRepo) FindByID(_ context.Context, id string) (*entity.ChannelLink, error) {
	for _, l := range r.links {
		if l.ID == id {
			return l, nil
		}
	}
	return nil, nil
}

func (r *fakeLinkRepo) FindByChannelID(_ context.Context, _ string) ([]*entity.ChannelLink, error) {
	sorted := slices.Clone(r.links)
	slices.SortFunc(sorted, func(a, b *entity.ChannelLink) int { return a.Position - b.Position })
	return sorted, nil
}

func (r *fakeLinkRepo) Create(_ context.Context, link *entity.ChannelLink) error {
	link.ID = link.Title
	r.links = append(r.links, link)
	return nil
}

func (r *fakeLinkRepo) UpdatePositions(_ context.Context, ids []string) error {
	for position, id := range ids {
		for _, l := range r.links {
			if l.ID == id {
				l.Position = position
			}
		}
	}
	return nil
}

type stubTxManager struct{}

func (stubTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newInteractor() (*Interactor, *fakeLinkRepo) {
	repo := &fakeLinkRepo{}
	permissionSvc := domainservice.NewPermissionService(stubWorkspaceRepo{}, stubPermissionRepo{})
	return New(repo, stubAccess{}, permissionSvc, stubTxManager{}), repo
}

func TestCreateLinkPermission(t *testing.T) {
	uc, repo := newInteractor()
	ctx := context.Background()

	for _, userID := range []string{memberID, adminID} {
		if _, err := uc.Create(ctx, LinkInput{ID: channelID, UserID: userID, Title: userID, URL: "https://example.com"}); err != nil {
			t.Fatalf("%s がリンクを追加できません: %v", userID, err)
		}
	}
	if _, err := uc.Create(ctx, LinkInput{ID: channelID, UserID: guestID, Title: "x", URL: "https://example.com"}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("既定ではゲストがリンクを追加できないはず: %v", err)
	}
	if repo.links[1].Position != 1 {
		t.Fatalf("新しいリンクが末尾に追加されていません: %d", repo.links[1].Position)
	}

	links, canEdit, err := uc.List(ctx, channelID, guestID)
	if err != nil || canEdit || len(links) != 2 {
		t.Fatalf("閲覧のみのユーザーへの一覧が正しくありません: %+v err=%v", links, err)
	}
}

func TestReorderLinks(t *testing.T) {
	uc, repo := newInteractor()
	ctx := context.Background()
	for _, title := range []string{"a", "b", "c"} {
		if _, err := uc.Create(ctx, LinkInput{ID: channelID, UserID: memberID, Title: title, URL: "https://example.com"}); err != nil {
			t.Fatal(err)
		}
	}

	if err := uc.Reorder(ctx, channelID, memberID, []string{"c", "a", "b"}); err != nil {
		t.Fatalf("並び替えできません: %v", err)
	}
	out, _ := repo.FindByChannelID(ctx, channelID)
	if out[0].ID != "c" || out[0].Position != 0 || out[2].ID != "b" {
		t.Fatalf("並び順が反映されていません: %+v", out)
	}

	err := uc.Reorder(ctx, channelID, memberID, []string{"c", "a"})
	if !errors.Is(err, domerr.ErrValidation) {
		t.Fatalf("一部のリンクだけの並び替えが拒否されていません: %v", err)
	}
}
