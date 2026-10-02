import { callUnaryMethod, useMutation, useTransport } from "@connectrpc/connect-query";
import { useMutation as useQueryMutation, useQueryClient } from "@tanstack/react-query";

import { CustomEmojiService } from "#/gen/chat/v1/custom_emoji_service_pb";
import { putToStorage } from "#/lib/storage";

import { prepareEmojiImage } from "../utils/prepareEmojiImage";
import { customEmojiListKey } from "./useCustomEmojis";

type RegisterInput = {
  name: string;
  file: File;
};

/** 画像を整えてストレージに置き、名前を付けて登録する。削除も含め、成功したら一覧を取り直す */
export const useCustomEmojiActions = (workspaceId: string) => {
  const transport = useTransport();
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({ queryKey: customEmojiListKey(workspaceId) });
  };

  const register = useQueryMutation({
    mutationFn: async ({ name, file }: RegisterInput) => {
      const image = await prepareEmojiImage(file);
      const { uploadId, uploadUrl } = await callUnaryMethod(
        transport,
        CustomEmojiService.method.presignCustomEmojiUpload,
        { contentType: image.type, sizeBytes: BigInt(image.size), workspaceId },
      );
      await putToStorage(image, uploadUrl, () => {});
      return callUnaryMethod(transport, CustomEmojiService.method.createCustomEmoji, {
        name,
        uploadId,
        workspaceId,
      });
    },
    onSuccess,
  });

  return {
    register,
    remove: useMutation(CustomEmojiService.method.deleteCustomEmoji, { onSuccess }),
  };
};
