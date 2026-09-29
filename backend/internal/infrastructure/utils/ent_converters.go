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
	return &entity.User{
		ID:           u.ID.String(),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		GoogleSub:    u.GoogleSub,
		DisplayName:  u.DisplayName,
		Bio:          StringPtrFromNullable(u.Bio),
		AvatarURL:    StringPtrFromNullable(u.AvatarURL),
		IsBot:        u.IsBot,
		Preferences: entity.UserPreferences{
			ThemeHue:     u.ThemeHue,
			ThemeChroma:  u.ThemeChroma,
			ThemeSidebar: entity.SidebarStyle(u.ThemeSidebar),
			ColorMode:    entity.ColorMode(u.ColorMode),
			Locale:       u.Locale,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// Session converters
func SessionToEntity(s *ent.Session) *entity.Session {
	if s == nil {
		return nil
	}
	var userID string
	if s.Edges.User != nil {
		userID = s.Edges.User.ID.String()
	}
	var revokedAt *time.Time
	if !s.RevokedAt.IsZero() {
		revokedAt = &s.RevokedAt
	}
	return &entity.Session{
		ID:               s.ID.String(),
		UserID:           userID,
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
	var createdBy string
	if w.Edges.CreatedBy != nil {
		createdBy = w.Edges.CreatedBy.ID.String()
	}
	return &entity.Workspace{
		ID:          w.ID,
		Name:        w.Name,
		Description: StringPtrFromNullable(w.Description),
		IconURL:     StringPtrFromNullable(w.IconURL),
		IsPublic:    w.IsPublic,
		CreatedBy:   createdBy,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

// WorkspaceMember converters
func WorkspaceMemberToEntity(wm *ent.WorkspaceMember) *entity.WorkspaceMember {
	if wm == nil {
		return nil
	}
	var workspaceID, userID string
	if wm.Edges.Workspace != nil {
		workspaceID = wm.Edges.Workspace.ID
	}
	if wm.Edges.User != nil {
		userID = wm.Edges.User.ID.String()
	}
	return &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
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
	var workspaceID, createdBy string
	if c.Edges.Workspace != nil {
		workspaceID = c.Edges.Workspace.ID
	}
	if c.Edges.CreatedBy != nil {
		createdBy = c.Edges.CreatedBy.ID.String()
	}

	channelType := entity.ChannelTypePublic
	if c.ChannelType != "" {
		channelType = entity.ChannelType(c.ChannelType)
	}

	return &entity.Channel{
		ID:          c.ID.String(),
		WorkspaceID: workspaceID,
		Name:        c.Name,
		Description: StringPtrFromNullable(c.Description),
		IsPrivate:   c.IsPrivate,
		Type:        channelType,
		ParentID:    UUIDPtrToStringPtr(c.ParentID),
		CreatedBy:   createdBy,
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
	var channelID, userID string
	if cm.Edges.Channel != nil {
		channelID = cm.Edges.Channel.ID.String()
	}
	if cm.Edges.User != nil {
		userID = cm.Edges.User.ID.String()
	}
	return &entity.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
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

		SenderName:      m.SenderName,
		SenderAvatarURL: m.SenderAvatarURL,
	}
}

// MessageReaction converters
func MessageReactionToEntity(mr *ent.MessageReaction) *entity.MessageReaction {
	if mr == nil {
		return nil
	}
	var messageID, userID string
	if mr.Edges.Message != nil {
		messageID = mr.Edges.Message.ID.String()
	}
	if mr.Edges.User != nil {
		userID = mr.Edges.User.ID.String()
	}
	return &entity.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     mr.Emoji,
		CreatedAt: mr.CreatedAt,
	}
}

// MessageBookmark converters
func MessageBookmarkToEntity(mb *ent.MessageBookmark) *entity.MessageBookmark {
	if mb == nil {
		return nil
	}
	var userID, messageID string
	var message *entity.Message
	if mb.Edges.User != nil {
		userID = mb.Edges.User.ID.String()
	}
	if mb.Edges.Message != nil {
		messageID = mb.Edges.Message.ID.String()
		message = MessageToEntity(mb.Edges.Message)
	}
	return &entity.MessageBookmark{
		UserID:    userID,
		MessageID: messageID,
		Message:   message,
		CreatedAt: mb.CreatedAt,
	}
}

// ChannelReadState converters
func ChannelReadStateToEntity(crs *ent.ChannelReadState) *entity.ChannelReadState {
	if crs == nil {
		return nil
	}
	var channelID, userID string
	if crs.Edges.Channel != nil {
		channelID = crs.Edges.Channel.ID.String()
	}
	if crs.Edges.User != nil {
		userID = crs.Edges.User.ID.String()
	}
	return &entity.ChannelReadState{
		ChannelID:  channelID,
		UserID:     userID,
		LastReadAt: crs.LastReadAt,
	}
}

// Attachment converters
func AttachmentToEntity(a *ent.Attachment) *entity.Attachment {
	if a == nil {
		return nil
	}

	var messageID *string
	if a.Edges.Message != nil {
		mid := a.Edges.Message.ID.String()
		messageID = &mid
	}

	var uploaderID, channelID string
	if a.Edges.Uploader != nil {
		uploaderID = a.Edges.Uploader.ID.String()
	}
	if a.Edges.Channel != nil {
		channelID = a.Edges.Channel.ID.String()
	}
	var thumbnail *entity.Thumbnail
	if a.ThumbnailStorageKey != nil && a.ThumbnailWidth != nil && a.ThumbnailHeight != nil {
		thumbnail = &entity.Thumbnail{StorageKey: *a.ThumbnailStorageKey, Width: *a.ThumbnailWidth, Height: *a.ThumbnailHeight}
	}
	return &entity.Attachment{
		ID:         a.ID.String(),
		MessageID:  messageID,
		UploaderID: uploaderID,
		ChannelID:  channelID,
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
	var workspaceID, createdBy string
	if ug.Edges.Workspace != nil {
		workspaceID = ug.Edges.Workspace.ID
	}
	if ug.Edges.CreatedBy != nil {
		createdBy = ug.Edges.CreatedBy.ID.String()
	}
	return &entity.UserGroup{
		ID:          ug.ID.String(),
		WorkspaceID: workspaceID,
		Name:        ug.Name,
		Description: StringPtrFromNullable(ug.Description),
		CreatedBy:   createdBy,
		CreatedAt:   ug.CreatedAt,
		UpdatedAt:   ug.UpdatedAt,
	}
}

// UserGroupMember converters
func UserGroupMemberToEntity(ugm *ent.UserGroupMember) *entity.UserGroupMember {
	if ugm == nil {
		return nil
	}
	var groupID, userID string
	if ugm.Edges.Group != nil {
		groupID = ugm.Edges.Group.ID.String()
	}
	if ugm.Edges.User != nil {
		userID = ugm.Edges.User.ID.String()
	}
	return &entity.UserGroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: ugm.JoinedAt,
	}
}

// MessageUserMention converters
func MessageUserMentionToEntity(mum *ent.MessageUserMention) *entity.MessageUserMention {
	if mum == nil {
		return nil
	}
	var messageID, userID string
	if mum.Edges.Message != nil {
		messageID = mum.Edges.Message.ID.String()
	}
	if mum.Edges.User != nil {
		userID = mum.Edges.User.ID.String()
	}
	return &entity.MessageUserMention{
		MessageID: messageID,
		UserID:    userID,
		CreatedAt: mum.CreatedAt,
	}
}

// MessageGroupMention converters
func MessageGroupMentionToEntity(mgm *ent.MessageGroupMention) *entity.MessageGroupMention {
	if mgm == nil {
		return nil
	}
	var messageID, groupID string
	if mgm.Edges.Message != nil {
		messageID = mgm.Edges.Message.ID.String()
	}
	if mgm.Edges.Group != nil {
		groupID = mgm.Edges.Group.ID.String()
	}
	return &entity.MessageGroupMention{
		MessageID: messageID,
		GroupID:   groupID,
		CreatedAt: mgm.CreatedAt,
	}
}

// MessageLink converters
func MessageLinkToEntity(ml *ent.MessageLink) *entity.MessageLink {
	if ml == nil {
		return nil
	}
	var messageID string
	if ml.Edges.Message != nil {
		messageID = ml.Edges.Message.ID.String()
	}
	var youtube *entity.YouTubeVideo
	if ml.YoutubeVideoID != nil {
		youtube = &entity.YouTubeVideo{
			VideoID:         *ml.YoutubeVideoID,
			ChannelName:     ml.YoutubeChannelName,
			DurationSeconds: ml.YoutubeDurationSeconds,
		}
	}
	return &entity.MessageLink{
		ID:        ml.ID.String(),
		MessageID: messageID,
		URL:       ml.URL,
		OGP: entity.OGPData{
			Title:       StringPtrFromNullable(ml.Title),
			Description: StringPtrFromNullable(ml.Description),
			ImageURL:    StringPtrFromNullable(ml.ImageURL),
			SiteName:    StringPtrFromNullable(ml.SiteName),
			CardType:    StringPtrFromNullable(ml.CardType),
			ImageWidth:  ml.ImageWidth,
			ImageHeight: ml.ImageHeight,
			YouTube:     youtube,
		},
		LinkedMessageID: UUIDPtrToStringPtr(ml.LinkedMessageID),
		CreatedAt:       ml.CreatedAt,
	}
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
