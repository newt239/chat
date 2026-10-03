package repository

import (
	"context"
	stdsql "database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

// parseUUID は形式の誤りを入力の検証エラーとして返します
func parseUUID(id string, label string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, domerr.New(domerr.ErrValidation, "invalid "+label+" format")
	}
	return parsed, nil
}

func parseUUIDs(ids []string, label string) ([]uuid.UUID, error) {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		u, err := parseUUID(id, label)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, u)
	}
	return parsed, nil
}

// parseUUIDPtr は nil と形式の誤りを nil にします
func parseUUIDPtr(id *string) *uuid.UUID {
	if id == nil {
		return nil
	}
	parsed, err := uuid.Parse(*id)
	if err != nil {
		return nil
	}
	return &parsed
}

func optionalString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	return new(id.String())
}

func uuidStrings(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// ignoreConflict は ON CONFLICT DO NOTHING で既存の行と重なったときに返る sql.ErrNoRows を無視します
func ignoreConflict(err error) error {
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil
	}
	return err
}

// orNil は ent の NotFound を見つからなかったこととして nil にします
func orNil[T any](found *T, err error) (*T, error) {
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return found, err
}

// queryUUIDs は 1 列の UUID を返す SQL を実行します
func queryUUIDs(ctx context.Context, client *ent.Client, query string, args ...any) ([]uuid.UUID, error) {
	rows, err := transaction.ResolveClient(ctx, client).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// queryByID は (UUID, 値) の 2 列を返す SQL を実行し、ID ごとの値にします
func queryByID[T any](ctx context.Context, client *ent.Client, query string, args ...any) (map[string]T, error) {
	rows, err := transaction.ResolveClient(ctx, client).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := map[string]T{}
	for rows.Next() {
		var id uuid.UUID
		var value T
		if err := rows.Scan(&id, &value); err != nil {
			return nil, err
		}
		result[id.String()] = value
	}
	return result, rows.Err()
}

func idSet[T any](items []T, id func(T) uuid.UUID) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[id(item).String()] = true
	}
	return set
}

func convertAll[T, U any](items []T, convert func(T) U) []U {
	result := make([]U, 0, len(items))
	for _, item := range items {
		result = append(result, convert(item))
	}
	return result
}

