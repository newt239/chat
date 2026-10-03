import { ConnectError } from "@connectrpc/connect";
import { useMutation } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { MessageService } from "#/gen/chat/v1/message_service_pb";

import { useInvalidateThreadMetadata } from "./useMessage";

import type { Message } from "#/gen/chat/v1/message_pb";

export const useMessageActions = () => {
  const { t } = useTranslation();
  const updateMessage = useMutation(MessageService.method.updateMessage);
  const deleteMessage = useMutation(MessageService.method.deleteMessage);
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
