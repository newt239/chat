package user

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
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
	ThemeHue:          262,
	ThemeChroma:       0.17,
	ThemeSidebar:      entity.SidebarStyleTinted,
	ColorMode:         entity.ColorModeDark,
	Locale:            "en",
	NotificationLevel: entity.NotificationLevelAll,
}

func TestUpdatePreferencesSavesAndReturnsPreferences(t *testing.T) {
	repo := &stubUserRepo{users: map[string]*entity.User{"alice": {ID: "alice", DisplayName: "Alice"}}}
	uc := New(repo, nil, nil, nil)

	got, err := uc.UpdatePreferences(context.Background(), "alice", cobalt)
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
	uc := New(&stubUserRepo{users: map[string]*entity.User{}}, nil, nil, nil)

	_, err := uc.UpdatePreferences(context.Background(), "ghost", cobalt)
	if !errors.Is(err, domerr.ErrUserNotFound) {
		t.Fatalf("存在しないユーザーの設定更新が拒否されていません: %v", err)
	}
}

func TestUpdatePreferencesRequiresLogin(t *testing.T) {
	uc := New(&stubUserRepo{users: map[string]*entity.User{}}, nil, nil, nil)

	_, err := uc.UpdatePreferences(context.Background(), "", cobalt)
	if !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("未ログインでの設定更新が拒否されていません: %v", err)
	}
}

func TestUpdatePreferencesValidatesTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{name: "IANA 名は保存する", timezone: "Asia/Tokyo"},
		{name: "空は未設定として保存する", timezone: ""},
		{name: "存在しない名前は拒否する", timezone: "Mars/Olympus", wantErr: true},
		{name: "サーバーのタイムゾーンを指す Local は拒否する", timezone: "Local", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := New(&stubUserRepo{users: map[string]*entity.User{"alice": {ID: "alice"}}}, nil, nil, nil)
			prefs := cobalt
			prefs.Timezone = tt.timezone

			_, err := uc.UpdatePreferences(context.Background(), "alice", prefs)
			if errors.Is(err, domerr.ErrInvalidTimeZone) != tt.wantErr {
				t.Fatalf("タイムゾーン %q の検証結果が期待と異なります: %v", tt.timezone, err)
			}
		})
	}
}
