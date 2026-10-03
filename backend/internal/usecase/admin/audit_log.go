package admin

import (
	"cmp"
	"context"
	"encoding/csv"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

var errInvalidPageToken = domerr.New(domerr.ErrValidation, "ページトークンが不正です")

const (
	// CSV の書き出しは 1 回あたりこの件数までに抑える
	maxExportAuditLogs   = 10000
	defaultAuditLogLimit = 50
)

// ListAuditLogs はページトークンに次のページの先頭の位置を入れて返します
func (i *Interactor) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (*ListAuditLogsOutput, error) {
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.WorkspaceID, input.RequesterID); err != nil {
		return nil, err
	}
	limit := cmp.Or(input.Limit, defaultAuditLogLimit)
	offset := 0
	if input.PageToken != "" {
		var err error
		if offset, err = strconv.Atoi(input.PageToken); err != nil || offset < 0 {
			return nil, errInvalidPageToken
		}
	}
	// 続きがあるかを知るため 1 件多く取る
	logs, err := i.auditLogRepo.List(ctx, input.filter(), limit+1, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	output := &ListAuditLogsOutput{}
	if len(logs) > limit {
		logs = logs[:limit]
		output.NextPageToken = strconv.Itoa(offset + limit)
	}
	if output.Logs, err = i.withActors(ctx, logs); err != nil {
		return nil, err
	}
	return output, nil
}

// ExportAuditLogs は絞り込んだ監査ログを CSV で返し、書き出したこと自体も監査ログに残します
func (i *Interactor) ExportAuditLogs(ctx context.Context, input AuditLogQuery) (*ExportOutput, error) {
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.WorkspaceID, input.RequesterID); err != nil {
		return nil, err
	}
	logs, err := i.auditLogRepo.List(ctx, input.filter(), maxExportAuditLogs, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	outputs, err := i.withActors(ctx, logs)
	if err != nil {
		return nil, err
	}
	content, err := auditLogsToCSV(outputs)
	if err != nil {
		return nil, err
	}

	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: input.WorkspaceID,
		ActorID:     &input.RequesterID,
		Action:      entity.AuditActionAuditLogExported,
		Metadata:    map[string]string{"count": strconv.Itoa(len(outputs))},
	})
	return &ExportOutput{
		Content:  content,
		FileName: fmt.Sprintf("audit-log-%s-%s.csv", input.WorkspaceID, time.Now().Format("20060102-150405")),
	}, nil
}

func (q AuditLogQuery) filter() entity.AuditLogFilter {
	return entity.AuditLogFilter{WorkspaceID: q.WorkspaceID, ActorID: q.ActorID, Actions: q.Actions, Since: q.Since, Until: q.Until}
}

func (i *Interactor) withActors(ctx context.Context, logs []*entity.AuditLog) ([]AuditLogOutput, error) {
	actorIDs := make([]string, 0, len(logs))
	for _, l := range logs {
		if l.ActorID != nil {
			actorIDs = append(actorIDs, *l.ActorID)
		}
	}
	actors, err := i.userRepo.FindByIDs(ctx, actorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load actors: %w", err)
	}
	outputs := make([]AuditLogOutput, 0, len(logs))
	for _, l := range logs {
		out := AuditLogOutput{AuditLog: *l}
		if l.ActorID != nil {
			if u := actors[*l.ActorID]; u != nil {
				out.Actor = new(message.NewUserInfo(u))
			}
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
