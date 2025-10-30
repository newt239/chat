package bookmark

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBookmarkInteractor_AddBookmark(t *testing.T) {
	tests := []struct {
		name          string
		input         AddBookmarkInput
		setupMocks    func(*mocks.MockBookmarkRepository, *mocks.MockMessageRepository, *mocks.MockChannelRepository, *mocks.MockChannelMemberRepository, *mocks.MockWorkspaceRepository)
		expectedError error
	}{
		{
			name: "successful bookmark addition",
			input: AddBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, _ *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				message := &entity.Message{
					ID:        "msg1",
					ChannelID: "ch1",
					UserID:    "user1",
					Body:      "test message",
				}
				channel := &entity.Channel{
					ID:          "ch1",
					WorkspaceID: "ws1",
					IsPrivate:   false,
				}
				workspaceMember := &entity.WorkspaceMember{
					WorkspaceID: "ws1",
					UserID:      "user1",
					Role:        entity.WorkspaceRoleMember,
				}

				messageRepo.On("FindByID", mock.Anything, "msg1").Return(message, nil)
				channelRepo.On("FindByID", mock.Anything, "ch1").Return(channel, nil)
				workspaceRepo.On("FindMember", mock.Anything, "ws1", "user1").Return(workspaceMember, nil)
				bookmarkRepo.On("IsBookmarked", mock.Anything, "user1", "msg1").Return(false, nil)
				bookmarkRepo.On("AddBookmark", mock.Anything, mock.AnythingOfType("*entity.MessageBookmark")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "message not found",
			input: AddBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, _ *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				messageRepo.On("FindByID", mock.Anything, "msg1").Return(nil, nil)
			},
			expectedError: ErrMessageNotFound,
		},
		{
			name: "bookmark already exists",
			input: AddBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, _ *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				message := &entity.Message{
					ID:        "msg1",
					ChannelID: "ch1",
					UserID:    "user1",
					Body:      "test message",
				}
				channel := &entity.Channel{
					ID:          "ch1",
					WorkspaceID: "ws1",
					IsPrivate:   false,
				}
				workspaceMember := &entity.WorkspaceMember{
					WorkspaceID: "ws1",
					UserID:      "user1",
					Role:        entity.WorkspaceRoleMember,
				}

				messageRepo.On("FindByID", mock.Anything, "msg1").Return(message, nil)
				channelRepo.On("FindByID", mock.Anything, "ch1").Return(channel, nil)
				workspaceRepo.On("FindMember", mock.Anything, "ws1", "user1").Return(workspaceMember, nil)
				bookmarkRepo.On("IsBookmarked", mock.Anything, "user1", "msg1").Return(true, nil)
			},
			expectedError: ErrBookmarkExists,
		},
		{
			name: "user not member of workspace",
			input: AddBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, _ *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				message := &entity.Message{
					ID:        "msg1",
					ChannelID: "ch1",
					UserID:    "user1",
					Body:      "test message",
				}
				channel := &entity.Channel{
					ID:          "ch1",
					WorkspaceID: "ws1",
					IsPrivate:   false,
				}

				messageRepo.On("FindByID", mock.Anything, "msg1").Return(message, nil)
				channelRepo.On("FindByID", mock.Anything, "ch1").Return(channel, nil)
				workspaceRepo.On("FindMember", mock.Anything, "ws1", "user1").Return(nil, nil)
			},
			expectedError: ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkRepo := mocks.NewMockBookmarkRepository(t)
			messageRepo := mocks.NewMockMessageRepository(t)
			channelRepo := mocks.NewMockChannelRepository(t)
			channelMemberRepo := mocks.NewMockChannelMemberRepository(t)
			workspaceRepo := mocks.NewMockWorkspaceRepository(t)

			tt.setupMocks(bookmarkRepo, messageRepo, channelRepo, channelMemberRepo, workspaceRepo)

			interactor := NewBookmarkInteractor(
				bookmarkRepo,
				messageRepo,
				channelRepo,
				channelMemberRepo,
				workspaceRepo,
				nil, // userRepo
				nil, // mentionRepo
				nil, // groupMentionRepo
				nil, // linkRepo
				nil, // attachmentRepo
				nil, // userGroupRepo
				nil, // channelAccessSvc
			)

			err := interactor.AddBookmark(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			bookmarkRepo.AssertExpectations(t)
			messageRepo.AssertExpectations(t)
			channelRepo.AssertExpectations(t)
			channelMemberRepo.AssertExpectations(t)
			workspaceRepo.AssertExpectations(t)
		})
	}
}

