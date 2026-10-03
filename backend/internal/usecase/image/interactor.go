package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

// KeyPrefix は画像を置くストレージのキーの接頭辞です。配信の URL も同じパスにします
const KeyPrefix = "images/"

type Purpose int

const (
	PurposeAvatar Purpose = iota
	PurposeWorkspaceIcon
	PurposeAppIcon
)

var ErrWorkspaceRequired = domerr.New(domerr.ErrValidation, "workspace_id を指定してください")

type PresignInput struct {
	UserID      string
	WorkspaceID string
	Purpose     Purpose
	ContentType string
	SizeBytes   int64
}

type PresignOutput struct {
	UploadURL string
	ImageURL  string
}

type Interactor struct {
	workspaceRepo domainrepository.WorkspaceRepository
	storage       domainservice.StorageService
	// 配信用の URL の起点になるバックエンドの公開 URL
	publicBaseURL string
}

func New(
	workspaceRepo domainrepository.WorkspaceRepository,
	storage domainservice.StorageService,
	publicBaseURL string,
) *Interactor {
	return &Interactor{
		workspaceRepo: workspaceRepo,
		storage:       storage,
		publicBaseURL: strings.TrimSuffix(publicBaseURL, "/"),
	}
}

// Presign はアイコン画像のアップロード先と、期限のない配信用の URL を発行します
func (i *Interactor) Presign(ctx context.Context, input PresignInput) (*PresignOutput, error) {
	dir, err := i.directory(ctx, input)
	if err != nil {
		return nil, err
	}
	key := KeyPrefix + dir + "/" + uuid.NewString()
	url, err := i.storage.GenerateUploadURL(ctx, key, input.ContentType, input.SizeBytes, domainservice.UploadURLExpires)
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload: %w", err)
	}
	return &PresignOutput{UploadURL: url, ImageURL: i.publicBaseURL + "/" + key}, nil
}

// directory は用途ごとの置き場所を返します。ワークスペースのものはメンバーだけが置けます
func (i *Interactor) directory(ctx context.Context, input PresignInput) (string, error) {
	if input.Purpose == PurposeAvatar {
		return "avatars/" + input.UserID, nil
	}
	if input.WorkspaceID == "" {
		return "", ErrWorkspaceRequired
	}
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID); err != nil {
		return "", err
	}
	if input.Purpose == PurposeWorkspaceIcon {
		return "workspaces/" + input.WorkspaceID, nil
	}
	return "apps/" + input.WorkspaceID, nil
}
