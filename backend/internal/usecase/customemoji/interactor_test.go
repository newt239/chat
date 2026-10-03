package customemoji

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
)

const (
	workspaceID = "ws"
	creatorID   = "creator"
	otherID     = "other"
	adminID     = "admin"
	outsiderID  = "outsider"
	guestID     = "guest"
	uploadID    = "11111111-1111-1111-1111-111111111111"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	switch userID {
	case outsiderID:
		return nil, nil
	case adminID:
		return &entity.WorkspaceMember{UserID: userID, Role: entity.WorkspaceRoleAdmin}, nil
	case guestID:
		return &entity.WorkspaceMember{UserID: userID, Role: entity.WorkspaceRoleGuest}, nil
	}
	return &entity.WorkspaceMember{UserID: userID, Role: entity.WorkspaceRoleMember}, nil
}

type stubPermission struct {
	domainservice.PermissionService
}

func (stubPermission) Ensure(ctx context.Context, wsID, userID string, p entity.Permission) (*entity.WorkspaceMember, error) {
	member, _ := stubWorkspaceRepo{}.FindMember(ctx, wsID, userID)
	if member == nil || !entity.DefaultPermissionMatrix().Allows(member.Role, p) {
		return nil, domerr.ErrUnauthorized
	}
	return member, nil
}

type fakeEmojiRepo struct {
	domainrepository.CustomEmojiRepository
	emojis map[string]*entity.CustomEmoji
}

func (r *fakeEmojiRepo) FindByID(_ context.Context, id string) (*entity.CustomEmoji, error) {
	return r.emojis[id], nil
}

func (r *fakeEmojiRepo) FindByWorkspaceID(_ context.Context, wsID string) ([]*entity.CustomEmoji, error) {
	var result []*entity.CustomEmoji
	for _, e := range r.emojis {
		if e.WorkspaceID == wsID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *fakeEmojiRepo) Create(_ context.Context, e *entity.CustomEmoji) error {
	for _, existing := range r.emojis {
		if existing.WorkspaceID == e.WorkspaceID && existing.Name == e.Name {
			return domerr.ErrCustomEmojiNameExists
		}
	}
	r.emojis[e.ID] = e
	return nil
}

func (r *fakeEmojiRepo) Delete(_ context.Context, id string) error {
	delete(r.emojis, id)
	return nil
}

type fakeUserRepo struct {
	domainrepository.UserRepository
}

func (fakeUserRepo) FindByIDs(_ context.Context, ids []string) (map[string]*entity.User, error) {
	users := make(map[string]*entity.User, len(ids))
	for _, id := range ids {
		users[id] = &entity.User{ID: id, DisplayName: id}
	}
	return users, nil
}

type fakeStorage struct {
	domainservice.StorageService
	uploadKeys []string
	deleted    []string
}

func (s *fakeStorage) GenerateUploadURL(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	s.uploadKeys = append(s.uploadKeys, key)
	return "https://storage/put/" + key, nil
}

func (s *fakeStorage) GenerateDownloadURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://storage/get/" + key, nil
}

func (s *fakeStorage) DeleteObject(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	return nil
}

type fakeNotifier struct {
	workspaceIDs []string
}

func (n *fakeNotifier) NotifyCustomEmojisChanged(workspaceID string) {
	n.workspaceIDs = append(n.workspaceIDs, workspaceID)
}

type fixture struct {
	uc       *Interactor
	repo     *fakeEmojiRepo
	storage  *fakeStorage
	notifier *fakeNotifier
	recorder *audittest.Recorder
}

func newFixture() fixture {
	repo := &fakeEmojiRepo{emojis: map[string]*entity.CustomEmoji{
		"e1": {ID: "e1", WorkspaceID: workspaceID, Name: "party", StorageKey: "custom-emojis/ws/e1", CreatedBy: creatorID},
		"e2": {ID: "e2", WorkspaceID: "other-ws", Name: "party", StorageKey: "custom-emojis/other-ws/e2", CreatedBy: creatorID},
	}}
	storage := &fakeStorage{}
	notifier := &fakeNotifier{}
	recorder := &audittest.Recorder{}
	uc := New(repo, fakeUserRepo{}, stubWorkspaceRepo{}, stubPermission{}, storage, notifier, recorder)
	return fixture{uc: uc, repo: repo, storage: storage, notifier: notifier, recorder: recorder}
}

