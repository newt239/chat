package channel

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

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
	ErrCannotModifyDM       = errors.New("DM とグループ DM は変更やアーカイブができません")
)

// ChannelRevoker はチャンネルを見られなくなった接続へのリアルタイム配信を止めます
type ChannelRevoker interface {
	RevokeChannel(workspaceID, channelID, userID string)
}

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
	revoker           ChannelRevoker
	logger            domainservice.Logger
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
	revoker ChannelRevoker,
	logger domainservice.Logger,
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
		revoker:           revoker,
		logger:            logger,
	}
}

func (i *channelInteractor) ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error) {
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
		return nil, fmt.Errorf("failed to get unread counts: %w", err)
	}
	mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, joinedIDs, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread mention counts: %w", err)
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
	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, domerr.ErrWorkspaceNotFound
	}

	permission, channelType := entity.PermissionCreatePublicChannel, entity.ChannelTypePublic
	if input.IsPrivate {
		permission, channelType = entity.PermissionCreatePrivateChannel, entity.ChannelTypePrivate
	}
	if _, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.UserID, permission); err != nil {
		return nil, err
	}

	channel, err := entity.NewChannel(entity.ChannelParams{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		Type:        channelType,
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
				Type:        ch.Type,
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
	i.revoker.RevokeChannel(ch.WorkspaceID, ch.ID, "")
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

// ensureCanManage は DM でなく、チャンネルの作成者かワークスペースの管理者であることを確認します
func (i *channelInteractor) ensureCanManage(ctx context.Context, ch *entity.Channel, userID string) error {
	if ch.IsDM() {
		return ErrCannotModifyDM
	}
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
		Metadata:    map[string]string{"private": fmt.Sprint(ch.IsPrivate())},
	})
}

func (i *channelInteractor) UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error) {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return nil, domerr.ErrChannelNotFound
	}
	if err := i.ensureCanManage(ctx, ch, input.UserID); err != nil {
		return nil, err
	}

	original := *ch
	if input.Name != nil {
		if err := ch.ChangeName(*input.Name); err != nil {
			return nil, err
		}
	}
	if input.Description != nil {
		ch.Description = input.Description
	}
	if input.IsPrivate != nil {
		ch.Type = entity.ChannelTypePublic
		if *input.IsPrivate {
			ch.Type = entity.ChannelTypePrivate
		}
	}
	ch.UpdatedAt = time.Now().UTC()

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if ch.Name != original.Name {
			if err := i.renameDescendants(txCtx, ch, original.Name); err != nil {
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

	if ch.Name != original.Name {
		i.recordSystemMessage(ctx, ch.ID, input.UserID, entity.SystemMessageKindChannelNameChanged, original.Name, ch.Name)
	}
	if from, to := derefString(original.Description), derefString(ch.Description); from != to {
		i.recordSystemMessage(ctx, ch.ID, input.UserID, entity.SystemMessageKindChannelDescriptionChanged, from, to)
	}
	if original.Type != ch.Type {
		i.recordSystemMessage(ctx, ch.ID, input.UserID, entity.SystemMessageKindChannelPrivacyChanged, string(original.Type), string(ch.Type))
		if ch.IsPrivate() {
			// 参加していないユーザーは見られなくなる
			i.revoker.RevokeChannel(ch.WorkspaceID, ch.ID, "")
		}
	}

	out := toChannelOutput(ch)
	return &out, nil
}

// recordSystemMessage はチャンネルの変更をタイムラインに残します。失敗しても変更は取り消さない
func (i *channelInteractor) recordSystemMessage(ctx context.Context, channelID, actorID string, kind entity.SystemMessageKind, from, to string) {
	if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
		ChannelID: channelID,
		Kind:      kind,
		Payload:   map[string]any{"from": from, "to": to},
		ActorID:   &actorID,
	}); err != nil {
		i.logger.Warn("チャンネルの変更をタイムラインに残せません", domainservice.LogField{Key: "channelId", Value: channelID}, domainservice.LogField{Key: "error", Value: err.Error()})
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
		IsPrivate:    channel.IsPrivate(),
		CreatedBy:    channel.CreatedBy,
		CreatedAt:    channel.CreatedAt,
		UpdatedAt:    channel.UpdatedAt,
		UnreadCount:  unreadCount,
		MentionCount: mentionCount,
		ParentID:     channel.ParentID,
		ArchivedAt:   channel.ArchivedAt,
	}
}
