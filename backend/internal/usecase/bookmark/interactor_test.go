package bookmark

import (
	"context"
	"testing"
	"time"

	"github.com/example/chat/internal/domain/entity"
	"github.com/example/chat/internal/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBookmarkInteractor_AddBookmark(t *testing.T) {
	tests := []struct {
		name          string
		input         AddBookmarkInput
		setupMocks    func(*mocks.MockBookmarkRepository, *mocks.MockMessageRepository, *mocks.MockChannelRepository, *mocks.MockWorkspaceRepository)
		expectedError error
	}{
		{
			name: "successful bookmark addition",
			input: AddBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
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
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
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
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkRepo := mocks.NewMockBookmarkRepository(t)
			messageRepo := mocks.NewMockMessageRepository(t)
			channelRepo := mocks.NewMockChannelRepository(t)
			workspaceRepo := mocks.NewMockWorkspaceRepository(t)

			tt.setupMocks(bookmarkRepo, messageRepo, channelRepo, workspaceRepo)

			interactor := NewBookmarkInteractor(bookmarkRepo, messageRepo, channelRepo, workspaceRepo)

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
			workspaceRepo.AssertExpectations(t)
		})
	}
}

func TestBookmarkInteractor_RemoveBookmark(t *testing.T) {
	tests := []struct {
		name          string
		input         RemoveBookmarkInput
		setupMocks    func(*mocks.MockBookmarkRepository, *mocks.MockMessageRepository, *mocks.MockChannelRepository, *mocks.MockWorkspaceRepository)
		expectedError error
	}{
		{
			name: "successful bookmark removal",
			input: RemoveBookmarkInput{
				UserID:    "user1",
				MessageID: "msg1",
			},
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkRepo := mocks.NewMockBookmarkRepository(t)
			messageRepo := mocks.NewMockMessageRepository(t)
			channelRepo := mocks.NewMockChannelRepository(t)
			workspaceRepo := mocks.NewMockWorkspaceRepository(t)

			tt.setupMocks(bookmarkRepo, messageRepo, channelRepo, workspaceRepo)

			interactor := NewBookmarkInteractor(bookmarkRepo, messageRepo, channelRepo, workspaceRepo)

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
			workspaceRepo.AssertExpectations(t)
		})
	}
}

func TestBookmarkInteractor_ListBookmarks(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*mocks.MockBookmarkRepository, *mocks.MockMessageRepository, *mocks.MockChannelRepository, *mocks.MockWorkspaceRepository)
		expectedOutput *ListBookmarksOutput
		expectedError  error
	}{
		{
			name:   "successful bookmark listing",
			userID: "user1",
			setupMocks: func(bookmarkRepo *mocks.MockBookmarkRepository, messageRepo *mocks.MockMessageRepository, channelRepo *mocks.MockChannelRepository, workspaceRepo *mocks.MockWorkspaceRepository) {
				bookmarks := []*entity.MessageBookmark{
					{
						UserID:    "user1",
						MessageID: "msg1",
						CreatedAt: time.Now(),
					},
					{
						UserID:    "user1",
						MessageID: "msg2",
						CreatedAt: time.Now().Add(-time.Hour),
					},
				}

				bookmarkRepo.On("FindByUserID", mock.Anything, "user1").Return(bookmarks, nil)
			},
			expectedOutput: &ListBookmarksOutput{
				Bookmarks: []BookmarkOutput{
					{
						UserID:    "user1",
						MessageID: "msg1",
						CreatedAt: time.Now(),
					},
					{
						UserID:    "user1",
						MessageID: "msg2",
						CreatedAt: time.Now().Add(-time.Hour),
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkRepo := mocks.NewMockBookmarkRepository(t)
			messageRepo := mocks.NewMockMessageRepository(t)
			channelRepo := mocks.NewMockChannelRepository(t)
			workspaceRepo := mocks.NewMockWorkspaceRepository(t)

			tt.setupMocks(bookmarkRepo, messageRepo, channelRepo, workspaceRepo)

			interactor := NewBookmarkInteractor(bookmarkRepo, messageRepo, channelRepo, workspaceRepo)

			output, err := interactor.ListBookmarks(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}

			bookmarkRepo.AssertExpectations(t)
			messageRepo.AssertExpectations(t)
			channelRepo.AssertExpectations(t)
			workspaceRepo.AssertExpectations(t)
		})
	}
}