// userToEntity は設定をまだ保存していなければ既定値を使います
func userToEntity(u *ent.User) *entity.User {
	preferences := entity.DefaultPreferences()
	if p := u.Edges.Preference; p != nil {
		preferences = entity.UserPreferences{
			ThemeHue:           p.ThemeHue,
			ThemeChroma:        p.ThemeChroma,
			ThemeSidebar:       entity.SidebarStyle(p.ThemeSidebar),
			ColorMode:          entity.ColorMode(p.ColorMode),
			Locale:             p.Locale,
			NotificationLevel:  entity.NotificationLevel(p.NotificationLevel),
			Timezone:           p.Timezone,
			TimezoneAutoUpdate: p.TimezoneAutoUpdate,
			ChannelSortOrder:   entity.ChannelSortOrder(p.ChannelSortOrder),
			HideJoinMessages:   p.HideJoinMessages,
		}
	}
	return &entity.User{
		ID:           u.ID.String(),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		GoogleSub:    u.GoogleSub,
		DisplayName:  u.DisplayName,
		Bio:          u.Bio,
		AvatarURL:    u.AvatarURL,
		Links:        convertAll(u.Edges.Links, func(l *ent.UserLink) string { return l.URL }),
		IsApp:        u.IsApp,
		IsOfficial:   u.IsOfficial,
		Preferences:  preferences,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func sessionToEntity(s *ent.Session) *entity.Session {
	return &entity.Session{
		ID:               s.ID.String(),
		UserID:           s.UserID.String(),
		RefreshTokenHash: s.RefreshTokenHash,
		ExpiresAt:        s.ExpiresAt,
		RevokedAt:        s.RevokedAt,
		IPAddress:        s.IPAddress,
		UserAgent:        s.UserAgent,
		CreatedAt:        s.CreatedAt,
	}
}

func workspaceToEntity(w *ent.Workspace) *entity.Workspace {
	return &entity.Workspace{
		ID:                 w.ID,
		Name:               w.Name,
		Description:        w.Description,
		IconURL:            w.IconURL,
		IsPublic:           w.IsPublic,
		SignupEnabled:      w.SignupEnabled,
		EmailSignupEnabled: w.EmailSignupEnabled,
		CreatedBy:          w.CreatedByID.String(),
		CreatedAt:          w.CreatedAt,
		UpdatedAt:          w.UpdatedAt,
	}
}

func workspaceMemberToEntity(wm *ent.WorkspaceMember) *entity.WorkspaceMember {
	return &entity.WorkspaceMember{
		WorkspaceID: wm.WorkspaceID,
		UserID:      wm.UserID.String(),
		Role:        entity.WorkspaceRole(wm.Role),
		JoinedAt:    wm.JoinedAt,
		SuspendedAt: wm.SuspendedAt,
	}
}

func channelToEntity(c *ent.Channel) *entity.Channel {
	return &entity.Channel{
		ID:          c.ID.String(),
		WorkspaceID: c.WorkspaceID,
		Name:        c.Name,
		Description: c.Description,
		Type:        entity.ChannelType(c.ChannelType),
		ParentID:    optionalString(c.ParentID),
		CreatedBy:   c.CreatedByID.String(),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func channelMemberToEntity(cm *ent.ChannelMember) *entity.ChannelMember {
	return &entity.ChannelMember{
		ChannelID: cm.ChannelID.String(),
		UserID:    cm.UserID.String(),
		Role:      entity.ChannelRole(cm.Role),
		JoinedAt:  cm.JoinedAt,
	}
}

// messageToEntity は Edges を読み込んでいない場合に nil を渡されても nil を返します
func messageToEntity(m *ent.Message) *entity.Message {
	if m == nil {
		return nil
	}
	return &entity.Message{
		ID:              m.ID.String(),
		ChannelID:       m.ChannelID.String(),
		UserID:          m.UserID.String(),
		ParentID:        optionalString(m.ParentID),
		Body:            m.Body,
		CreatedAt:       m.CreatedAt,
		EditedAt:        m.EditedAt,
		DeletedAt:       m.DeletedAt,
		DeletedBy:       optionalString(m.DeletedBy),
		Location:        locationToEntity(m.LocationLatitude, m.LocationLongitude, m.LocationAccuracy, m.LocationLabel),
		MentionsChannel: m.MentionsChannel,
		MentionsHere:    m.MentionsHere,
	}
}

// locationToEntity は緯度・経度がなければ nil を返します
func locationToEntity(latitude, longitude, accuracy *float64, label *string) *entity.MessageLocation {
	if latitude == nil || longitude == nil {
		return nil
	}
	return &entity.MessageLocation{Latitude: *latitude, Longitude: *longitude, AccuracyMeters: accuracy, Label: label}
}

func attachmentToEntity(a *ent.Attachment) *entity.Attachment {
	var thumbnail *entity.Thumbnail
	if a.ThumbnailStorageKey != nil && a.ThumbnailWidth != nil && a.ThumbnailHeight != nil {
		thumbnail = &entity.Thumbnail{StorageKey: *a.ThumbnailStorageKey, Width: *a.ThumbnailWidth, Height: *a.ThumbnailHeight}
	}
	return &entity.Attachment{
		ID:         a.ID.String(),
		MessageID:  optionalString(a.MessageID),
		UploaderID: a.UploaderID.String(),
		ChannelID:  a.ChannelID.String(),
		FileName:   a.FileName,
		MimeType:   a.MimeType,
		SizeBytes:  a.SizeBytes,
		Media:      entity.MediaMetadata{Width: a.Width, Height: a.Height, DurationSeconds: a.DurationSeconds, Thumbnail: thumbnail},
		StorageKey: a.StorageKey,
		Status:     entity.AttachmentStatus(a.Status),
		CreatedAt:  a.CreatedAt,
	}
}

func userGroupToEntity(ug *ent.UserGroup) *entity.UserGroup {
	return &entity.UserGroup{
		ID:          ug.ID.String(),
		WorkspaceID: ug.WorkspaceID,
		Name:        ug.Name,
		Description: ug.Description,
		CreatedBy:   ug.CreatedByID.String(),
		CreatedAt:   ug.CreatedAt,
		UpdatedAt:   ug.UpdatedAt,
	}
}

// linkPreviewToOGP は YouTube と X の付加情報を読み込んだ link_preview を OGP に変換します
func linkPreviewToOGP(lp *ent.LinkPreview) entity.OGPData {
	if lp == nil {
		return entity.OGPData{}
	}
	ogp := entity.OGPData{
		Title:       lp.Title,
		Description: lp.Description,
		ImageURL:    lp.ImageURL,
		SiteName:    lp.SiteName,
		CardType:    lp.CardType,
		ImageWidth:  lp.ImageWidth,
		ImageHeight: lp.ImageHeight,
	}
	if yt := lp.Edges.Youtube; yt != nil {
		ogp.YouTube = &entity.YouTubeVideo{VideoID: yt.VideoID, ChannelName: yt.ChannelName, DurationSeconds: yt.DurationSeconds}
	}
	if x := lp.Edges.XPost; x != nil {
		ogp.XPost = &entity.XPost{AuthorName: x.AuthorName, AuthorHandle: x.AuthorHandle}
	}
	return ogp
}
