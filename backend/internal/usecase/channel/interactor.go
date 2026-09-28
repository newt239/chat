package channel

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
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/usecase/audit"
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

var (
	ErrUnauthorized      = errors.New("この操作を行う権限がありません")
	ErrWorkspaceNotFound = errors.New("ワークスペースが見つかりません")
	ErrChannelNotFound   = errors.New("チャンネルが見つかりません")
	ErrCannotArchiveDM   = errors.New("DM はアーカイブできません")
)

type ChannelUseCase interface {
	GetChannel(ctx context.Context, input GetChannelInput) (*ChannelOutput, error)
	DeleteChannel(ctx context.Context, input DeleteChannelInput) error
	ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error)
	UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error)
	SetArchived(ctx context.Context, input SetArchivedInput) (*ChannelOutput, error)
}

type channelInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	readStateRepo     domainrepository.ReadStateRepository
	txManager         domaintransaction.Manager
	systemMessageUC   systemmessage.UseCase
	channelAccessSvc  domainservice.ChannelAccessService
	permissionSvc     domainservice.PermissionService
	recorder          audit.Recorder
}

func NewChannelInteractor(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	readStateRepo domainrepository.ReadStateRepository,
	txManager domaintransaction.Manager,
	systemMessageUC systemmessage.UseCase,
	channelAccessSvc domainservice.ChannelAccessService,
	permissionSvc domainservice.PermissionService,
	recorder audit.Recorder,
) ChannelUseCase {
	return &channelInteractor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		readStateRepo:     readStateRepo,
		txManager:         txManager,
		systemMessageUC:   systemMessageUC,
		channelAccessSvc:  channelAccessSvc,
		permissionSvc:     permissionSvc,
		recorder:          recorder,
	}
}

func (i *channelInteractor) ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error) {
	if err := validateWorkspaceID(input.WorkspaceID); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	channels, err := i.channelRepo.FindAccessibleChannels(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	// チャネルIDリストを作成
	channelIDs := make([]string, len(channels))
	for idx, ch := range channels {
		channelIDs[idx] = ch.ID
	}

	unreadCounts, err := i.readStateRepo.GetUnreadCountBatch(ctx, channelIDs, input.UserID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get unread counts: userID=%s err=%v\n", input.UserID, err)
		unreadCounts = make(map[string]int)
	}

	mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, channelIDs, input.UserID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get unread mention counts: userID=%s err=%v\n", input.UserID, err)
		mentionCounts = make(map[string]int)
	}

	output := make([]ChannelOutput, 0, len(channels))
	for _, ch := range channels {
		output = append(output, toChannelOutputWithUnread(ch, unreadCounts[ch.ID], mentionCounts[ch.ID] > 0))
	}

	return output, nil
}

func (i *channelInteractor) CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error) {
	if err := validateWorkspaceID(input.WorkspaceID); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	permission := entity.PermissionCreatePublicChannel
	if input.IsPrivate {
		permission = entity.PermissionCreatePrivateChannel
	}
	if _, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.UserID, permission); err != nil {
		return nil, err
	}

	channel, err := entity.NewChannel(entity.ChannelParams{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		IsPrivate:   input.IsPrivate,
		CreatedBy:   input.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create channel entity: %w", err)
	}

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.channelRepo.Create(txCtx, channel); err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}

		if channel.IsPrivate {
			member := &entity.ChannelMember{
				ChannelID: channel.ID,
				UserID:    input.UserID,
				Role:      entity.ChannelRoleAdmin,
				JoinedAt:  time.Now(),
			}
			if err := i.channelMemberRepo.AddMember(txCtx, member); err != nil {
				return fmt.Errorf("failed to add creator to private channel: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	i.recordChannelAction(ctx, channel, input.UserID, entity.AuditActionChannelCreated)
	output := toChannelOutputWithUnread(channel, 0, false)
	return &output, nil
}

func (i *channelInteractor) GetChannel(ctx context.Context, input GetChannelInput) (*ChannelOutput, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	output := toChannelOutput(ch)
	return &output, nil
}

func (i *channelInteractor) DeleteChannel(ctx context.Context, input DeleteChannelInput) error {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	if err := i.ensureCanManage(ctx, ch, input.UserID); err != nil {
		return err
	}

	if err := i.channelRepo.Delete(ctx, input.ChannelID); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	i.recordChannelAction(ctx, ch, input.UserID, entity.AuditActionChannelDeleted)
	return nil
}

// SetArchived はチャンネルをアーカイブ・解除します。アーカイブ中は投稿できません
func (i *channelInteractor) SetArchived(ctx context.Context, input SetArchivedInput) (*ChannelOutput, error) {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}
	if ch.Type == entity.ChannelTypeDM || ch.Type == entity.ChannelTypeGroupDM {
		return nil, ErrCannotArchiveDM
	}
	if err := i.ensureCanManage(ctx, ch, input.UserID); err != nil {
		return nil, err
	}

	if (ch.ArchivedAt != nil) != input.Archived {
		action := entity.AuditActionChannelUnarchived
		ch.ArchivedAt = nil
		if input.Archived {
			now := time.Now()
			ch.ArchivedAt = &now
			action = entity.AuditActionChannelArchived
		}
		if err := i.channelRepo.Update(ctx, ch); err != nil {
			return nil, fmt.Errorf("failed to update channel: %w", err)
		}
		i.recordChannelAction(ctx, ch, input.UserID, action)
	}

	out := toChannelOutput(ch)
	return &out, nil
}

// ensureCanManage はチャンネルの作成者かワークスペースの管理者であることを確認します
func (i *channelInteractor) ensureCanManage(ctx context.Context, ch *entity.Channel, userID string) error {
	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil || (ch.CreatedBy != userID && !member.IsAdmin()) {
		return ErrUnauthorized
	}
	return nil
}

func (i *channelInteractor) recordChannelAction(ctx context.Context, ch *entity.Channel, actorID string, action entity.AuditAction) {
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: ch.WorkspaceID,
		ActorID:     &actorID,
		Action:      action,
		TargetType:  entity.AuditTargetChannel,
		TargetID:    ch.ID,
		TargetLabel: ch.Name,
		Metadata:    map[string]string{"private": fmt.Sprint(ch.IsPrivate)},
	})
}

