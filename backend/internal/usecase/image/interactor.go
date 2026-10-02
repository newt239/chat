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

var ErrWorkspaceRequired = fmt.Errorf("%w: workspace_id を指定してください", domerr.ErrValidation)

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
	storageConfig domainservice.StorageConfig
	// 配信用の URL の起点になるバックエンドの公開 URL
	publicBaseURL string
}

func NewInteractor(
	workspaceRepo domainrepository.WorkspaceRepository,
	storage domainservice.StorageService,
	storageConfig domainservice.StorageConfig,
	publicBaseURL string,
) *Interactor {
	return &Interactor{
		workspaceRepo: workspaceRepo,
		storage:       storage,
		storageConfig: storageConfig,
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
	url, err := i.storage.GenerateUploadURL(ctx, key, input.ContentType, input.SizeBytes, 0)
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
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return "", domerr.ErrUnauthorized
	}
	if input.Purpose == PurposeWorkspaceIcon {
		return "workspaces/" + input.WorkspaceID, nil
	}
	return "apps/" + input.WorkspaceID, nil
}
