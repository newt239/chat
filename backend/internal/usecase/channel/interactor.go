package channel

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/usecase/audit"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrChannelNameExists    = domerr.New(domerr.ErrAlreadyExists, "同じ名前のチャンネルがすでに存在します")
	ErrMemberNotInWorkspace = domerr.New(domerr.ErrValidation, "ワークスペースのメンバーではないユーザーが含まれています")
	ErrCannotModifyDM       = domerr.New(domerr.ErrFailedPrecondition, "DM とグループ DM は変更できません")
)

// ChannelRevoker はチャンネルを見られなくなった接続へのリアルタイム配信を止めます
type ChannelRevoker interface {
	RevokeChannel(workspaceID, channelID, userID string)
}

type Interactor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	channelStarRepo   domainrepository.ChannelStarRepository
	channelMuteRepo   domainrepository.ChannelMuteRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	readStateRepo     domainrepository.ReadStateRepository
	txManager         domaintransaction.Manager
	systemMessages    *messageuc.SystemMessages
	channelAccessSvc  domainservice.ChannelAccessService
	permissionSvc     domainservice.PermissionService
	recorder          audit.Recorder
	revoker           ChannelRevoker
}

func New(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	channelStarRepo domainrepository.ChannelStarRepository,
	channelMuteRepo domainrepository.ChannelMuteRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	readStateRepo domainrepository.ReadStateRepository,
	txManager domaintransaction.Manager,
	systemMessages *messageuc.SystemMessages,
	channelAccessSvc domainservice.ChannelAccessService,
	permissionSvc domainservice.PermissionService,
	recorder audit.Recorder,
	revoker ChannelRevoker,
) *Interactor {
	return &Interactor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		channelStarRepo:   channelStarRepo,
		channelMuteRepo:   channelMuteRepo,
		workspaceRepo:     workspaceRepo,
		readStateRepo:     readStateRepo,
		txManager:         txManager,
		systemMessages:    systemMessages,
		channelAccessSvc:  channelAccessSvc,
		permissionSvc:     permissionSvc,
		recorder:          recorder,
		revoker:           revoker,
	}
}

// ListChannels は参加中のチャンネルと、ツリーを組み立てるための未参加の祖先チャンネルを名前順に返します
func (i *Interactor) ListChannels(ctx context.Context, workspaceID, userID string) ([]ChannelOutput, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, userID); err != nil {
		return nil, err
	}
	joined, err := i.channelRepo.FindAccessibleChannels(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}
	ancestors, err := i.findMissingAncestors(ctx, workspaceID, userID, joined)
	if err != nil {
		return nil, err
	}

	joinedIDs := channelIDs(joined)
	unreadCounts, err := i.readStateRepo.GetUnreadCountBatch(ctx, joinedIDs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread counts: %w", err)
	}
	mentionCounts, err := i.readStateRepo.GetUnreadMentionCountBatch(ctx, joinedIDs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread mention counts: %w", err)
	}
	all := slices.Concat(joined, ancestors)
	allIDs := channelIDs(all)
	starred, err := i.channelStarRepo.FindStarredChannelIDs(ctx, userID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch starred channels: %w", err)
	}
	muted, err := i.channelMuteRepo.FindMutedChannelIDs(ctx, userID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch muted channels: %w", err)
	}
	lastMessageAt, err := i.channelRepo.FindLastMessageAtBatch(ctx, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch last message times: %w", err)
	}

	output := make([]ChannelOutput, 0, len(all))
	for idx, ch := range all {
		out := ChannelOutput{
			Channel:      ch,
			IsMember:     idx < len(joined),
			UnreadCount:  unreadCounts[ch.ID],
			MentionCount: mentionCounts[ch.ID],
			IsStarred:    starred[ch.ID],
			IsMuted:      muted[ch.ID],
		}
		if at, ok := lastMessageAt[ch.ID]; ok {
			out.LastMessageAt = &at
		}
		output = append(output, out)
	}
	slices.SortFunc(output, func(a, b ChannelOutput) int { return strings.Compare(a.Name, b.Name) })
	return output, nil
}

