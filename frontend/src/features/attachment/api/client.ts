import { useMutation } from "@connectrpc/connect-query";

import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";

export const usePresignUpload = () => useMutation(AttachmentService.method.presignUpload);

export const useDownloadUrl = () => useMutation(AttachmentService.method.getDownloadUrl);
