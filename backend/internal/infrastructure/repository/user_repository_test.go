package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestUserRepositoryPreferencesAndLinks(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewUserRepository(client)
	ctx := context.Background()

	u, err := repo.FindByID(ctx, f.alice.ID.String())
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if u.Preferences != entity.DefaultPreferences() || len(u.Links) != 0 {
		t.Fatalf("設定を保存していなければ既定値を返すことを期待しましたが %+v %v でした", u.Preferences, u.Links)
	}

	for _, links := range [][]string{{"https://a.example", "https://b.example"}, {"https://c.example"}} {
		u.Preferences.Locale = "en"
		u.Links = links
		if err := repo.Update(ctx, u); err != nil {
			t.Fatalf("更新に失敗しました: %v", err)
		}
		got, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("取得に失敗しました: %v", err)
		}
		if got.Preferences.Locale != "en" || !reflect.DeepEqual(got.Links, links) {
			t.Errorf("設定とリンクが保存されていません: %+v %v", got.Preferences, got.Links)
		}
	}
}
