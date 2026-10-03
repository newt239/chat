package user

import (
	"context"
	"errors"
	"testing"
	"time"

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

func (r *stubUserRepo) Delete(_ context.Context, id string) error {
	delete(r.users, id)
	return nil
}

type stubWorkspaceRepo struct {
	repository.WorkspaceRepository
	memberships []*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindMembershipsByUserID(context.Context, string) ([]*entity.WorkspaceMember, error) {
	return r.memberships, nil
}

type stubCloser struct{ closed []string }

func (c *stubCloser) CloseUser(userID string) { c.closed = append(c.closed, userID) }

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
	uc := New(repo, nil, nil, nil, nil)

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
	uc := New(&stubUserRepo{users: map[string]*entity.User{}}, nil, nil, nil, nil)

	_, err := uc.UpdatePreferences(context.Background(), "ghost", cobalt)
	if !errors.Is(err, domerr.ErrUserNotFound) {
		t.Fatalf("存在しないユーザーの設定更新が拒否されていません: %v", err)
	}
}

func TestUpdatePreferencesRequiresLogin(t *testing.T) {
	uc := New(&stubUserRepo{users: map[string]*entity.User{}}, nil, nil, nil, nil)

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
			uc := New(&stubUserRepo{users: map[string]*entity.User{"alice": {ID: "alice"}}}, nil, nil, nil, nil)
			prefs := cobalt
			prefs.Timezone = tt.timezone

			_, err := uc.UpdatePreferences(context.Background(), "alice", prefs)
			if errors.Is(err, domerr.ErrInvalidTimeZone) != tt.wantErr {
				t.Fatalf("タイムゾーン %q の検証結果が期待と異なります: %v", tt.timezone, err)
			}
		})
	}
}

func TestDeleteMe(t *testing.T) {
	tests := []struct {
		name    string
		role    entity.WorkspaceRole
		wantErr error
	}{
		{name: "メンバーは退会でき、接続も切る", role: entity.WorkspaceRoleMember},
		{name: "オーナーはワークスペースを残したまま退会できない", role: entity.WorkspaceRoleOwner, wantErr: ErrOwnerCannotDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubUserRepo{users: map[string]*entity.User{"alice": {ID: "alice"}}}
			closer := &stubCloser{}
			ws := &stubWorkspaceRepo{memberships: []*entity.WorkspaceMember{{WorkspaceID: "ws", UserID: "alice", Role: tt.role}}}
			err := New(repo, nil, ws, nil, closer).DeleteMe(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("期待したエラーと異なります: %v", err)
			}
			_, remains := repo.users["alice"]
			if remains != (tt.wantErr != nil) || (len(closer.closed) == 1) != (tt.wantErr == nil) {
				t.Fatalf("削除と接続の切断が期待と異なります: remains=%v closed=%v", remains, closer.closed)
			}
		})
	}
}

func TestGetMeRejectsDeletedUser(t *testing.T) {
	now := time.Now()
	uc := New(&stubUserRepo{users: map[string]*entity.User{"alice": {ID: "alice", DeletedAt: &now}}}, nil, nil, nil, nil)
	if _, err := uc.GetMe(context.Background(), "alice"); !errors.Is(err, domerr.ErrUserNotFound) {
		t.Fatalf("退会済みのユーザーを返しています: %v", err)
	}
}
