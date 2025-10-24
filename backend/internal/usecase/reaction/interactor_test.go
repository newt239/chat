package reaction

import (
	"errors"
	"testing"
	"time"

	"github.com/example/chat/internal/domain"
)

// Mock repositories for testing
type mockMessageRepository struct {
	findByIDFunc     func(id string) (*domain.Message, error)
	findReactionsFunc func(messageID string) ([]*domain.MessageReaction, error)
	addReactionFunc   func(reaction *domain.MessageReaction) error
	removeReactionFunc func(messageID, userID, emoji string) error
}

func (m *mockMessageRepository) FindByID(id string) (*domain.Message, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return nil, nil
}

func (m *mockMessageRepository) FindByChannelID(channelID string, limit int, since, until *time.Time) ([]*domain.Message, error) {
	return nil, nil
}

func (m *mockMessageRepository) FindThreadReplies(parentID string) ([]*domain.Message, error) {
	return nil, nil
}

func (m *mockMessageRepository) Create(message *domain.Message) error {
	return nil
}

func (m *mockMessageRepository) Update(message *domain.Message) error {
	return nil
}

func (m *mockMessageRepository) Delete(id string) error {
	return nil
}

func (m *mockMessageRepository) AddReaction(reaction *domain.MessageReaction) error {
	if m.addReactionFunc != nil {
		return m.addReactionFunc(reaction)
	}
	return nil
}

func (m *mockMessageRepository) RemoveReaction(messageID, userID, emoji string) error {
	if m.removeReactionFunc != nil {
		return m.removeReactionFunc(messageID, userID, emoji)
	}
	return nil
}

func (m *mockMessageRepository) FindReactions(messageID string) ([]*domain.MessageReaction, error) {
	if m.findReactionsFunc != nil {
		return m.findReactionsFunc(messageID)
	}
	return nil, nil
}

type mockChannelRepository struct {
	findByIDFunc func(id string) (*domain.Channel, error)
	isMemberFunc func(channelID, userID string) (bool, error)
}

func (m *mockChannelRepository) FindByID(id string) (*domain.Channel, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return nil, nil
}

func (m *mockChannelRepository) FindByWorkspaceID(workspaceID string) ([]*domain.Channel, error) {
	return nil, nil
}

func (m *mockChannelRepository) Create(channel *domain.Channel) error {
	return nil
}

func (m *mockChannelRepository) Update(channel *domain.Channel) error {
	return nil
}

func (m *mockChannelRepository) Delete(id string) error {
	return nil
}

func (m *mockChannelRepository) AddMember(channelID, userID string) error {
	return nil
}

func (m *mockChannelRepository) RemoveMember(channelID, userID string) error {
	return nil
}

func (m *mockChannelRepository) IsMember(channelID, userID string) (bool, error) {
	if m.isMemberFunc != nil {
		return m.isMemberFunc(channelID, userID)
	}
	return false, nil
}

type mockWorkspaceRepository struct {
	findMemberFunc func(workspaceID, userID string) (*domain.WorkspaceMember, error)
}

func (m *mockWorkspaceRepository) FindByID(id string) (*domain.Workspace, error) {
	return nil, nil
}

func (m *mockWorkspaceRepository) FindByUserID(userID string) ([]*domain.Workspace, error) {
	return nil, nil
}

func (m *mockWorkspaceRepository) Create(workspace *domain.Workspace) error {
	return nil
}

func (m *mockWorkspaceRepository) Update(workspace *domain.Workspace) error {
	return nil
}

func (m *mockWorkspaceRepository) Delete(id string) error {
	return nil
}

func (m *mockWorkspaceRepository) AddMember(member *domain.WorkspaceMember) error {
	return nil
}

func (m *mockWorkspaceRepository) UpdateMemberRole(workspaceID, userID string, role domain.Role) error {
	return nil
}

func (m *mockWorkspaceRepository) RemoveMember(workspaceID, userID string) error {
	return nil
}

func (m *mockWorkspaceRepository) FindMember(workspaceID, userID string) (*domain.WorkspaceMember, error) {
	if m.findMemberFunc != nil {
		return m.findMemberFunc(workspaceID, userID)
	}
	return nil, nil
}

func (m *mockWorkspaceRepository) ListMembers(workspaceID string) ([]*domain.WorkspaceMember, error) {
	return nil, nil
}

type mockUserRepository struct {
	findByIDsFunc func(ids []string) ([]*domain.User, error)
}

func (m *mockUserRepository) FindByID(id string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) FindByEmail(email string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) FindByIDs(ids []string) ([]*domain.User, error) {
	if m.findByIDsFunc != nil {
		return m.findByIDsFunc(ids)
	}
	return nil, nil
}

func (m *mockUserRepository) Create(user *domain.User) error {
	return nil
}