func TestBookmarkInteractor_RemoveBookmark(t *testing.T) {
	tests := []struct {
		name          string
		input         RemoveBookmarkInput
		setupMocks    func(*mocks.MockBookmarkRepository, *mocks.MockMessageRepository, *mocks.MockChannelRepository, *mocks.MockChannelMemberRepository, *mocks.MockWorkspaceRepository)
		expectedError error
	}{
		{
			name: "successful bookmark removal",
			input: RemoveBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, _ *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {

				message := &entity.Message{
					ID:        "msg1",
					ChannelID: "ch1",
					UserID:    "user1",
					Body:      "test message",
				}
				channel := &entity.Channel{
					ID:          "ch1",
					WorkspaceID: "ws1",
					IsPrivate:   false,
				}
				workspaceMember := &entity.WorkspaceMember{
					WorkspaceID: "ws1",
					UserID:      "user1",
					Role:        entity.WorkspaceRoleMember,
				}

				messageRepo.On("FindByID", mock.Anything, "msg1").Return(message, nil)
				channelRepo.On("FindByID", mock.Anything, "ch1").Return(channel, nil)
				workspaceRepo.On("FindMember", mock.Anything, "ws1", "user1").Return(workspaceMember, nil)
				bookmarkRepo.On("RemoveBookmark", mock.Anything, "user1", "msg1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "message not found",
			input: RemoveBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, _ *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				messageRepo.On("FindByID", mock.Anything, "msg1").Return(nil, nil)
			},
			expectedError: ErrMessageNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkRepo := mocks.NewMockBookmarkRepository(t)
			messageRepo := mocks.NewMockMessageRepository(t)
			channelRepo := mocks.NewMockChannelRepository(t)
			channelMemberRepo := mocks.NewMockChannelMemberRepository(t)
			workspaceRepo := mocks.NewMockWorkspaceRepository(t)

			tt.setupMocks(bookmarkRepo, messageRepo, channelRepo, channelMemberRepo, workspaceRepo)

			interactor := NewBookmarkInteractor(
				bookmarkRepo,
				messageRepo,
				channelRepo,
				channelMemberRepo,
				workspaceRepo,
				nil, // userRepo
				nil, // mentionRepo
				nil, // groupMentionRepo
				nil, // linkRepo
				nil, // attachmentRepo
				nil, // userGroupRepo
				nil, // channelAccessSvc
			)

			err := interactor.RemoveBookmark(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			bookmarkRepo.AssertExpectations(t)
			messageRepo.AssertExpectations(t)
			channelRepo.AssertExpectations(t)
			channelMemberRepo.AssertExpectations(t)
			workspaceRepo.AssertExpectations(t)
		})
	}
}

func TestBookmarkInteractor_ListBookmarks(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*mocks.MockBookmarkRepository, *mocks.MockMessageRepository, *mocks.MockChannelRepository, *mocks.MockChannelMemberRepository, *mocks.MockWorkspaceRepository)
		expectedOutput *ListBookmarksOutput
		expectedError  error
	}{
		{
			name:   "successful bookmark listing",
			userID: "user1",
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, channelMemberRepo *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				_ = channelMemberRepo

				now := time.Now()
				bookmarks := []*entity.MessageBookmark{
					{
						UserID:    "user1",
						MessageID: "msg1",
						CreatedAt: now,
					},
					{
						UserID:    "user1",
						MessageID: "msg2",
						CreatedAt: now.Add(-time.Hour),
					},
				}

				bookmarkRepo.On("FindByUserID", mock.Anything, "user1").Return(bookmarks, nil)
			},
			// ListBookmarks は bookmark.Message が nil の場合、その要素をスキップする実装のため
			// 本テスト条件では空配列が返ることを期待する
			expectedOutput: &ListBookmarksOutput{Bookmarks: []BookmarkWithMessageOutput{}},
			expectedError:  nil,
		},
		{
			name:   "empty bookmark list",
			userID: "user1",
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, channelMemberRepo *mocks.MockChannelMemberRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				_ = channelMemberRepo
				bookmarkRepo.On("FindByUserID", mock.Anything, "user1").Return([]*entity.MessageBookmark{}, nil)
			},
			expectedOutput: &ListBookmarksOutput{
				Bookmarks: []BookmarkWithMessageOutput{},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkRepo := mocks.NewMockBookmarkRepository(t)
			messageRepo := mocks.NewMockMessageRepository(t)
			channelRepo := mocks.NewMockChannelRepository(t)
			channelMemberRepo := mocks.NewMockChannelMemberRepository(t)
			workspaceRepo := mocks.NewMockWorkspaceRepository(t)

			tt.setupMocks(bookmarkRepo, messageRepo, channelRepo, channelMemberRepo, workspaceRepo)

			interactor := NewBookmarkInteractor(
				bookmarkRepo,
				messageRepo,
				channelRepo,
				channelMemberRepo,
				workspaceRepo,
				nil, // userRepo
				nil, // mentionRepo
				nil, // groupMentionRepo
				nil, // linkRepo
				nil, // attachmentRepo
				nil, // userGroupRepo
				nil, // channelAccessSvc
			)

			output, err := interactor.ListBookmarks(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != nil {
					assert.Equal(t, len(tt.expectedOutput.Bookmarks), len(output.Bookmarks))
				}
			}

			bookmarkRepo.AssertExpectations(t)
			messageRepo.AssertExpectations(t)
			channelRepo.AssertExpectations(t)
			channelMemberRepo.AssertExpectations(t)
			workspaceRepo.AssertExpectations(t)
		})
	}
}