func TestList(t *testing.T) {
	t.Run("ワークスペースの絵文字だけを返し、削除できるかを付ける", func(t *testing.T) {
		f := newFixture()
		out, err := f.uc.List(context.Background(), ListInput{WorkspaceID: workspaceID, UserID: otherID})
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 1 || out[0].Name != "party" || out[0].CanDelete {
			t.Fatalf("unexpected emojis: %+v", out)
		}
		if out[0].ImageURL != "https://storage/get/custom-emojis/ws/e1" {
			t.Fatalf("unexpected url: %s", out[0].ImageURL)
		}
	})

	t.Run("メンバー以外は参照できない", func(t *testing.T) {
		f := newFixture()
		_, err := f.uc.List(context.Background(), ListInput{WorkspaceID: workspaceID, UserID: outsiderID})
		if !errors.Is(err, domerr.ErrUnauthorized) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestPresign(t *testing.T) {
	t.Run("アップロード先はワークスペースごとの場所にする", func(t *testing.T) {
		f := newFixture()
		out, err := f.uc.Presign(context.Background(), PresignInput{WorkspaceID: workspaceID, UserID: otherID, ContentType: "image/png"})
		if err != nil {
			t.Fatal(err)
		}
		if f.storage.uploadKeys[0] != "custom-emojis/ws/"+out.UploadID {
			t.Fatalf("unexpected key: %s", f.storage.uploadKeys[0])
		}
	})

	t.Run("登録の権限がなければ発行しない", func(t *testing.T) {
		f := newFixture()
		_, err := f.uc.Presign(context.Background(), PresignInput{WorkspaceID: workspaceID, UserID: guestID, ContentType: "image/png"})
		if !errors.Is(err, domerr.ErrUnauthorized) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestCreate(t *testing.T) {
	t.Run("登録して監査ログに残し、全員に知らせる", func(t *testing.T) {
		f := newFixture()
		out, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, UserID: otherID, Name: "tada", UploadID: uploadID})
		if err != nil {
			t.Fatal(err)
		}
		if f.repo.emojis[uploadID].StorageKey != "custom-emojis/ws/"+uploadID || !out.CanDelete {
			t.Fatalf("unexpected emoji: %+v", f.repo.emojis[uploadID])
		}
		if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionCustomEmojiCreated}) {
			t.Fatalf("unexpected audit: %v", f.recorder.Actions())
		}
		if !slices.Equal(f.notifier.workspaceIDs, []string{workspaceID}) {
			t.Fatalf("unexpected notification: %v", f.notifier.workspaceIDs)
		}
	})

	tests := []struct {
		name  string
		input CreateInput
		want  error
	}{
		{name: "同じ名前は登録できない", input: CreateInput{WorkspaceID: workspaceID, UserID: otherID, Name: "party", UploadID: uploadID}, want: domerr.ErrCustomEmojiNameExists},
		{name: "ゲストは既定で登録できない", input: CreateInput{WorkspaceID: workspaceID, UserID: guestID, Name: "tada", UploadID: uploadID}, want: domerr.ErrUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			_, err := f.uc.Create(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
			if len(f.recorder.Logs) != 0 || len(f.notifier.workspaceIDs) != 0 {
				t.Fatal("失敗したのに記録・通知された")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		emojiID string
		want    error
	}{
		{name: "登録者は削除できる", userID: creatorID, emojiID: "e1"},
		{name: "管理者は削除できる", userID: adminID, emojiID: "e1"},
		{name: "他のメンバーは削除できない", userID: otherID, emojiID: "e1", want: domerr.ErrUnauthorized},
		{name: "他のワークスペースの絵文字は見つからない", userID: adminID, emojiID: "e2", want: ErrEmojiNotFound},
		{name: "メンバー以外は削除できない", userID: outsiderID, emojiID: "e1", want: domerr.ErrUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			err := f.uc.Delete(context.Background(), DeleteInput{WorkspaceID: workspaceID, UserID: tt.userID, EmojiID: tt.emojiID})
			if tt.want != nil {
				if !errors.Is(err, tt.want) {
					t.Fatalf("got %v, want %v", err, tt.want)
				}
				if _, ok := f.repo.emojis[tt.emojiID]; !ok {
					t.Fatal("削除されてしまった")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := f.repo.emojis[tt.emojiID]; ok {
				t.Fatal("削除されていない")
			}
			if !slices.Equal(f.storage.deleted, []string{"custom-emojis/ws/e1"}) {
				t.Fatalf("unexpected storage deletion: %v", f.storage.deleted)
			}
			if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionCustomEmojiDeleted}) {
				t.Fatalf("unexpected audit: %v", f.recorder.Actions())
			}
			if !slices.Equal(f.notifier.workspaceIDs, []string{workspaceID}) {
				t.Fatalf("unexpected notification: %v", f.notifier.workspaceIDs)
			}
		})
	}
}
