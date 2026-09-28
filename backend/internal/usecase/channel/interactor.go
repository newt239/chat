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
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

var (
	ErrUnauthorized         = errors.New("この操作を行う権限がありません")
	ErrWorkspaceNotFound    = errors.New("ワークスペースが見つかりません")
	ErrChannelNotFound      = errors.New("チャンネルが見つかりません")
	ErrChannelNameExists    = errors.New("同じ名前のチャンネルがすでに存在します")
	ErrChannelHasChildren   = errors.New("下の階層にチャンネルがあるため削除できません")
	ErrMemberNotInWorkspace = errors.New("ワークスペースのメンバーではないユーザーが含まれています")
)

type ChannelUseCase interface {
	GetChannel(ctx context.Context, input GetChannelInput) (*ChannelOutput, error)
	DeleteChannel(ctx context.Context, input DeleteChannelInput) error
	ListChannels(ctx context.Context, input ListChannelsInput) ([]ChannelOutput, error)
	CreateChannel(ctx context.Context, input CreateChannelInput) (*ChannelOutput, error)
	UpdateChannel(ctx context.Context, input UpdateChannelInput) (*ChannelOutput, error)
	SetChannelStarred(ctx context.Context, input SetChannelStarredInput) error
}

type channelInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	channelStarRepo   domainrepository.ChannelStarRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	readStateRepo     domainrepository.ReadStateRepository
	txManager         domaintransaction.Manager
	systemMessageUC   systemmessage.UseCase
	channelAccessSvc  domainservice.ChannelAccessService
}

func NewChannelInteractor(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	channelStarRepo domainrepository.ChannelStarRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	readStateRepo domainrepository.ReadStateRepository,
	txManager domaintransaction.Manager,
	systemMessageUC systemmessage.UseCase,
	channelAccessSvc domainservice.ChannelAccessService,
) ChannelUseCase {
	return &channelInteractor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		channelStarRepo:   channelStarRepo,
		workspaceRepo:     workspaceRepo,
		readStateRepo:     readStateRepo,
		txManager:         txManager,
		systemMessageUC:   systemMessageUC,
		channelAccessSvc:  channelAccessSvc,
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

	output := make([]ChannelOutput, 0, len(all))
	for _, ch := range joined {
		out := toChannelOutputWithUnread(ch, unreadCounts[ch.ID], mentionCounts[ch.ID] > 0)
		out.IsMember = true
		out.IsStarred = starred[ch.ID]
		output = append(output, out)
	}
	for _, ch := range ancestors {
		out := toChannelOutput(ch)
		out.IsStarred = starred[ch.ID]
		output = append(output, out)
	}
	slices.SortFunc(output, func(a, b ChannelOutput) int { return strings.Compare(a.Name, b.Name) })

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

// ensureCanCreateChannel は権限設定 (#16) の導入時に差し替える前提で作成権限の判定をまとめています
func ensureCanCreateChannel(member *entity.WorkspaceMember, isPrivate bool) error {
	if member == nil {
		return ErrUnauthorized
	}
	// 非公開チャンネルは 10 人を超える DM の代わりにも使うため、ゲスト以外なら作成できる
	if isPrivate && member.Role != entity.WorkspaceRoleGuest {
		return nil
	}
	if !member.CanCreateChannel() {
		return ErrUnauthorized
	}
	return nil
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

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if err := ensureCanCreateChannel(member, input.IsPrivate); err != nil {
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
				return nil, ErrUnauthorized
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

	output := toChannelOutput(ch)
	output.IsMember = isMember
	output.IsStarred = starred[ch.ID]
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

func (i *channelInteractor) DeleteChannel(ctx context.Context, input DeleteChannelInput) error {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to fetch channel: %w", err)
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	// 作成者かワークスペースの管理者のみ削除できる
	if ch.CreatedBy != input.UserID {
		member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, input.UserID)
		if err != nil {
			return fmt.Errorf("failed to verify membership: %w", err)
		}
		if member == nil || !member.CanCreateChannel() {
			return ErrUnauthorized
		}
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
	return nil
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
	if err != nil || wsMember == nil || !wsMember.CanCreateChannel() {
		return nil, ErrUnauthorized
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
		ParentID:    channel.ParentID,
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
