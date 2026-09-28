package user

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
)

type stubUserRepo struct {
	repository.UserRepository
	users map[string]*entity.User
}

func (r *stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	copied := *u
	return &copied, nil
}

func (r *stubUserRepo) Update(_ context.Context, u *entity.User) error {
	copied := *u
	r.users[u.ID] = &copied
	return nil
}

var cobalt = entity.UserPreferences{
	ThemeHue:     262,
	ThemeChroma:  0.17,
	ThemeSidebar: entity.SidebarStyleTinted,
	ColorMode:    entity.ColorModeDark,
	Locale:       "en",
}

func TestUpdatePreferencesSavesAndReturnsPreferences(t *testing.T) {
	repo := &stubUserRepo{users: map[string]*entity.User{"alice": {ID: "alice", DisplayName: "Alice"}}}
	uc := NewInteractor(repo, nil, nil)

	got, err := uc.UpdatePreferences(context.Background(), UpdatePreferencesInput{UserID: "alice", Preferences: cobalt})
	if err != nil {
		t.Fatalf("設定の保存に失敗しました: %v", err)
	}
	if *got != cobalt {
		t.Fatalf("返り値が保存した設定と一致しません: %+v", got)
	}

	me, err := uc.GetMe(context.Background(), "alice")
	if err != nil {
		t.Fatalf("プロフィールの取得に失敗しました: %v", err)
	}
	if me.Preferences != cobalt {
		t.Fatalf("保存した設定がプロフィールに反映されていません: %+v", me.Preferences)
	}
}

func TestUpdatePreferencesRejectsUnknownUser(t *testing.T) {
	uc := NewInteractor(&stubUserRepo{users: map[string]*entity.User{}}, nil, nil)

	_, err := uc.UpdatePreferences(context.Background(), UpdatePreferencesInput{UserID: "ghost", Preferences: cobalt})
	if !errors.Is(err, entity.ErrUserNotFound) {
		t.Fatalf("存在しないユーザーの設定更新が拒否されていません: %v", err)
	}
}

func TestUpdatePreferencesRequiresLogin(t *testing.T) {
	uc := NewInteractor(&stubUserRepo{users: map[string]*entity.User{}}, nil, nil)

	_, err := uc.UpdatePreferences(context.Background(), UpdatePreferencesInput{Preferences: cobalt})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("未ログインでの設定更新が拒否されていません: %v", err)
	}
}
