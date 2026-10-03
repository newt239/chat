package customemoji

import (
	"context"
	"fmt"
	"log/slog"
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

var ErrEmojiNotFound = domerr.New(domerr.ErrNotFound, "指定されたカスタム絵文字が見つかりません")

type Interactor struct {
	emojiRepo     domainrepository.CustomEmojiRepository
	userRepo      domainrepository.UserRepository
	workspaceRepo domainrepository.WorkspaceRepository
	permissionSvc domainservice.PermissionService
	storage       domainservice.StorageService
	notifier      Notifier
	recorder      audit.Recorder
}

func New(
	emojiRepo domainrepository.CustomEmojiRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	permissionSvc domainservice.PermissionService,
	storage domainservice.StorageService,
	notifier Notifier,
	recorder audit.Recorder,
) *Interactor {
	return &Interactor{
		emojiRepo:     emojiRepo,
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		permissionSvc: permissionSvc,
		storage:       storage,
		notifier:      notifier,
		recorder:      recorder,
	}
}

func (i *Interactor) List(ctx context.Context, input ListInput) ([]Output, error) {
	member, err := domainservice.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}
	emojis, err := i.emojiRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load custom emojis: %w", err)
	}
	return i.toOutputs(ctx, emojis, member)
}

// Presign は画像のアップロード先を発行します。登録は画像を置いたあと Create で行います
func (i *Interactor) Presign(ctx context.Context, input PresignInput) (*PresignOutput, error) {
	if _, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.UserID, entity.PermissionCreateCustomEmoji); err != nil {
		return nil, err
	}
	uploadID := uuid.NewString()
	url, err := i.storage.GenerateUploadURL(ctx, storageKey(input.WorkspaceID, uploadID), input.ContentType, input.SizeBytes, domainservice.UploadURLExpires)
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
	// アップロード先はサーバーで組み立て直し、他のワークスペースの画像を指せないようにする
	emoji := &entity.CustomEmoji{
		ID:          input.UploadID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		StorageKey:  storageKey(input.WorkspaceID, input.UploadID),
		CreatedBy:   input.UserID,
	}
	// 同じ名前があれば ErrCustomEmojiNameExists をそのまま返す
	if err := i.emojiRepo.Create(ctx, emoji); err != nil {
		return nil, err
	}
	i.record(ctx, emoji, input.UserID, entity.AuditActionCustomEmojiCreated)
	i.notifier.NotifyCustomEmojisChanged(emoji.WorkspaceID)

	outputs, err := i.toOutputs(ctx, []*entity.CustomEmoji{emoji}, member)
	if err != nil {
		return nil, err
	}
	return &outputs[0], nil
}

func (i *Interactor) Delete(ctx context.Context, input DeleteInput) error {
	member, err := domainservice.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID)
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
		slog.WarnContext(ctx, "カスタム絵文字の画像の削除に失敗しました", "storageKey", emoji.StorageKey, "error", err)
	}
	i.record(ctx, emoji, input.UserID, entity.AuditActionCustomEmojiDeleted)
	i.notifier.NotifyCustomEmojisChanged(emoji.WorkspaceID)
	return nil
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
			CreatedBy: messageuc.UserInfoOf(e.CreatedBy, creators),
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