func (i *Interactor) ListBrowsableChannels(ctx context.Context, workspaceID, userID string) ([]BrowsableChannelOutput, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, userID); err != nil {
		return nil, err
	}
	channels, err := i.channelRepo.FindBrowsableChannels(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}
	return i.toBrowsableOutputs(ctx, userID, channels)
}

func (i *Interactor) SearchBrowsableChannels(ctx context.Context, input SearchBrowsableChannelsInput) (*SearchBrowsableChannelsOutput, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
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
	outputs, err := i.toBrowsableOutputs(ctx, input.UserID, channels)
	if err != nil {
		return nil, err
	}
	return &SearchBrowsableChannelsOutput{Channels: outputs, Total: total}, nil
}

// toBrowsableOutputs はメンバー数と自分が参加しているかを付けます
func (i *Interactor) toBrowsableOutputs(ctx context.Context, userID string, channels []*entity.Channel) ([]BrowsableChannelOutput, error) {
	memberCounts, err := i.channelRepo.CountMembersBatch(ctx, channelIDs(channels))
	if err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}
	joined, err := i.channelMemberRepo.FindJoinedChannelIDs(ctx, userID, channelIDs(channels))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch joined channels: %w", err)
	}

	output := make([]BrowsableChannelOutput, 0, len(channels))
	for _, ch := range channels {
		output = append(output, BrowsableChannelOutput{Channel: ChannelOutput{Channel: ch, IsMember: joined[ch.ID]}, MemberCount: memberCounts[ch.ID]})
	}
	return output, nil
}

// findMissingAncestors はツリーを組み立てるため、参加していないが閲覧できる祖先チャンネルを返します
func (i *Interactor) findMissingAncestors(ctx context.Context, workspaceID, userID string, joined []*entity.Channel) ([]*entity.Channel, error) {
	joinedNames := make(map[string]bool, len(joined))
	for _, ch := range joined {
		joinedNames[ch.Name] = true
	}
	var missing []string
	for _, ch := range joined {
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

func (i *Interactor) CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error) {
	permission, channelType := entity.PermissionCreatePublicChannel, entity.ChannelTypePublic
	if input.IsPrivate {
		permission, channelType = entity.PermissionCreatePrivateChannel, entity.ChannelTypePrivate
	}
	if _, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.UserID, permission); err != nil {
		return nil, err
	}
	channel, err := entity.NewChannel(entity.Channel{
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
		if !slices.Contains(memberIDs, id) {
			memberIDs = append(memberIDs, id)
		}
	}
	active, err := i.workspaceRepo.FindActiveMemberIDs(ctx, input.WorkspaceID, memberIDs[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if len(active) != len(memberIDs)-1 {
		return nil, ErrMemberNotInWorkspace
	}

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if channel.ParentID, err = i.ensureAncestors(txCtx, channel); err != nil {
			return err
		}
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

	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: channel.WorkspaceID,
		ActorID:     &input.UserID,
		Action:      entity.AuditActionChannelCreated,
		TargetType:  entity.AuditTargetChannel,
		TargetID:    channel.ID,
		TargetLabel: channel.Name,
		Metadata:    map[string]string{"private": fmt.Sprint(channel.IsPrivate())},
	})
	return &ChannelOutput{Channel: channel, IsMember: true}, nil
}

// ensureAncestors は存在しない祖先チャンネルを子と同じ公開範囲で作って作成者を参加させ、直近の親の ID を返します
func (i *Interactor) ensureAncestors(ctx context.Context, ch *entity.Channel) (*string, error) {
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
			parent = &entity.Channel{WorkspaceID: ch.WorkspaceID, Name: path, Type: ch.Type, ParentID: parentID, CreatedBy: ch.CreatedBy}
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

func (i *Interactor) addMember(ctx context.Context, channelID, userID string, role entity.ChannelRole) error {
	if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: channelID, UserID: userID, Role: role}); err != nil {
		return fmt.Errorf("failed to add channel member: %w", err)
	}
	return nil
}

