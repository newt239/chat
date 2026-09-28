import { skipToken, useMutation, useQuery } from "@connectrpc/connect-query";

import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";

// 署名付き URL の有効期限（5 分）より少し短い間だけ使い回す
const URL_STALE_TIME_MS = 4 * 60 * 1000;

export const usePresignUpload = () => useMutation(AttachmentService.method.presignUpload);

export const useDownloadUrl = () => useMutation(AttachmentService.method.getDownloadUrl);

// 画像の表示用。同じ添付を複数の場所で表示しても 1 回だけ取得する
export const useAttachmentUrl = (attachmentId: string | null) =>
  useQuery(
    AttachmentService.method.getDownloadUrl,
    attachmentId === null ? skipToken : { attachmentId },
    {
      refetchOnWindowFocus: false,
      select: (res) => res.url,
      staleTime: URL_STALE_TIME_MS,
    },
  );