func (m *mockUserRepository) Update(user *domain.User) error {
	return nil
}

func (m *mockUserRepository) Delete(id string) error {
	return nil
}

func TestAddReaction(t *testing.T) {
	t.Run("成功: 公開チャンネルのメッセージにリアクションを追加", func(t *testing.T) {
		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return &domain.Message{
					ID:        "message-1",
					ChannelID: "channel-1",
					UserID:    "user-1",
					Body:      "Test message",
					CreatedAt: time.Now(),
				}, nil
			},
			addReactionFunc: func(reaction *domain.MessageReaction) error {
				if reaction.MessageID != "message-1" {
					t.Errorf("expected message ID 'message-1', got '%s'", reaction.MessageID)
				}
				if reaction.UserID != "user-2" {
					t.Errorf("expected user ID 'user-2', got '%s'", reaction.UserID)
				}
				if reaction.Emoji != "👍" {
					t.Errorf("expected emoji '👍', got '%s'", reaction.Emoji)
				}
				return nil
			},
		}

		channelRepo := &mockChannelRepository{
			findByIDFunc: func(id string) (*domain.Channel, error) {
				return &domain.Channel{
					ID:          "channel-1",
					WorkspaceID: "workspace-1",
					Name:        "general",
					IsPrivate:   false,
				}, nil
			},
		}

		workspaceRepo := &mockWorkspaceRepository{
			findMemberFunc: func(workspaceID, userID string) (*domain.WorkspaceMember, error) {
				return &domain.WorkspaceMember{
					WorkspaceID: "workspace-1",
					UserID:      "user-2",
					Role:        domain.RoleMember,
				}, nil
			},
		}

		interactor := NewReactionInteractor(messageRepo, channelRepo, workspaceRepo, &mockUserRepository{})

		input := AddReactionInput{
			MessageID: "message-1",
			UserID:    "user-2",
			Emoji:     "👍",
		}

		err := interactor.AddReaction(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("失敗: メッセージが存在しない", func(t *testing.T) {
		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return nil, nil
			},
		}

		interactor := NewReactionInteractor(messageRepo, &mockChannelRepository{}, &mockWorkspaceRepository{}, &mockUserRepository{})

		input := AddReactionInput{
			MessageID: "non-existent",
			UserID:    "user-1",
			Emoji:     "👍",
		}

		err := interactor.AddReaction(input)
		if err != ErrMessageNotFound {
			t.Errorf("expected ErrMessageNotFound, got %v", err)
		}
	})

	t.Run("失敗: ワークスペースメンバーでない", func(t *testing.T) {
		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return &domain.Message{
					ID:        "message-1",
					ChannelID: "channel-1",
					UserID:    "user-1",
					Body:      "Test message",
					CreatedAt: time.Now(),
				}, nil
			},
		}

		channelRepo := &mockChannelRepository{
			findByIDFunc: func(id string) (*domain.Channel, error) {
				return &domain.Channel{
					ID:          "channel-1",
					WorkspaceID: "workspace-1",
					Name:        "general",
					IsPrivate:   false,
				}, nil
			},
		}

		workspaceRepo := &mockWorkspaceRepository{
			findMemberFunc: func(workspaceID, userID string) (*domain.WorkspaceMember, error) {
				return nil, nil // メンバーでない
			},
		}

		interactor := NewReactionInteractor(messageRepo, channelRepo, workspaceRepo, &mockUserRepository{})

		input := AddReactionInput{
			MessageID: "message-1",
			UserID:    "user-2",
			Emoji:     "👍",
		}

		err := interactor.AddReaction(input)
		if err != ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestRemoveReaction(t *testing.T) {
	t.Run("成功: リアクションを削除", func(t *testing.T) {
		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return &domain.Message{
					ID:        "message-1",
					ChannelID: "channel-1",
					UserID:    "user-1",
					Body:      "Test message",
					CreatedAt: time.Now(),
				}, nil
			},
			removeReactionFunc: func(messageID, userID, emoji string) error {
				if messageID != "message-1" || userID != "user-2" || emoji != "👍" {
					t.Errorf("unexpected parameters: messageID=%s, userID=%s, emoji=%s", messageID, userID, emoji)
				}
				return nil
			},
		}

		channelRepo := &mockChannelRepository{
			findByIDFunc: func(id string) (*domain.Channel, error) {
				return &domain.Channel{
					ID:          "channel-1",
					WorkspaceID: "workspace-1",
					Name:        "general",
					IsPrivate:   false,
				}, nil
			},
		}

		workspaceRepo := &mockWorkspaceRepository{
			findMemberFunc: func(workspaceID, userID string) (*domain.WorkspaceMember, error) {
				return &domain.WorkspaceMember{
					WorkspaceID: "workspace-1",
					UserID:      "user-2",
					Role:        domain.RoleMember,
				}, nil
			},
		}

		interactor := NewReactionInteractor(messageRepo, channelRepo, workspaceRepo, &mockUserRepository{})

		input := RemoveReactionInput{
			MessageID: "message-1",
			UserID:    "user-2",
			Emoji:     "👍",
		}

		err := interactor.RemoveReaction(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestListReactions(t *testing.T) {
	t.Run("成功: リアクション一覧を取得", func(t *testing.T) {
		now := time.Now()
		avatarURL := "https://example.com/avatar.jpg"

		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return &domain.Message{
					ID:        "message-1",
					ChannelID: "channel-1",
					UserID:    "user-1",
					Body:      "Test message",
					CreatedAt: time.Now(),
				}, nil
			},
			findReactionsFunc: func(messageID string) ([]*domain.MessageReaction, error) {
				return []*domain.MessageReaction{
					{
						MessageID: "message-1",
						UserID:    "user-2",
						Emoji:     "👍",
						CreatedAt: now,
					},
					{
						MessageID: "message-1",
						UserID:    "user-3",
						Emoji:     "❤️",
						CreatedAt: now,
					},
				}, nil
			},
		}

		channelRepo := &mockChannelRepository{
			findByIDFunc: func(id string) (*domain.Channel, error) {
				return &domain.Channel{
					ID:          "channel-1",
					WorkspaceID: "workspace-1",
					Name:        "general",
					IsPrivate:   false,
				}, nil
			},
		}

		workspaceRepo := &mockWorkspaceRepository{
			findMemberFunc: func(workspaceID, userID string) (*domain.WorkspaceMember, error) {
				return &domain.WorkspaceMember{
					WorkspaceID: "workspace-1",
					UserID:      userID,
					Role:        domain.RoleMember,
				}, nil
			},
		}

		userRepo := &mockUserRepository{
			findByIDsFunc: func(ids []string) ([]*domain.User, error) {
				return []*domain.User{
					{
						ID:          "user-2",
						Email:       "user2@example.com",
						DisplayName: "User 2",
						AvatarURL:   &avatarURL,
					},
					{
						ID:          "user-3",
						Email:       "user3@example.com",
						DisplayName: "User 3",
						AvatarURL:   nil,
					},
				}, nil
			},
		}

		interactor := NewReactionInteractor(messageRepo, channelRepo, workspaceRepo, userRepo)

		output, err := interactor.ListReactions("message-1", "user-1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(output.Reactions) != 2 {
			t.Errorf("expected 2 reactions, got %d", len(output.Reactions))
		}

		if output.Reactions[0].User.DisplayName != "User 2" {
			t.Errorf("expected display name 'User 2', got '%s'", output.Reactions[0].User.DisplayName)
		}

		if output.Reactions[0].Emoji != "👍" {
			t.Errorf("expected emoji '👍', got '%s'", output.Reactions[0].Emoji)
		}
	})

	t.Run("成功: リアクションが0件の場合", func(t *testing.T) {
		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return &domain.Message{
					ID:        "message-1",
					ChannelID: "channel-1",
					UserID:    "user-1",
					Body:      "Test message",
					CreatedAt: time.Now(),
				}, nil
			},
			findReactionsFunc: func(messageID string) ([]*domain.MessageReaction, error) {
				return []*domain.MessageReaction{}, nil
			},
		}

		channelRepo := &mockChannelRepository{
			findByIDFunc: func(id string) (*domain.Channel, error) {
				return &domain.Channel{
					ID:          "channel-1",
					WorkspaceID: "workspace-1",
					Name:        "general",
					IsPrivate:   false,
				}, nil
			},
		}

		workspaceRepo := &mockWorkspaceRepository{
			findMemberFunc: func(workspaceID, userID string) (*domain.WorkspaceMember, error) {
				return &domain.WorkspaceMember{
					WorkspaceID: "workspace-1",
					UserID:      userID,
					Role:        domain.RoleMember,
				}, nil
			},
		}

		interactor := NewReactionInteractor(messageRepo, channelRepo, workspaceRepo, &mockUserRepository{})

		output, err := interactor.ListReactions("message-1", "user-1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(output.Reactions) != 0 {
			t.Errorf("expected 0 reactions, got %d", len(output.Reactions))
		}
	})

	t.Run("失敗: メッセージが存在しない", func(t *testing.T) {
		messageRepo := &mockMessageRepository{
			findByIDFunc: func(id string) (*domain.Message, error) {
				return nil, nil
			},
		}

		interactor := NewReactionInteractor(messageRepo, &mockChannelRepository{}, &mockWorkspaceRepository{}, &mockUserRepository{})

		_, err := interactor.ListReactions("non-existent", "user-1")
		if err != ErrMessageNotFound {
			t.Errorf("expected ErrMessageNotFound, got %v", err)
		}
	})
}
