package utils

import (
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/entity"
)

// User converters
func UserToEntity(u *ent.User) *entity.User {
	if u == nil {
		return nil
	}
	links := make([]string, 0, len(u.Edges.Links))
	for _, l := range u.Edges.Links {
		links = append(links, l.URL)
	}
	return &entity.User{
		ID:           u.ID.String(),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		GoogleSub:    u.GoogleSub,
		DisplayName:  u.DisplayName,
		Bio:          StringPtrFromNullable(u.Bio),
		AvatarURL:    StringPtrFromNullable(u.AvatarURL),
		Links:        links,
		IsApp:        u.IsApp,
		IsOfficial:   u.IsOfficial,
		Preferences:  UserPreferenceToEntity(u.Edges.Preference),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// UserPreferenceToEntity は設定を変換します。まだ保存していなければ既定値を返します
func UserPreferenceToEntity(p *ent.UserPreference) entity.UserPreferences {
	if p == nil {
		return entity.DefaultPreferences()
	}
	return entity.UserPreferences{
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

// Session converters
func SessionToEntity(s *ent.Session) *entity.Session {
	if s == nil {
		return nil
	}
	var revokedAt *time.Time
	if !s.RevokedAt.IsZero() {
		revokedAt = &s.RevokedAt
	}
	return &entity.Session{
		ID:               s.ID.String(),
		UserID:           s.UserID.String(),
		RefreshTokenHash: s.RefreshTokenHash,
		ExpiresAt:        s.ExpiresAt,
		RevokedAt:        revokedAt,
		IPAddress:        s.IPAddress,
		UserAgent:        s.UserAgent,
		CreatedAt:        s.CreatedAt,
	}
}

// Workspace converters
func WorkspaceToEntity(w *ent.Workspace) *entity.Workspace {
	if w == nil {
		return nil
	}
	return &entity.Workspace{
		ID:                 w.ID,
		Name:               w.Name,
		Description:        StringPtrFromNullable(w.Description),
		IconURL:            StringPtrFromNullable(w.IconURL),
		IsPublic:           w.IsPublic,
		SignupEnabled:      w.SignupEnabled,
		EmailSignupEnabled: w.EmailSignupEnabled,
		CreatedBy:          w.CreatedByID.String(),
		CreatedAt:          w.CreatedAt,
		UpdatedAt:          w.UpdatedAt,
	}
}

// WorkspaceMember converters
func WorkspaceMemberToEntity(wm *ent.WorkspaceMember) *entity.WorkspaceMember {
	if wm == nil {
		return nil
	}
	return &entity.WorkspaceMember{
		WorkspaceID: wm.WorkspaceID,
		UserID:      wm.UserID.String(),
		Role:        entity.WorkspaceRole(wm.Role),
		JoinedAt:    wm.JoinedAt,
		SuspendedAt: wm.SuspendedAt,
	}
}

// Channel converters
func ChannelToEntity(c *ent.Channel) *entity.Channel {
	if c == nil {
		return nil
	}

	channelType := entity.ChannelTypePublic
	if c.ChannelType != "" {
		channelType = entity.ChannelType(c.ChannelType)
	}

	return &entity.Channel{
		ID:          c.ID.String(),
		WorkspaceID: c.WorkspaceID,
		Name:        c.Name,
		Description: StringPtrFromNullable(c.Description),
		Type:        channelType,
		ParentID:    UUIDPtrToStringPtr(c.ParentID),
		CreatedBy:   c.CreatedByID.String(),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		ArchivedAt:  c.ArchivedAt,
	}
}

// ChannelMember converters
func ChannelMemberToEntity(cm *ent.ChannelMember) *entity.ChannelMember {
	if cm == nil {
		return nil
	}
	return &entity.ChannelMember{
		ChannelID: cm.ChannelID.String(),
		UserID:    cm.UserID.String(),
		Role:      entity.ChannelRole(cm.Role),
		JoinedAt:  cm.JoinedAt,
	}
}

// Message converters
func MessageToEntity(m *ent.Message) *entity.Message {
	if m == nil {
		return nil
	}

	var parentID *string
	if m.ParentID != nil {
		pid := m.ParentID.String()
		parentID = &pid
	}

	var deletedBy *string
	if m.DeletedBy != uuid.Nil {
		db := m.DeletedBy.String()
		deletedBy = &db
	}

	var editedAt *time.Time
	if !m.EditedAt.IsZero() {
		editedAt = &m.EditedAt
	}

	var deletedAt *time.Time
	if !m.DeletedAt.IsZero() {
		deletedAt = &m.DeletedAt
	}

	return &entity.Message{
		ID:        m.ID.String(),
		ChannelID: m.ChannelID.String(),
		UserID:    m.UserID.String(),
		ParentID:  parentID,
		Body:      m.Body,
		CreatedAt: m.CreatedAt,
		EditedAt:  editedAt,
		DeletedAt: deletedAt,
		DeletedBy: deletedBy,

		Location:        LocationToEntity(m.LocationLatitude, m.LocationLongitude, m.LocationAccuracy, m.LocationLabel),
		MentionsChannel: m.MentionsChannel,
		MentionsHere:    m.MentionsHere,
	}
}

// LocationToEntity は位置情報の列をまとめます。緯度・経度がなければ nil を返します
func LocationToEntity(latitude, longitude, accuracy *float64, label *string) *entity.MessageLocation {
	if latitude == nil || longitude == nil {
		return nil
	}
	return &entity.MessageLocation{Latitude: *latitude, Longitude: *longitude, AccuracyMeters: accuracy, Label: label}
}

// MessageReaction converters
func MessageReactionToEntity(mr *ent.MessageReaction) *entity.MessageReaction {
	if mr == nil {
		return nil
	}
	return &entity.MessageReaction{
		MessageID: mr.MessageID.String(),
		UserID:    mr.UserID.String(),
		Emoji:     mr.Emoji,
		CreatedAt: mr.CreatedAt,
	}
}

// MessageBookmark converters
func MessageBookmarkToEntity(mb *ent.MessageBookmark) *entity.MessageBookmark {
	if mb == nil {
		return nil
	}
	return &entity.MessageBookmark{
		UserID:    mb.UserID.String(),
		MessageID: mb.MessageID.String(),
		Message:   MessageToEntity(mb.Edges.Message),
		CreatedAt: mb.CreatedAt,
	}
}

// ChannelReadState converters
func ChannelReadStateToEntity(crs *ent.ChannelReadState) *entity.ChannelReadState {
	if crs == nil {
		return nil
	}
	return &entity.ChannelReadState{
		ChannelID:  crs.ChannelID.String(),
		UserID:     crs.UserID.String(),
		LastReadAt: crs.LastReadAt,
	}
}

// Attachment converters
func AttachmentToEntity(a *ent.Attachment) *entity.Attachment {
	if a == nil {
		return nil
	}

	var thumbnail *entity.Thumbnail
	if a.ThumbnailStorageKey != nil && a.ThumbnailWidth != nil && a.ThumbnailHeight != nil {
		thumbnail = &entity.Thumbnail{StorageKey: *a.ThumbnailStorageKey, Width: *a.ThumbnailWidth, Height: *a.ThumbnailHeight}
	}
	return &entity.Attachment{
		ID:         a.ID.String(),
		MessageID:  UUIDPtrToStringPtr(a.MessageID),
		UploaderID: a.UploaderID.String(),
		ChannelID:  a.ChannelID.String(),
		FileName:   a.FileName,
		MimeType:   a.MimeType,
		SizeBytes:  a.SizeBytes,
		Media:      entity.MediaMetadata{Width: a.Width, Height: a.Height, DurationSeconds: a.DurationSeconds, Thumbnail: thumbnail},
		StorageKey: a.StorageKey,
		Status:     entity.AttachmentStatus(a.Status),
		UploadedAt: &a.UploadedAt,
		ExpiresAt:  &a.ExpiresAt,
		CreatedAt:  a.CreatedAt,
	}
}

// UserGroup converters
func UserGroupToEntity(ug *ent.UserGroup) *entity.UserGroup {
	if ug == nil {
		return nil
	}
	return &entity.UserGroup{
		ID:          ug.ID.String(),
		WorkspaceID: ug.WorkspaceID,
		Name:        ug.Name,
		Description: StringPtrFromNullable(ug.Description),
		CreatedBy:   ug.CreatedByID.String(),
		CreatedAt:   ug.CreatedAt,
		UpdatedAt:   ug.UpdatedAt,
	}
}

// UserGroupMember converters
func UserGroupMemberToEntity(ugm *ent.UserGroupMember) *entity.UserGroupMember {
	if ugm == nil {
		return nil
	}
	return &entity.UserGroupMember{
		GroupID:  ugm.GroupID.String(),
		UserID:   ugm.UserID.String(),
		JoinedAt: ugm.JoinedAt,
	}
}

// MessageUserMention converters
func MessageUserMentionToEntity(mum *ent.MessageUserMention) *entity.MessageUserMention {
	if mum == nil {
		return nil
	}
	return &entity.MessageUserMention{
		MessageID:  mum.MessageID.String(),
		UserID:     mum.UserID.String(),
		ViaGroupID: UUIDPtrToStringPtr(mum.ViaGroupID),
		CreatedAt:  mum.CreatedAt,
	}
}

// MessageGroupMention converters
func MessageGroupMentionToEntity(mgm *ent.MessageGroupMention) *entity.MessageGroupMention {
	if mgm == nil {
		return nil
	}
	return &entity.MessageGroupMention{
		MessageID: mgm.MessageID.String(),
		GroupID:   mgm.GroupID.String(),
		CreatedAt: mgm.CreatedAt,
	}
}

// MessageLinkToEntity は保存済みのリンクを変換します。OGP は Edges.LinkPreview を読み込んだときだけ埋まります
func MessageLinkToEntity(ml *ent.MessageLink) *entity.MessageLink {
	if ml == nil {
		return nil
	}
	return &entity.MessageLink{
		ID:              ml.ID.String(),
		MessageID:       ml.MessageID.String(),
		URL:             ml.URL,
		OGP:             LinkPreviewToOGP(ml.Edges.LinkPreview),
		LinkedMessageID: UUIDPtrToStringPtr(ml.LinkedMessageID),
		CreatedAt:       ml.CreatedAt,
	}
}

// LinkPreviewToOGP は YouTube と X の付加情報を読み込んだ link_preview を OGP に変換します
func LinkPreviewToOGP(lp *ent.LinkPreview) entity.OGPData {
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

// Helper functions
func StringPtrFromNullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func TimePtrFromNullable(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return t
}

func ParseUUIDOrNil(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

func ParseUUIDPtrOrNil(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	id, _ := uuid.Parse(*s)
	return &id
}
