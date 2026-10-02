package channel

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
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
	ErrChannelNameExists    = errors.New("同じ名前のチャンネルがすでに存在します")
	ErrChannelHasChildren   = errors.New("下の階層にチャンネルがあるため削除できません")
	ErrMemberNotInWorkspace = fmt.Errorf("%w: ワークスペースのメンバーではないユーザーが含まれています", domerr.ErrValidation)
	ErrCannotArchiveDM      = errors.New("DM はアーカイブできません")
)

type ChannelUseCase interface {
	GetChannel(ctx context.Context, input GetChannelInput) (*ChannelOutput, error)
	DeleteChannel(ctx context.Context, input DeleteChannelInput) error
	ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error)
	UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error)
	SetArchived(ctx context.Context, input SetArchivedInput) (*ChannelOutput, error)
	SetChannelStarred(ctx context.Context, input SetChannelStarredInput) error
	SetChannelMuted(ctx context.Context, input SetChannelMutedInput) error
	ListBrowsableChannels(ctx context.Context, input ListChannelsInput) ([]BrowsableChannelOutput, error)
	SearchBrowsableChannels(ctx context.Context, input SearchBrowsableChannelsInput) (*SearchBrowsableChannelsOutput, error)
}

type channelInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	channelStarRepo   domainrepository.ChannelStarRepository
	channelMuteRepo   domainrepository.ChannelMuteRepository
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
	channelStarRepo domainrepository.ChannelStarRepository,
	channelMuteRepo domainrepository.ChannelMuteRepository,
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
		channelStarRepo:   channelStarRepo,
		channelMuteRepo:   channelMuteRepo,
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
		return nil, domerr.ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}

	joined, err := i.channelRepo.FindAccessibleChannels(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	ancestors, err := i.findMissingAncestors(ctx, input.WorkspaceID, input.UserID, joined)
	if err != nil {
		return nil, err
	}

	joinedIDs := make([]string, len(joined))
	for idx, ch := range joined {
		joinedIDs[idx] = ch.ID
	}

	unreadCounts, err := i.readStateRepo.GetUnreadCountBatch(ctx, joinedIDs, input.UserID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get unread counts: userID=%s err=%v\n", input.UserID, err)
		unreadCounts = make(map[string]int)
	}

	mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, joinedIDs, input.UserID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get unread mention counts: userID=%s err=%v\n", input.UserID, err)
		mentionCounts = make(map[string]int)
	}

	all := slices.Concat(joined, ancestors)
	allIDs := make([]string, len(all))
	for idx, ch := range all {
		allIDs[idx] = ch.ID
	}
	starred, err := i.channelStarRepo.FindStarredChannelIDs(ctx, input.UserID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch starred channels: %w", err)
	}
	muted, err := i.channelMuteRepo.FindMutedChannelIDs(ctx, input.UserID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch muted channels: %w", err)
	}
	lastMessageAt, err := i.channelRepo.FindLastMessageAtBatch(ctx, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch last message times: %w", err)
	}

	output := make([]ChannelOutput, 0, len(all))
	for _, ch := range joined {
		out := toChannelOutputWithUnread(ch, unreadCounts[ch.ID], mentionCounts[ch.ID])
		out.IsMember = true
		output = append(output, out)
	}
	for _, ch := range ancestors {
		output = append(output, toChannelOutput(ch))
	}
	for idx := range output {
		out := &output[idx]
		out.IsStarred = starred[out.ID]
		out.IsMuted = muted[out.ID]
		if at, ok := lastMessageAt[out.ID]; ok {
			out.LastMessageAt = &at
		}
	}
	slices.SortFunc(output, func(a, b ChannelOutput) int { return strings.Compare(a.Name, b.Name) })

	return output, nil
}

func (i *channelInteractor) ListBrowsableChannels(ctx context.Context, input ListChannelsInput) ([]BrowsableChannelOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}

	channels, err := i.channelRepo.FindBrowsableChannels(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}
	return i.toBrowsableOutputs(ctx, input.WorkspaceID, input.UserID, channels)
}

func (i *channelInteractor) SearchBrowsableChannels(ctx context.Context, input SearchBrowsableChannelsInput) (*SearchBrowsableChannelsOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}

	channels, total, err := i.channelRepo.SearchBrowsableChannels(ctx, input.WorkspaceID, input.UserID, domainrepository.BrowsableChannelFilter{
		Query:      input.Query,
		Membership: input.Membership,
		Sort:       input.Sort,
		Limit:      input.PerPage,
		Offset:     (input.Page - 1) * input.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search channels: %w", err)
	}
	outputs, err := i.toBrowsableOutputs(ctx, input.WorkspaceID, input.UserID, channels)
	if err != nil {
		return nil, err
	}
	return &SearchBrowsableChannelsOutput{Channels: outputs, Total: total}, nil
}