func (i *Interactor) GetChannel(ctx context.Context, channelID, userID string) (*ChannelOutput, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}
	isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify channel membership: %w", err)
	}
	starred, err := i.channelStarRepo.FindStarredChannelIDs(ctx, userID, []string{ch.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch starred channels: %w", err)
	}
	muted, err := i.channelMuteRepo.FindMutedChannelIDs(ctx, userID, []string{ch.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch muted channels: %w", err)
	}
	return &ChannelOutput{Channel: ch, IsMember: isMember, IsStarred: starred[ch.ID], IsMuted: muted[ch.ID]}, nil
}

func (i *Interactor) SetChannelStarred(ctx context.Context, input SetFlagInput) error {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}
	if err := i.channelStarRepo.SetStarred(ctx, input.UserID, input.ChannelID, input.Value); err != nil {
		return fmt.Errorf("failed to update star: %w", err)
	}
	return nil
}

func (i *Interactor) SetChannelMuted(ctx context.Context, input SetFlagInput) error {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}
	if err := i.channelMuteRepo.SetMuted(ctx, input.UserID, input.ChannelID, input.Value); err != nil {
		return fmt.Errorf("failed to update mute: %w", err)
	}
	return nil
}

// UpdateChannel は DM でないチャンネルを、作成者かワークスペースの管理者だけが変更できます
func (i *Interactor) UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error) {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return nil, domerr.ErrChannelNotFound
	}
	if ch.IsDM() {
		return nil, ErrCannotModifyDM
	}
	member, err := domainservice.EnsureMember(ctx, i.workspaceRepo, ch.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}
	if ch.CreatedBy != input.UserID && !member.IsAdmin() {
		return nil, domerr.ErrUnauthorized
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

	record := func(kind entity.SystemMessageKind, from, to string) {
		i.systemMessages.Record(ctx, ch, kind, input.UserID, map[string]any{"from": from, "to": to})
	}
	if ch.Name != original.Name {
		record(entity.SystemMessageKindChannelNameChanged, original.Name, ch.Name)
	}
	if from, to := derefString(original.Description), derefString(ch.Description); from != to {
		record(entity.SystemMessageKindChannelDescriptionChanged, from, to)
	}
	if original.Type != ch.Type {
		record(entity.SystemMessageKindChannelPrivacyChanged, string(original.Type), string(ch.Type))
		if ch.IsPrivate() {
			// 参加していないユーザーは見られなくなる
			i.revoker.RevokeChannel(ch.WorkspaceID, ch.ID, "")
		}
	}
	return &ChannelOutput{Channel: ch}, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// renameDescendants は名前を変えたチャンネルの子孫のパスを付け替えます
func (i *Interactor) renameDescendants(ctx context.Context, ch *entity.Channel, originalName string) error {
	conflicts, err := i.channelRepo.FindByNames(ctx, ch.WorkspaceID, []string{ch.Name})
	if err != nil {
		return fmt.Errorf("failed to check channel name: %w", err)
	}
	if slices.ContainsFunc(conflicts, func(c *entity.Channel) bool { return c.ID != ch.ID }) {
		return ErrChannelNameExists
	}
	if err := i.channelRepo.RenameDescendants(ctx, ch.WorkspaceID, originalName, ch.Name); err != nil {
		return fmt.Errorf("failed to rename descendant channels: %w", err)
	}
	return nil
}

func channelIDs(channels []*entity.Channel) []string {
	ids := make([]string, len(channels))
	for idx, ch := range channels {
		ids[idx] = ch.ID
	}
	return ids
}
