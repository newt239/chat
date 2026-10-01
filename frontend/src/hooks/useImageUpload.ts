import { callUnaryMethod, useTransport } from "@connectrpc/connect-query";
import { useMutation } from "@tanstack/react-query";

import { ImageService } from "#/gen/chat/v1/image_service_pb";

import type { ImagePurpose } from "#/gen/chat/v1/image_service_pb";

/** アイコン画像をストレージに置き、保存に使う配信用の URL を返す */
export const useImageUpload = (purpose: ImagePurpose, workspaceId: string | null) => {
  const transport = useTransport();
  return useMutation({
    mutationFn: async (image: Blob) => {
      const { uploadUrl, imageUrl } = await callUnaryMethod(
        transport,
        ImageService.method.presignImageUpload,
        {
          contentType: image.type,
          purpose,
          sizeBytes: BigInt(image.size),
          workspaceId: workspaceId ?? "",
        },
      );
      const res = await fetch(uploadUrl, {
        body: image,
        headers: { "Content-Type": image.type },
        method: "PUT",
      });
      if (!res.ok) {
        throw new Error(`upload failed: ${res.status}`);
      }
      return imageUrl;
    },
  });
};
