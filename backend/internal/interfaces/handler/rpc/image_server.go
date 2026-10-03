package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	imageuc "github.com/newt239/chat/internal/usecase/image"
)

type ImageServer struct {
	UC *imageuc.Interactor
}

var imagePurposes = map[chatv1.ImagePurpose]imageuc.Purpose{
	chatv1.ImagePurpose_IMAGE_PURPOSE_AVATAR:         imageuc.PurposeAvatar,
	chatv1.ImagePurpose_IMAGE_PURPOSE_WORKSPACE_ICON: imageuc.PurposeWorkspaceIcon,
	chatv1.ImagePurpose_IMAGE_PURPOSE_APP_ICON:       imageuc.PurposeAppIcon,
}

func (s *ImageServer) PresignImageUpload(ctx context.Context, req *chatv1.PresignImageUploadRequest) (*chatv1.PresignImageUploadResponse, error) {
	out, err := s.UC.Presign(ctx, imageuc.PresignInput{
		UserID:      userIDFrom(ctx),
		WorkspaceID: req.WorkspaceId,
		Purpose:     imagePurposes[req.Purpose],
		ContentType: req.ContentType,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.PresignImageUploadResponse{UploadUrl: out.UploadURL, ImageUrl: out.ImageURL}, nil
}
