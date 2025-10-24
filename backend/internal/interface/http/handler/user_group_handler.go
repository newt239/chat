package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/usecase/user_group"
)

type UserGroupHandler struct {
	userGroupUseCase user_group.UserGroupUseCase
}

func NewUserGroupHandler(userGroupUseCase user_group.UserGroupUseCase) *UserGroupHandler {
	return &UserGroupHandler{
		userGroupUseCase: userGroupUseCase,
	}
}

// CreateUserGroup グループ作成
func (h *UserGroupHandler) CreateUserGroup(c *gin.Context) {
	var input user_group.CreateUserGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.userGroupUseCase.CreateUserGroup(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, output)
}

// UpdateUserGroup グループ更新
func (h *UserGroupHandler) UpdateUserGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group ID is required"})
		return
	}

	var input user_group.UpdateUserGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id

	output, err := h.userGroupUseCase.UpdateUserGroup(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// DeleteUserGroup グループ削除
func (h *UserGroupHandler) DeleteUserGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group ID is required"})
		return
	}

	// リクエストからユーザーIDを取得（認証ミドルウェアで設定される想定）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := user_group.DeleteUserGroupInput{
		ID:        id,
		DeletedBy: userID.(string),
	}

	output, err := h.userGroupUseCase.DeleteUserGroup(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// GetUserGroup グループ取得
func (h *UserGroupHandler) GetUserGroup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group ID is required"})
		return
	}

	// リクエストからユーザーIDを取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := user_group.GetUserGroupInput{
		ID:     id,
		UserID: userID.(string),
	}

	output, err := h.userGroupUseCase.GetUserGroup(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// ListUserGroups グループ一覧取得
func (h *UserGroupHandler) ListUserGroups(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	// リクエストからユーザーIDを取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := user_group.ListUserGroupsInput{
		WorkspaceID: workspaceID,
		UserID:      userID.(string),
	}

	output, err := h.userGroupUseCase.ListUserGroups(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// AddMember メンバー追加
func (h *UserGroupHandler) AddMember(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group ID is required"})
		return
	}

	var requestBody struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// リクエストからユーザーIDを取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := user_group.AddMemberInput{
		GroupID: groupID,
		UserID:  requestBody.UserID,
		AddedBy: userID.(string),
	}

	output, err := h.userGroupUseCase.AddMember(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// RemoveMember メンバー削除
func (h *UserGroupHandler) RemoveMember(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group ID is required"})
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// リクエストからユーザーIDを取得
	requestUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := user_group.RemoveMemberInput{
		GroupID:   groupID,
		UserID:    userID,
		RemovedBy: requestUserID.(string),
	}

	output, err := h.userGroupUseCase.RemoveMember(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

// ListMembers メンバー一覧取得
func (h *UserGroupHandler) ListMembers(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group ID is required"})
		return
	}

	// リクエストからユーザーIDを取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := user_group.ListMembersInput{
		GroupID: groupID,
		UserID:  userID.(string),
	}

	output, err := h.userGroupUseCase.ListMembers(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
