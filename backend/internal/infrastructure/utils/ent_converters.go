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
		DisplayName:  u.DisplayName,
		AvatarURL:    StringPtrFromNullable(u.AvatarURL),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
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
		ID:          w.ID.String(),
		Name:        w.Name,
		Description: StringPtrFromNullable(w.Description),
		IconURL:     StringPtrFromNullable(w.IconURL),
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
		workspaceID = wm.Edges.Workspace.ID.String()
	}
	if wm.Edges.User != nil {
		userID = wm.Edges.User.ID.String()
	}
	return &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        entity.WorkspaceRole(wm.Role),
		JoinedAt:    wm.JoinedAt,
	}
}

// Channel converters
func ChannelToEntity(c *ent.Channel) *entity.Channel {
	if c == nil {
		return nil
	}
	var workspaceID, createdBy string
	if c.Edges.Workspace != nil {
		workspaceID = c.Edges.Workspace.ID.String()
	}
	if c.Edges.CreatedBy != nil {
		createdBy = c.Edges.CreatedBy.ID.String()
	}
	return &entity.Channel{
		ID:          c.ID.String(),
		WorkspaceID: workspaceID,
		Name:        c.Name,
		Description: StringPtrFromNullable(c.Description),
		IsPrivate:   c.IsPrivate,
		CreatedBy:   createdBy,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
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
	if m.Edges.Parent != nil {
		pid := m.Edges.Parent.ID.String()
		parentID = &pid
	}

	var deletedBy *string
	if m.DeletedBy != uuid.Nil {
		db := m.DeletedBy.String()
		deletedBy = &db
	}

	var channelID, userID string
	if m.Edges.Channel != nil {
		channelID = m.Edges.Channel.ID.String()
	}
	if m.Edges.User != nil {
		userID = m.Edges.User.ID.String()
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
		ChannelID: channelID,
		UserID:    userID,
		ParentID:  parentID,
		Body:      m.Body,
		CreatedAt: m.CreatedAt,
		EditedAt:  editedAt,
		DeletedAt: deletedAt,
		DeletedBy: deletedBy,
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
	if mb.Edges.User != nil {
		userID = mb.Edges.User.ID.String()
	}
	if mb.Edges.Message != nil {
		messageID = mb.Edges.Message.ID.String()
	}
	return &entity.MessageBookmark{
		UserID:    userID,
		MessageID: messageID,
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
	return &entity.Attachment{
		ID:         a.ID.String(),
		MessageID:  messageID,
		UploaderID: uploaderID,
		ChannelID:  channelID,
		FileName:   a.FileName,
		MimeType:   a.MimeType,
		SizeBytes:  a.SizeBytes,
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
		workspaceID = ug.Edges.Workspace.ID.String()
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
	return &entity.MessageLink{
		ID:          ml.ID.String(),
		MessageID:   messageID,
		URL:         ml.URL,
		Title:       StringPtrFromNullable(ml.Title),
		Description: StringPtrFromNullable(ml.Description),
		ImageURL:    StringPtrFromNullable(ml.ImageURL),
		SiteName:    StringPtrFromNullable(ml.SiteName),
		CardType:    StringPtrFromNullable(ml.CardType),
		CreatedAt:   ml.CreatedAt,
	}
}

// ThreadMetadata converters
func ThreadMetadataToEntity(tm *ent.ThreadMetadata) *entity.ThreadMetadata {
	if tm == nil {
		return nil
	}

	var lastReplyUserID *string
	if tm.Edges.LastReplyUser != nil {
		lruid := tm.Edges.LastReplyUser.ID.String()
		lastReplyUserID = &lruid
	}

	participantIDs := make([]string, len(tm.ParticipantUserIds))
	for i, id := range tm.ParticipantUserIds {
		participantIDs[i] = id.String()
	}

	var messageID string
	if tm.Edges.Message != nil {
		messageID = tm.Edges.Message.ID.String()
	}
	return &entity.ThreadMetadata{
		MessageID:          messageID,
		ReplyCount:         tm.ReplyCount,
		LastReplyAt:        &tm.LastReplyAt,
		LastReplyUserID:    lastReplyUserID,
		ParticipantUserIDs: participantIDs,
		CreatedAt:          tm.CreatedAt,
		UpdatedAt:          tm.UpdatedAt,
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
