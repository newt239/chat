package admin

import (
	"context"
	"encoding/csv"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

const (
	// CSV の書き出しは 1 回あたりこの件数までに抑える
	maxExportAuditLogs   = 10000
	defaultAuditLogLimit = 50
)

func (i *Interactor) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (*ListAuditLogsOutput, error) {
	if _, err := i.ensureAdmin(ctx, input.WorkspaceID, input.RequesterID); err != nil {
		return nil, err
	}
	if input.Limit <= 0 {
		input.Limit = defaultAuditLogLimit
	}

	page, err := i.auditLogRepo.List(ctx, input.filter(input.Limit, input.PageToken))
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	outputs, err := i.withActors(ctx, page.Logs)
	if err != nil {
		return nil, err
	}
	return &ListAuditLogsOutput{Logs: outputs, NextPageToken: page.NextPageToken}, nil
}

// ExportAuditLogs は絞り込んだ監査ログを CSV で返し、書き出したこと自体も監査ログに残します
func (i *Interactor) ExportAuditLogs(ctx context.Context, input AuditLogQuery) (*ExportOutput, error) {
	if _, err := i.ensureAdmin(ctx, input.WorkspaceID, input.RequesterID); err != nil {
		return nil, err
	}

	page, err := i.auditLogRepo.List(ctx, input.filter(maxExportAuditLogs, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	outputs, err := i.withActors(ctx, page.Logs)
	if err != nil {
		return nil, err
	}
	content, err := auditLogsToCSV(outputs)
	if err != nil {
		return nil, err
	}

	now := i.now()
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: input.WorkspaceID,
		ActorID:     &input.RequesterID,
		Action:      entity.AuditActionAuditLogExported,
		Metadata:    map[string]string{"count": strconv.Itoa(len(outputs))},
	})
	return &ExportOutput{
		Content:  content,
		FileName: fmt.Sprintf("audit-log-%s-%s.csv", input.WorkspaceID, now.Format("20060102-150405")),
	}, nil
}

func (q AuditLogQuery) filter(limit int, pageToken string) entity.AuditLogFilter {
	return entity.AuditLogFilter{
		WorkspaceID: q.WorkspaceID,
		ActorID:     q.ActorID,
		Actions:     q.Actions,
		Since:       q.Since,
		Until:       q.Until,
		Limit:       limit,
		PageToken:   pageToken,
	}
}

func (i *Interactor) withActors(ctx context.Context, logs []*entity.AuditLog) ([]AuditLogOutput, error) {
	actorIDs := make([]string, 0, len(logs))
	for _, l := range logs {
		if l.ActorID != nil && !slices.Contains(actorIDs, *l.ActorID) {
			actorIDs = append(actorIDs, *l.ActorID)
		}
	}
	users, err := i.userRepo.FindByIDs(ctx, actorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load actors: %w", err)
	}
	actors := make(map[string]*UserSummary, len(users))
	for _, u := range users {
		actors[u.ID] = &UserSummary{ID: u.ID, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL}
	}

	outputs := make([]AuditLogOutput, 0, len(logs))
	for _, l := range logs {
		out := AuditLogOutput{AuditLog: *l}
		if l.ActorID != nil {
			out.Actor = actors[*l.ActorID]
		}
		outputs = append(outputs, out)
	}
	return outputs, nil
}

func auditLogsToCSV(logs []AuditLogOutput) (string, error) {
	var b strings.Builder
	// Excel で文字化けしないよう BOM を付ける
	b.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(&b)
	rows := [][]string{{"日時", "実行者 ID", "実行者", "操作", "対象の種類", "対象 ID", "対象", "詳細", "IP アドレス", "User-Agent"}}
	for _, l := range logs {
		actorID, actorName := "", ""
		if l.ActorID != nil {
			actorID = *l.ActorID
		}
		if l.Actor != nil {
			actorName = l.Actor.DisplayName
		}
		rows = append(rows, []string{
			l.CreatedAt.UTC().Format(time.RFC3339),
			actorID,
			actorName,
			string(l.Action),
			string(l.TargetType),
			l.TargetID,
			l.TargetLabel,
			formatMetadata(l.Metadata),
			l.IPAddress,
			l.UserAgent,
		})
	}
	for _, row := range rows {
		for idx, cell := range row {
			row[idx] = escapeFormula(cell)
		}
	}
	if err := w.WriteAll(rows); err != nil {
		return "", fmt.Errorf("failed to write csv: %w", err)
	}
	return b.String(), nil
}

func formatMetadata(metadata map[string]string) string {
	parts := make([]string, 0, len(metadata))
	for _, k := range slices.Sorted(maps.Keys(metadata)) {
		parts = append(parts, k+"="+metadata[k])
	}
	return strings.Join(parts, " ")
}

// escapeFormula は表計算ソフトで数式として解釈されないよう先頭に ' を付けます
func escapeFormula(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}