func (i *channelInteractor) UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error) {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	// 権限: ワークスペースの管理権限（チャンネル編集権限として流用）
	wsMember, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, input.UserID)
	if err != nil || wsMember == nil || !wsMember.IsAdmin() {
		return nil, ErrUnauthorized
	}

	// 変更適用
	originalName := ch.Name
	originalDesc := ch.Description
	originalPrivate := ch.IsPrivate

	nameChanged := false
	descChanged := false
	privChanged := false

	if input.Name != nil {
		_ = ch.ChangeName(*input.Name)
		nameChanged = (originalName != ch.Name)
	}
	if input.Description != nil {
		ch.Description = input.Description
		ch.UpdatedAt = time.Now().UTC()
		// detect change
		old := ""
		if originalDesc != nil {
			old = *originalDesc
		}
		now := ""
		if ch.Description != nil {
			now = *ch.Description
		}
		descChanged = (old != now)
	}
	if input.IsPrivate != nil {
		ch.IsPrivate = *input.IsPrivate
		ch.UpdatedAt = time.Now().UTC()
		privChanged = (originalPrivate != ch.IsPrivate)
	}

	if err := i.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	// 変更に応じてシステムメッセージ作成
	actorID := input.UserID
	if i.systemMessageUC != nil {
		if nameChanged {
			if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
				ChannelID: ch.ID,
				Kind:      entity.SystemMessageKindChannelNameChanged,
				Payload:   map[string]any{"from": originalName, "to": ch.Name},
				ActorID:   &actorID,
			}); err != nil {
				fmt.Printf("[WARN] Failed to create system message for channel name change: channelID=%s err=%v\n", ch.ID, err)
			}
		}
		if descChanged {
			from := ""
			if originalDesc != nil {
				from = *originalDesc
			}
			to := ""
			if ch.Description != nil {
				to = *ch.Description
			}
			if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
				ChannelID: ch.ID,
				Kind:      entity.SystemMessageKindChannelDescriptionChanged,
				Payload:   map[string]any{"from": from, "to": to},
				ActorID:   &actorID,
			}); err != nil {
				fmt.Printf("[WARN] Failed to create system message for channel description change: channelID=%s err=%v\n", ch.ID, err)
			}
		}
		if privChanged {
			from := "public"
			if originalPrivate {
				from = "private"
			}
			to := "public"
			if ch.IsPrivate {
				to = "private"
			}
			if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
				ChannelID: ch.ID,
				Kind:      entity.SystemMessageKindChannelPrivacyChanged,
				Payload:   map[string]any{"from": from, "to": to},
				ActorID:   &actorID,
			}); err != nil {
				fmt.Printf("[WARN] Failed to create system message for channel privacy change: channelID=%s err=%v\n", ch.ID, err)
			}
		}
	}

	out := toChannelOutput(ch)
	return &out, nil
}

func toChannelOutput(channel *entity.Channel) ChannelOutput {
	return toChannelOutputWithUnread(channel, 0, false)
}

func toChannelOutputWithUnread(channel *entity.Channel, unreadCount int, hasMention bool) ChannelOutput {
	return ChannelOutput{
		ID:          channel.ID,
		WorkspaceID: channel.WorkspaceID,
		Name:        channel.Name,
		Description: channel.Description,
		IsPrivate:   channel.IsPrivate,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   channel.CreatedAt,
		UpdatedAt:   channel.UpdatedAt,
		UnreadCount: unreadCount,
		HasMention:  hasMention,
		ArchivedAt:  channel.ArchivedAt,
	}
}

func validateUUID(id string, label string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: invalid %s format", domerr.ErrValidation, label)
	}
	return nil
}

func validateWorkspaceID(id string) error {
	// ワークスペースIDはslug形式（3-12文字の英小文字、数字、ハイフン）またはUUID形式を許可
	if err := entity.ValidateWorkspaceSlug(id); err != nil {
		// slug形式でない場合、UUID形式かチェック
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("%w: invalid workspace ID format", domerr.ErrValidation)
		}
	}
	return nil
}
