// Package export はワークスペースのデータの書き出しを扱うユースケースです
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit"
)

type Input struct {
	WorkspaceID string
	RequesterID string
}

type Output struct {
	Content  string
	FileName string
}

type exportedMessage struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	UserName  string     `json:"userName"`
	ParentID  *string    `json:"parentId,omitempty"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
	EditedAt  *time.Time `json:"editedAt,omitempty"`
}

type exportedChannel struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	IsPrivate   bool              `json:"isPrivate"`
	Messages    []exportedMessage `json:"messages"`
}

type exportedWorkspace struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	ExportedAt time.Time         `json:"exportedAt"`
	Channels   []exportedChannel `json:"channels"`
}

type Interactor struct {
	workspaceRepo     domainrepository.WorkspaceRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	messageRepo       domainrepository.MessageRepository
	userRepo          domainrepository.UserRepository
	permissionSvc     domainservice.PermissionService
	recorder          audit.Recorder
}

func NewInteractor(
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	permissionSvc domainservice.PermissionService,
	recorder audit.Recorder,
) *Interactor {
	return &Interactor{
		workspaceRepo:     workspaceRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		messageRepo:       messageRepo,
		userRepo:          userRepo,
		permissionSvc:     permissionSvc,
		recorder:          recorder,
	}
}

// ExportMessages は書き出す人が閲覧できるチャンネル（DM を除く）のメッセージを JSON で返します
func (i *Interactor) ExportMessages(ctx context.Context, input Input) (*Output, error) {
	if _, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.RequesterID, entity.PermissionExportData); err != nil {
		return nil, err
	}
	ws, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if ws == nil {
		return nil, domerr.ErrNotFound
	}

	channels, err := i.readableChannels(ctx, input)
	if err != nil {
		return nil, err
	}
	channelIDs := make([]string, 0, len(channels))
	for _, ch := range channels {
		channelIDs = append(channelIDs, ch.ID)
	}
	messages, err := i.messageRepo.FindAllByChannelIDs(ctx, channelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	userNames, err := i.userNames(ctx, messages)
	if err != nil {
		return nil, err
	}

	byChannel := make(map[string][]exportedMessage, len(channels))
	for _, m := range messages {
		byChannel[m.ChannelID] = append(byChannel[m.ChannelID], exportedMessage{
			ID:        m.ID,
			UserID:    m.UserID,
			UserName:  userNames[m.UserID],
			ParentID:  m.ParentID,
			Body:      m.Body,
			CreatedAt: m.CreatedAt,
			EditedAt:  m.EditedAt,
		})
	}

	now := time.Now()
	doc := exportedWorkspace{ID: ws.ID, Name: ws.Name, ExportedAt: now, Channels: make([]exportedChannel, 0, len(channels))}
	for _, ch := range channels {
		msgs := byChannel[ch.ID]
		if msgs == nil {
			msgs = []exportedMessage{}
		}
		doc.Channels = append(doc.Channels, exportedChannel{ID: ch.ID, Name: ch.Name, Description: ch.Description, IsPrivate: ch.IsPrivate, Messages: msgs})
	}
	content, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode export: %w", err)
	}

	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: input.WorkspaceID,
		ActorID:     &input.RequesterID,
		Action:      entity.AuditActionDataExported,
		TargetType:  entity.AuditTargetData,
		TargetLabel: "メッセージ",
		Metadata: map[string]string{
			"kind":     "messages",
			"format":   "json",
			"channels": strconv.Itoa(len(channels)),
			"count":    strconv.Itoa(len(messages)),
		},
	})
	return &Output{
		Content:  string(content),
		FileName: fmt.Sprintf("messages-%s-%s.json", ws.ID, now.Format("20060102-150405")),
	}, nil
}

func (i *Interactor) readableChannels(ctx context.Context, input Input) ([]*entity.Channel, error) {
	channels, err := i.channelRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channels: %w", err)
	}
	result := make([]*entity.Channel, 0, len(channels))
	for _, ch := range channels {
		if ch.Type == entity.ChannelTypeDM || ch.Type == entity.ChannelTypeGroupDM {
			continue
		}
		if ch.IsPrivate {
			isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, input.RequesterID)
			if err != nil {
				return nil, fmt.Errorf("failed to verify channel membership: %w", err)
			}
			if !isMember {
				continue
			}
		}
		result = append(result, ch)
	}
	return result, nil
}

func (i *Interactor) userNames(ctx context.Context, messages []*entity.Message) (map[string]string, error) {
	seen := make(map[string]bool)
	ids := make([]string, 0)
	for _, m := range messages {
		if !seen[m.UserID] {
			seen[m.UserID] = true
			ids = append(ids, m.UserID)
		}
	}
	users, err := i.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}
	names := make(map[string]string, len(users))
	for _, u := range users {
		names[u.ID] = u.DisplayName
	}
	return names, nil
}
