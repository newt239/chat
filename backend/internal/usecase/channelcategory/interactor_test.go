package channelcategory

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
	workspaceID = "general"
	aliceID     = "alice"
	bobID       = "bob"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return &entity.WorkspaceMember{UserID: userID}, nil
}

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: workspaceID}, nil
}

type fakeCategoryRepo struct {
	domainrepository.ChannelCategoryRepository
	categories []*entity.ChannelCategory
}

func (r *fakeCategoryRepo) FindByID(_ context.Context, id string) (*entity.ChannelCategory, error) {
	for _, c := range r.categories {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (r *fakeCategoryRepo) FindByUser(_ context.Context, userID string, _ string) ([]*entity.ChannelCategory, error) {
	var result []*entity.ChannelCategory
	for _, c := range r.categories {
		if c.UserID == userID {
			result = append(result, c)
		}
	}
	slices.SortFunc(result, func(a, b *entity.ChannelCategory) int { return a.Position - b.Position })
	return result, nil
}

func (r *fakeCategoryRepo) Create(_ context.Context, c *entity.ChannelCategory) error {
	c.ID = c.Name
	r.categories = append(r.categories, c)
	return nil
}

func (r *fakeCategoryRepo) UpdatePositions(_ context.Context, ids []string) error {
	for position, id := range ids {
		for _, c := range r.categories {
			if c.ID == id {
				c.Position = position
			}
		}
	}
	return nil
}

func (r *fakeCategoryRepo) SetChannel(_ context.Context, _ string, channelID string, categoryID *string) error {
	for _, c := range r.categories {
		c.ChannelIDs = slices.DeleteFunc(c.ChannelIDs, func(id string) bool { return id == channelID })
		if categoryID != nil && c.ID == *categoryID {
			c.ChannelIDs = append(c.ChannelIDs, channelID)
		}
	}
	return nil
}

type stubTxManager struct{}

func (stubTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newInteractor() (UseCase, *fakeCategoryRepo) {
	repo := &fakeCategoryRepo{}
	return NewInteractor(repo, stubWorkspaceRepo{}, stubAccess{}, stubTxManager{}), repo
}

func TestCreateAndReorder(t *testing.T) {
	uc, _ := newInteractor()
	ctx := context.Background()
	for _, name := range []string{"a", "b", "c"} {
		if _, err := uc.Create(ctx, CreateInput{WorkspaceID: workspaceID, UserID: aliceID, Name: " " + name + " "}); err != nil {
			t.Fatal(err)
		}
	}

	out, err := uc.Reorder(ctx, ReorderInput{WorkspaceID: workspaceID, UserID: aliceID, CategoryIDs: []string{"c", "a", "b"}})
	if err != nil {
		t.Fatalf("並び替えできません: %v", err)
	}
	if out[0].ID != "c" || out[0].Position != 0 || out[2].ID != "b" {
		t.Fatalf("並び順が反映されていません: %+v", out)
	}

	_, err = uc.Reorder(ctx, ReorderInput{WorkspaceID: workspaceID, UserID: aliceID, CategoryIDs: []string{"c"}})
	if !errors.Is(err, domerr.ErrValidation) {
		t.Fatalf("一部のカテゴリだけの並び替えが拒否されていません: %v", err)
	}
}

func TestSetChannelMovesBetweenCategories(t *testing.T) {
	uc, repo := newInteractor()
	ctx := context.Background()
	for _, name := range []string{"a", "b"} {
		if _, err := uc.Create(ctx, CreateInput{WorkspaceID: workspaceID, UserID: aliceID, Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	a, b := "a", "b"
	if err := uc.SetChannel(ctx, SetChannelInput{ChannelID: "ch", UserID: aliceID, CategoryID: &a}); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetChannel(ctx, SetChannelInput{ChannelID: "ch", UserID: aliceID, CategoryID: &b}); err != nil {
		t.Fatal(err)
	}
	if len(repo.categories[0].ChannelIDs) != 0 || len(repo.categories[1].ChannelIDs) != 1 {
		t.Fatalf("チャンネルが移動していません: %+v", repo.categories)
	}
	if err := uc.SetChannel(ctx, SetChannelInput{ChannelID: "ch", UserID: aliceID}); err != nil {
		t.Fatal(err)
	}
	if len(repo.categories[1].ChannelIDs) != 0 {
		t.Fatalf("既定のカテゴリに戻っていません: %+v", repo.categories)
	}
}

func TestOtherUsersCategoryIsHidden(t *testing.T) {
	uc, _ := newInteractor()
	ctx := context.Background()
	if _, err := uc.Create(ctx, CreateInput{WorkspaceID: workspaceID, UserID: aliceID, Name: "a"}); err != nil {
		t.Fatal(err)
	}
	a := "a"
	if err := uc.SetChannel(ctx, SetChannelInput{ChannelID: "ch", UserID: bobID, CategoryID: &a}); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("他人のカテゴリに割り当てられています: %v", err)
	}
	if err := uc.Delete(ctx, DeleteInput{CategoryID: "a", UserID: bobID}); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("他人のカテゴリを削除できています: %v", err)
	}
}
