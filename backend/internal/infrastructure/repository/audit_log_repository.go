package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/auditlog"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type auditLogRepository struct {
	client *ent.Client
}

func NewAuditLogRepository(client *ent.Client) domainrepository.AuditLogRepository {
	return &auditLogRepository{client: client}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	saved, err := transaction.ResolveClient(ctx, r.client).AuditLog.Create().
		SetWorkspaceID(log.WorkspaceID).
		SetNillableActorID(parseUUIDPtr(log.ActorID)).
		SetAction(string(log.Action)).
		SetTargetType(string(log.TargetType)).
		SetTargetID(log.TargetID).
		SetTargetLabel(log.TargetLabel).
		SetMetadata(log.Metadata).
		SetIPAddress(log.IPAddress).
		SetUserAgent(log.UserAgent).
		Save(ctx)
	if err != nil {
		return err
	}
	*log = *auditLogToEntity(saved)
	return nil
}

func (r *auditLogRepository) List(ctx context.Context, filter entity.AuditLogFilter, limit, offset int) ([]*entity.AuditLog, error) {
	query := transaction.ResolveClient(ctx, r.client).AuditLog.Query().Where(auditlog.WorkspaceID(filter.WorkspaceID))
	if actorID := parseUUIDPtr(filter.ActorID); actorID != nil {
		query.Where(auditlog.ActorID(*actorID))
	}
	if len(filter.Actions) > 0 {
		query.Where(auditlog.ActionIn(convertAll(filter.Actions, func(a entity.AuditAction) string { return string(a) })...))
	}
	if filter.Since != nil {
		query.Where(auditlog.CreatedAtGTE(*filter.Since))
	}
	if filter.Until != nil {
		query.Where(auditlog.CreatedAtLT(*filter.Until))
	}
	logs, err := query.
		Order(ent.Desc(auditlog.FieldCreatedAt), ent.Desc(auditlog.FieldID)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(logs, auditLogToEntity), nil
}

func auditLogToEntity(l *ent.AuditLog) *entity.AuditLog {
	return &entity.AuditLog{
		ID:          l.ID.String(),
		WorkspaceID: l.WorkspaceID,
		ActorID:     optionalString(l.ActorID),
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