// toBrowsableOutputs はメンバー数と自分が参加しているかを付けます
func (i *channelInteractor) toBrowsableOutputs(ctx context.Context, workspaceID, userID string, channels []*entity.Channel) ([]BrowsableChannelOutput, error) {
	ids := make([]string, len(channels))
	for idx, ch := range channels {
		ids[idx] = ch.ID
	}
	memberCounts, err := i.channelRepo.CountMembersBatch(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}
	joined, err := i.channelRepo.FindAccessibleChannels(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch joined channels: %w", err)
	}
	joinedIDs := make(map[string]bool, len(joined))
	for _, ch := range joined {
		joinedIDs[ch.ID] = true
	}

	output := make([]BrowsableChannelOutput, 0, len(channels))
	for _, ch := range channels {
		out := toChannelOutput(ch)
		out.IsMember = joinedIDs[ch.ID]
		output = append(output, BrowsableChannelOutput{Channel: out, MemberCount: memberCounts[ch.ID]})
	}
	return output, nil
}

// findMissingAncestors はツリーを組み立てるため、参加していないが閲覧できる祖先チャンネルを返します
func (i *channelInteractor) findMissingAncestors(ctx context.Context, workspaceID, userID string, joined []*entity.Channel) ([]*entity.Channel, error) {
	joinedNames := make(map[string]bool, len(joined))
	for _, ch := range joined {
		joinedNames[ch.Name] = true
	}

	missing := make([]string, 0)
	for _, ch := range joined {
		if ch.Type != entity.ChannelTypePublic && ch.Type != entity.ChannelTypePrivate {
			continue
		}
		for _, path := range entity.AncestorChannelPaths(ch.Name) {
			if !joinedNames[path] && !slices.Contains(missing, path) {
				missing = append(missing, path)
			}
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}

	ancestors, err := i.channelRepo.FindByNames(ctx, workspaceID, missing)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ancestor channels: %w", err)
	}
	return i.channelAccessSvc.FilterAccessible(ctx, ancestors, userID)
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
		return nil, domerr.ErrWorkspaceNotFound
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
		return nil, err
	}

	memberIDs := []string{input.UserID}
	for _, id := range input.MemberIDs {
		if slices.Contains(memberIDs, id) {
			continue
		}
		m, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, id)
		if err != nil {
			return nil, fmt.Errorf("failed to verify membership: %w", err)
		}
		if m == nil {
			return nil, ErrMemberNotInWorkspace
		}
		memberIDs = append(memberIDs, id)
	}

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		parentID, err := i.ensureAncestors(txCtx, channel)
		if err != nil {
			return err
		}
		channel.ParentID = parentID

		if err := i.channelRepo.Create(txCtx, channel); err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		for idx, userID := range memberIDs {
			role := entity.ChannelRoleMember
			if idx == 0 {
				role = entity.ChannelRoleAdmin
			}
			if err := i.addMember(txCtx, channel.ID, userID, role); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	i.recordChannelAction(ctx, channel, input.UserID, entity.AuditActionChannelCreated)
	output := toChannelOutput(channel)
	output.IsMember = true
	return &output, nil
}

// ensureAncestors は存在しない祖先チャンネルを作成し、直近の親の ID を返します
// 自動作成する祖先は子と同じ公開範囲にし、作成者を参加させます
func (i *channelInteractor) ensureAncestors(ctx context.Context, ch *entity.Channel) (*string, error) {
	ancestorPaths := entity.AncestorChannelPaths(ch.Name)
	existing, err := i.channelRepo.FindByNames(ctx, ch.WorkspaceID, append(ancestorPaths, ch.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ancestor channels: %w", err)
	}
	byName := make(map[string]*entity.Channel, len(existing))
	for _, c := range existing {
		byName[c.Name] = c
	}
	if byName[ch.Name] != nil {
		return nil, ErrChannelNameExists
	}

	var parentID *string
	for _, path := range ancestorPaths {
		parent := byName[path]
		if parent != nil {
			accessible, err := i.channelAccessSvc.FilterAccessible(ctx, []*entity.Channel{parent}, ch.CreatedBy)
			if err != nil {
				return nil, err
			}
			if len(accessible) == 0 {
				return nil, domerr.ErrUnauthorized
			}
		} else {
			parent, err = entity.NewChannel(entity.ChannelParams{
				WorkspaceID: ch.WorkspaceID,
				Name:        path,
				IsPrivate:   ch.IsPrivate,
				ParentID:    parentID,
				CreatedBy:   ch.CreatedBy,
			})
			if err != nil {
				return nil, err
			}
			if err := i.channelRepo.Create(ctx, parent); err != nil {
				return nil, fmt.Errorf("failed to create ancestor channel: %w", err)
			}
			if err := i.addMember(ctx, parent.ID, ch.CreatedBy, entity.ChannelRoleAdmin); err != nil {
				return nil, err
			}
		}
		parentID = &parent.ID
	}
	return parentID, nil
}

func (i *channelInteractor) addMember(ctx context.Context, channelID, userID string, role entity.ChannelRole) error {
	if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to add channel member: %w", err)
	}
	return nil
}

