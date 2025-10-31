import { useCallback } from "react";

import { notifications } from "@mantine/notifications";

import { useDeleteMessage, useUpdateMessage } from "@/features/message/hooks/useMessage";

const resolveErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return fallback;
};

export const useMessageActions = (currentChannelId: string | null) => {
  const updateMessage = useUpdateMessage(currentChannelId);
  const deleteMessage = useDeleteMessage(currentChannelId);

  const handleEdit = useCallback(
    async (messageId: string, nextBody: string) => {
      try {
        await updateMessage.mutateAsync({ messageId, body: nextBody });
        notifications.show({
          title: "更新しました",
          message: "メッセージを更新しました",
        });
      } catch (error) {
        notifications.show({
          title: "エラー",
          message: resolveErrorMessage(error, "メッセージの更新に失敗しました"),
          color: "red",
        });
        throw error;
      }
    },
    [updateMessage]
  );

  const handleDelete = useCallback(
    async (messageId: string) => {
      try {
        await deleteMessage.mutateAsync({ messageId });
        notifications.show({
          title: "削除しました",
          message: "メッセージを削除しました",
        });
      } catch (error) {
        notifications.show({
          title: "エラー",
          message: resolveErrorMessage(error, "メッセージの削除に失敗しました"),
          color: "red",
        });
      }
    },
    [deleteMessage]
  );

  return { handleEdit, handleDelete };
};


