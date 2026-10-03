package bookmark

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type stubBookmarkRepo struct {
	repository.BookmarkRepository
	bookmarks []*entity.MessageBookmark
}

func (r stubBookmarkRepo) FindByUserID(context.Context, string, string) ([]*entity.MessageBookmark, error) {
	return r.bookmarks, nil
}

type stubAccess struct {
	service.ChannelAccessService
	accessible map[string]*entity.Channel
}

func (s stubAccess) AccessibleChannelsByIDs(context.Context, []string, string) (map[string]*entity.Channel, error) {
	return s.accessible, nil
}

func TestListBookmarksExcludesInaccessibleChannels(t *testing.T) {
	repo := stubBookmarkRepo{bookmarks: []*entity.MessageBookmark{{MessageID: "m1", Message: &entity.Message{ID: "m1", ChannelID: "left"}}}}
	uc := New(repo, nil, stubAccess{accessible: map[string]*entity.Channel{}})

	got, err := uc.ListBookmarks(context.Background(), "alice", "ws")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("閲覧できなくなったチャンネルのブックマークが含まれています: %+v", got)
	}
}
