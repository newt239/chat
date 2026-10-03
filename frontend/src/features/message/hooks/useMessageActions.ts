import { ConnectError } from "@connectrpc/connect";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import {
  useDeleteMessage,
  useInvalidateThreadMetadata,
  useUpdateMessage,
} from "#/features/message/hooks/useMessage";

import type { Message } from "#/gen/chat/v1/message_pb";

export const useMessageActions = () => {
  const { t } = useTranslation();
  const updateMessage = useUpdateMessage();
  const deleteMessage = useDeleteMessage();
  const invalidateThreadMetadata = useInvalidateThreadMetadata();

  const handleEdit = async (messageId: string, nextBody: string) => {
    try {
      await updateMessage.mutateAsync({ body: nextBody, messageId });
      toast(t("message.edit.done"), { tone: "success" });
    } catch (error) {
      toast(t("message.edit.failed"), {
        description: ConnectError.from(error).message || undefined,
        tone: "danger",
      });
      throw error;
    }
  };

  const handleDelete = async ({ id, channelId, parentId }: Message) => {
    try {
      await deleteMessage.mutateAsync({ messageId: id });
      toast(t("message.delete.done"), { tone: "success" });
      // 返信を消すと親のスレッドの件数が変わる
      if (parentId !== undefined) {
        await invalidateThreadMetadata(channelId);
      }
    } catch (error) {
      toast(t("message.delete.failed"), {
        description: ConnectError.from(error).message || undefined,
        tone: "danger",
      });
    }
  };

  return { handleDelete, handleEdit, isDeleting: deleteMessage.isPending };
};
