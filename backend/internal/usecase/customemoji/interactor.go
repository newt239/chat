package customemoji

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

// imageURLExpires は一覧で返す画像 URL の有効期限です。表示のたびに取り直さないよう長めにします
const imageURLExpires = 12 * time.Hour

var (
	ErrEmojiNotFound = errors.New("指定されたカスタム絵文字が見つかりません")
	ErrNameExists    = errors.New("同じ名前のカスタム絵文字がすでにあります")
	ErrInvalidName   = fmt.Errorf("%w: 名前は英小文字・数字・_・- の 32 文字以内で指定してください", domerr.ErrValidation)
)

type Interactor struct {
	emojiRepo     domainrepository.CustomEmojiRepository
	userRepo      domainrepository.UserRepository
	workspaceRepo domainrepository.WorkspaceRepository
	permissionSvc domainservice.PermissionService
	storage       domainservice.StorageService
	notifier      Notifier
	recorder      audit.Recorder
	logger        domainservice.Logger
}

func NewInteractor(
	emojiRepo domainrepository.CustomEmojiRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	permissionSvc domainservice.PermissionService,
	storage domainservice.StorageService,
	notifier Notifier,
	recorder audit.Recorder,
	logger domainservice.Logger,
) *Interactor {
	return &Interactor{
		emojiRepo:     emojiRepo,
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		permissionSvc: permissionSvc,
		storage:       storage,
		notifier:      notifier,
		recorder:      recorder,
		logger:        logger,
	}
}

func (i *Interactor) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	member, err := i.ensureMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}
	emojis, err := i.emojiRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load custom emojis: %w", err)
	}
	expiresAt := time.Now().Add(imageURLExpires)
	outputs, err := i.toOutputs(ctx, emojis, member)
	if err != nil {
		return nil, err
	}
	return &ListOutput{Emojis: outputs, ExpiresAt: expiresAt}, nil
}

// Presign は画像のアップロード先を発行します。登録は画像を置いたあと Create で行います
func (i *Interactor) Presign(ctx context.Context, input PresignInput) (*PresignOutput, error) {
	if _, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.UserID, entity.PermissionCreateCustomEmoji); err != nil {
		return nil, err
	}
	uploadID := uuid.NewString()
	url, err := i.storage.GenerateUploadURL(ctx, storageKey(input.WorkspaceID, uploadID), input.ContentType, input.SizeBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload: %w", err)
	}
	return &PresignOutput{UploadID: uploadID, UploadURL: url}, nil
}

func (i *Interactor) Create(ctx context.Context, input CreateInput) (*Output, error) {
	member, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.UserID, entity.PermissionCreateCustomEmoji)
	if err != nil {
		return nil, err
	}
	if !entity.IsValidCustomEmojiName(input.Name) {
		return nil, ErrInvalidName
	}
	if uuid.Validate(input.UploadID) != nil {
		return nil, fmt.Errorf("%w: upload_id が不正です", domerr.ErrValidation)
	}
	// アップロード先はサーバーで組み立て直し、他のワークスペースの画像を指せないようにする
	emoji := &entity.CustomEmoji{
		ID:          input.UploadID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		StorageKey:  storageKey(input.WorkspaceID, input.UploadID),
		CreatedBy:   input.UserID,
	}
	if err := i.emojiRepo.Create(ctx, emoji); err != nil {
		if errors.Is(err, domerr.ErrConflict) {
			return nil, ErrNameExists
		}
		return nil, fmt.Errorf("failed to create custom emoji: %w", err)
	}
	i.record(ctx, emoji, input.UserID, entity.AuditActionCustomEmojiCreated)
	i.notifier.NotifyCustomEmojiCreated(emoji.WorkspaceID, Notification{ID: emoji.ID, Name: emoji.Name})

	outputs, err := i.toOutputs(ctx, []*entity.CustomEmoji{emoji}, member)
	if err != nil {
		return nil, err
	}
	return &outputs[0], nil
}

func (i *Interactor) Delete(ctx context.Context, input DeleteInput) error {
	member, err := i.ensureMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return err
	}
	emoji, err := i.emojiRepo.FindByID(ctx, input.EmojiID)
	if err != nil {
		return fmt.Errorf("failed to load custom emoji: %w", err)
	}
	if emoji == nil || emoji.WorkspaceID != input.WorkspaceID {
		return ErrEmojiNotFound
	}
	if !canDelete(emoji, member) {
		return domerr.ErrUnauthorized
	}
	if err := i.emojiRepo.Delete(ctx, emoji.ID); err != nil {
		return fmt.Errorf("failed to delete custom emoji: %w", err)
	}
	// 画像が残っても表示されることはないため、削除の失敗は記録だけにとどめる
	if err := i.storage.DeleteObject(ctx, emoji.StorageKey); err != nil {
		i.logger.Warn("カスタム絵文字の画像の削除に失敗しました",
			domainservice.LogField{Key: "storageKey", Value: emoji.StorageKey},
			domainservice.LogField{Key: "error", Value: err.Error()},
		)
	}
	i.record(ctx, emoji, input.UserID, entity.AuditActionCustomEmojiDeleted)
	i.notifier.NotifyCustomEmojiDeleted(emoji.WorkspaceID, Notification{ID: emoji.ID, Name: emoji.Name})
	return nil
}

func (i *Interactor) ensureMember(ctx context.Context, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}
	return member, nil
}

func (i *Interactor) toOutputs(ctx context.Context, emojis []*entity.CustomEmoji, viewer *entity.WorkspaceMember) ([]Output, error) {
	creatorIDs := make([]string, 0, len(emojis))
	for _, e := range emojis {
		creatorIDs = append(creatorIDs, e.CreatedBy)
	}
	creators, err := i.userRepo.FindByIDs(ctx, creatorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load creators: %w", err)
	}
	byID := make(map[string]*entity.User, len(creators))
	for _, u := range creators {
		byID[u.ID] = u
	}

	outputs := make([]Output, 0, len(emojis))
	for _, e := range emojis {
		url, err := i.storage.GenerateDownloadURL(ctx, e.StorageKey, imageURLExpires)
		if err != nil {
			return nil, fmt.Errorf("failed to presign download: %w", err)
		}
		outputs = append(outputs, Output{
			ID:        e.ID,
			Name:      e.Name,
			ImageURL:  url,
			CreatedBy: messageuc.UserInfoOf(e.CreatedBy, byID),
			CreatedAt: e.CreatedAt,
			CanDelete: canDelete(e, viewer),
		})
	}
	return outputs, nil
}

func (i *Interactor) record(ctx context.Context, emoji *entity.CustomEmoji, actorID string, action entity.AuditAction) {
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: emoji.WorkspaceID,
		ActorID:     &actorID,
		Action:      action,
		TargetType:  entity.AuditTargetCustomEmoji,
		TargetID:    emoji.ID,
		TargetLabel: ":" + emoji.Name + ":",
	})
}

func canDelete(emoji *entity.CustomEmoji, member *entity.WorkspaceMember) bool {
	return emoji.CreatedBy == member.UserID || member.IsAdmin()
}

func storageKey(workspaceID, uploadID string) string {
	return fmt.Sprintf("custom-emojis/%s/%s", workspaceID, uploadID)
}
