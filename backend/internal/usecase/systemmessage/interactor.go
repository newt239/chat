package systemmessage

import (
    "context"
    "fmt"
    "time"

    "github.com/newt239/chat/internal/domain/entity"
    domainrepository "github.com/newt239/chat/internal/domain/repository"
    domainservice "github.com/newt239/chat/internal/domain/service"
)

type CreateInput struct {
    ChannelID string
    Kind      entity.SystemMessageKind
    Payload   map[string]any
    ActorID   *string
}

type UseCase interface {
    Create(ctx context.Context, input CreateInput) (*entity.SystemMessage, error)
}

type interactor struct {
    systemMsgRepo domainrepository.SystemMessageRepository
    channelRepo   domainrepository.ChannelRepository
    notification  domainservice.NotificationService
}

func New(systemMsgRepo domainrepository.SystemMessageRepository, channelRepo domainrepository.ChannelRepository, notification domainservice.NotificationService) UseCase {
    return &interactor{
        systemMsgRepo: systemMsgRepo,
        channelRepo:   channelRepo,
        notification:  notification,
    }
}

func (i *interactor) Create(ctx context.Context, input CreateInput) (*entity.SystemMessage, error) {
    if input.ChannelID == "" {
        return nil, fmt.Errorf("channel id is required")
    }
    if input.Kind == "" {
        return nil, fmt.Errorf("kind is required")
    }

    msg := &entity.SystemMessage{
        ChannelID: input.ChannelID,
        Kind:      input.Kind,
        Payload:   input.Payload,
        ActorID:   input.ActorID,
        CreatedAt: time.Now(),
    }

    if err := i.systemMsgRepo.Create(ctx, msg); err != nil {
        return nil, err
    }

    // 通知（workspaceID はチャネルから解決）
    ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
    if err == nil && ch != nil && i.notification != nil {
        i.notification.NotifySystemMessageCreated(ch.WorkspaceID, input.ChannelID, map[string]any{
            "id":        msg.ID,
            "channelId": msg.ChannelID,
            "kind":      string(msg.Kind),
            "payload":   msg.Payload,
            "actorId":   msg.ActorID,
            "createdAt": msg.CreatedAt.Format(time.RFC3339),
        })
    }

    return msg, nil
}


