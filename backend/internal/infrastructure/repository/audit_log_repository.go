package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/auditlog"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type auditLogRepository struct {
	client *ent.Client
}

func NewAuditLogRepository(client *ent.Client) domainrepository.AuditLogRepository {
	return &auditLogRepository{client: client}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	client := transaction.ResolveClient(ctx, r.client)
	builder := client.AuditLog.Create().
		SetWorkspaceID(log.WorkspaceID).
		SetAction(string(log.Action)).
		SetTargetType(string(log.TargetType)).
		SetTargetID(log.TargetID).
		SetTargetLabel(log.TargetLabel).
		SetMetadata(log.Metadata).
		SetIPAddress(log.IPAddress).
		SetUserAgent(log.UserAgent)
	if log.ActorID != nil {
		actorID, err := utils.ParseUUID(*log.ActorID, "actor ID")
		if err != nil {
			return err
		}
		builder = builder.SetActorID(actorID)
	}
	if !log.CreatedAt.IsZero() {
		builder = builder.SetCreatedAt(log.CreatedAt)
	}

	saved, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*log = *auditLogToEntity(saved)
	return nil
}

func (r *auditLogRepository) List(ctx context.Context, filter entity.AuditLogFilter) ([]*entity.AuditLog, int, error) {
	client := transaction.ResolveClient(ctx, r.client)
	query := client.AuditLog.Query().Where(auditlog.WorkspaceID(filter.WorkspaceID))
	if filter.ActorID != nil {
		actorID, err := utils.ParseUUID(*filter.ActorID, "actor ID")
		if err != nil {
			return nil, 0, err
		}
		query = query.Where(auditlog.ActorID(actorID))
	}
	if len(filter.Actions) > 0 {
		actions := make([]string, 0, len(filter.Actions))
		for _, a := range filter.Actions {
			actions = append(actions, string(a))
		}
		query = query.Where(auditlog.ActionIn(actions...))
	}
	if filter.Since != nil {
		query = query.Where(auditlog.CreatedAtGTE(*filter.Since))
	}
	if filter.Until != nil {
		query = query.Where(auditlog.CreatedAtLT(*filter.Until))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	logs, err := query.
		Order(ent.Desc(auditlog.FieldCreatedAt), ent.Desc(auditlog.FieldID)).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*entity.AuditLog, 0, len(logs))
	for _, l := range logs {
		result = append(result, auditLogToEntity(l))
	}
	return result, total, nil
}

func auditLogToEntity(l *ent.AuditLog) *entity.AuditLog {
	return &entity.AuditLog{
		ID:          l.ID.String(),
		WorkspaceID: l.WorkspaceID,
		ActorID:     utils.UUIDPtrToStringPtr(l.ActorID),
		Action:      entity.AuditAction(l.Action),
		TargetType:  entity.AuditTargetType(l.TargetType),
		TargetID:    l.TargetID,
		TargetLabel: l.TargetLabel,
		Metadata:    l.Metadata,
		IPAddress:   l.IPAddress,
		UserAgent:   l.UserAgent,
		CreatedAt:   l.CreatedAt,
	}
}