func (i *channelInteractor) GetChannel(ctx context.Context, input GetChannelInput) (*ChannelOutput, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify channel membership: %w", err)
	}
	starred, err := i.channelStarRepo.FindStarredChannelIDs(ctx, input.UserID, []string{ch.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch starred channels: %w", err)
	}
	muted, err := i.channelMuteRepo.FindMutedChannelIDs(ctx, input.UserID, []string{ch.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch muted channels: %w", err)
	}

	output := toChannelOutput(ch)
	output.IsMember = isMember
	output.IsStarred = starred[ch.ID]
	output.IsMuted = muted[ch.ID]
	return &output, nil
}

func (i *channelInteractor) SetChannelStarred(ctx context.Context, input SetChannelStarredInput) error {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}
	if err := i.channelStarRepo.SetStarred(ctx, input.UserID, input.ChannelID, input.Starred); err != nil {
		return fmt.Errorf("failed to update star: %w", err)
	}
	return nil
}

func (i *channelInteractor) SetChannelMuted(ctx context.Context, input SetChannelMutedInput) error {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}
	if err := i.channelMuteRepo.SetMuted(ctx, input.UserID, input.ChannelID, input.Muted); err != nil {
		return fmt.Errorf("failed to update mute: %w", err)
	}
	return nil
}

func (i *channelInteractor) DeleteChannel(ctx context.Context, input DeleteChannelInput) error {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return domerr.ErrChannelNotFound
	}

	if err := i.ensureCanManage(ctx, ch, input.UserID); err != nil {
		return err
	}

	descendants, err := i.channelRepo.FindDescendants(ctx, ch)
	if err != nil {
		return fmt.Errorf("failed to fetch descendant channels: %w", err)
	}
	if len(descendants) > 0 {
		return ErrChannelHasChildren
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
		return nil, domerr.ErrChannelNotFound
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
		return domerr.ErrUnauthorized
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
		return nil, domerr.ErrChannelNotFound
	}

	// 権限: ワークスペースの管理権限（チャンネル編集権限として流用）
	wsMember, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, input.UserID)
	if err != nil || wsMember == nil || !wsMember.IsAdmin() {
		return nil, domerr.ErrUnauthorized
	}

	// 変更適用
	originalName := ch.Name
	originalDesc := ch.Description
	originalPrivate := ch.IsPrivate

	descChanged := false
	privChanged := false

	if input.Name != nil {
		if err := ch.ChangeName(*input.Name); err != nil {
			return nil, err
		}
	}
	nameChanged := originalName != ch.Name
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

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if nameChanged {
			if err := i.renameDescendants(txCtx, ch, originalName); err != nil {
				return err
			}
		}
		if err := i.channelRepo.Update(txCtx, ch); err != nil {
			return fmt.Errorf("failed to update channel: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
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

// renameDescendants は名前を変えたチャンネルの子孫のパスを付け替えます
func (i *channelInteractor) renameDescendants(ctx context.Context, ch *entity.Channel, originalName string) error {
	conflicts, err := i.channelRepo.FindByNames(ctx, ch.WorkspaceID, []string{ch.Name})
	if err != nil {
		return fmt.Errorf("failed to check channel name: %w", err)
	}
	if slices.ContainsFunc(conflicts, func(c *entity.Channel) bool { return c.ID != ch.ID }) {
		return ErrChannelNameExists
	}

	descendants, err := i.channelRepo.FindDescendants(ctx, &entity.Channel{WorkspaceID: ch.WorkspaceID, Name: originalName})
	if err != nil {
		return fmt.Errorf("failed to fetch descendant channels: %w", err)
	}
	for _, d := range descendants {
		d.Name = ch.Name + strings.TrimPrefix(d.Name, originalName)
		if err := i.channelRepo.Update(ctx, d); err != nil {
			return fmt.Errorf("failed to rename descendant channel: %w", err)
		}
	}
	return nil
}

func toChannelOutput(channel *entity.Channel) ChannelOutput {
	return toChannelOutputWithUnread(channel, 0, 0)
}

func toChannelOutputWithUnread(channel *entity.Channel, unreadCount, mentionCount int) ChannelOutput {
	return ChannelOutput{
		ID:           channel.ID,
		WorkspaceID:  channel.WorkspaceID,
		Name:         channel.Name,
		Description:  channel.Description,
		IsPrivate:    channel.IsPrivate,
		CreatedBy:    channel.CreatedBy,
		CreatedAt:    channel.CreatedAt,
		UpdatedAt:    channel.UpdatedAt,
		UnreadCount:  unreadCount,
		MentionCount: mentionCount,
		ParentID:     channel.ParentID,
		ArchivedAt:   channel.ArchivedAt,
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
