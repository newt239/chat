package image

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

const memberID = "11111111-1111-1111-1111-111111111111"

type fakeWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (fakeWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	if userID != memberID {
		return nil, nil
	}
	return &entity.WorkspaceMember{UserID: userID}, nil
}

type fakeStorage struct{ key string }

func (s *fakeStorage) GenerateUploadURL(_ context.Context, key, _ string, _ int64, _ time.Duration) (string, error) {
	s.key = key
	return "https://storage.example.com/" + key, nil
}
func (*fakeStorage) GenerateDownloadURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (*fakeStorage) DeleteObject(context.Context, string) error { return nil }

func TestPresign(t *testing.T) {
	storage := &fakeStorage{}
	uc := NewInteractor(fakeWorkspaceRepo{}, storage, "https://api.example.com/")

	out, err := uc.Presign(context.Background(), PresignInput{UserID: memberID, Purpose: PurposeAvatar, ContentType: "image/png", SizeBytes: 10})
	if err != nil {
		t.Fatalf("発行できません: %v", err)
	}
	if !strings.HasPrefix(storage.key, "images/avatars/"+memberID+"/") || out.ImageURL != "https://api.example.com/"+storage.key {
		t.Fatalf("置き場所か配信 URL が正しくありません: key=%s url=%s", storage.key, out.ImageURL)
	}

	if _, err := uc.Presign(context.Background(), PresignInput{UserID: memberID, WorkspaceID: "ws", Purpose: PurposeWorkspaceIcon}); err != nil || !strings.HasPrefix(storage.key, "images/workspaces/ws/") {
		t.Fatalf("ワークスペースのアイコンを置けません: key=%s err=%v", storage.key, err)
	}
	if _, err := uc.Presign(context.Background(), PresignInput{UserID: memberID, Purpose: PurposeAppIcon}); !errors.Is(err, domerr.ErrValidation) {
		t.Fatalf("workspace_id なしで発行できています: %v", err)
	}
	if _, err := uc.Presign(context.Background(), PresignInput{UserID: "other", WorkspaceID: "ws", Purpose: PurposeAppIcon}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("メンバー以外が発行できています: %v", err)
	}
}
