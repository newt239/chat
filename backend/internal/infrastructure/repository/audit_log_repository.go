package repository

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/auditlog"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type auditLogRepository struct {
	client *ent.Client
}

func NewAuditLogRepository(client *ent.Client) domainrepository.AuditLogRepository {
	return &auditLogRepository{client: client}
}

// Create は呼び出し元のトランザクションに参加しない（保存先を別の DB に移しても振る舞いを変えないため）
func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	builder := r.client.AuditLog.Create().
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

func (r *auditLogRepository) List(ctx context.Context, filter entity.AuditLogFilter) (*entity.AuditLogPage, error) {
	query := r.client.AuditLog.Query().Where(auditlog.WorkspaceID(filter.WorkspaceID))
	if filter.ActorID != nil {
		actorID, err := utils.ParseUUID(*filter.ActorID, "actor ID")
		if err != nil {
			return nil, err
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
	if filter.PageToken != "" {
		createdAt, id, err := decodeAuditLogPageToken(filter.PageToken)
		if err != nil {
			return nil, err
		}
		query = query.Where(auditlog.Or(
			auditlog.CreatedAtLT(createdAt),
			auditlog.And(auditlog.CreatedAt(createdAt), auditlog.IDLT(id)),
		))
	}

	// 続きがあるかを知るため 1 件多く取る
	logs, err := query.
		Order(ent.Desc(auditlog.FieldCreatedAt), ent.Desc(auditlog.FieldID)).
		Limit(filter.Limit + 1).
		All(ctx)
	if err != nil {
		return nil, err
	}
	page := &entity.AuditLogPage{}
	if len(logs) > filter.Limit {
		logs = logs[:filter.Limit]
		last := logs[len(logs)-1]
		page.NextPageToken = encodeAuditLogPageToken(last.CreatedAt, last.ID)
	}
	page.Logs = make([]*entity.AuditLog, 0, len(logs))
	for _, l := range logs {
		page.Logs = append(page.Logs, auditLogToEntity(l))
	}
	return page, nil
}

// ページトークンは最後の行の (created_at, id) で、同時刻の行も取りこぼさないようにする
func encodeAuditLogPageToken(createdAt time.Time, id uuid.UUID) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(createdAt.UnixMicro(), 10) + "_" + id.String()))
}

func decodeAuditLogPageToken(token string) (time.Time, uuid.UUID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, uuid.Nil, entity.ErrInvalidAuditLogPageToken
	}
	micros, idText, ok := strings.Cut(string(raw), "_")
	if !ok {
		return time.Time{}, uuid.Nil, entity.ErrInvalidAuditLogPageToken
	}
	unix, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return time.Time{}, uuid.Nil, entity.ErrInvalidAuditLogPageToken
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		return time.Time{}, uuid.Nil, entity.ErrInvalidAuditLogPageToken
	}
	return time.UnixMicro(unix), id, nil
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
