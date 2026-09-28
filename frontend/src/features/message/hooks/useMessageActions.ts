import { useCallback } from "react";

import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";
import { useDeleteMessage, useUpdateMessage } from "#/features/message/hooks/useMessage";

const errorDescription = (error: unknown) =>
  error instanceof Error && error.message ? error.message : undefined;

export const useMessageActions = () => {
  const { t } = useTranslation();
  const updateMessage = useUpdateMessage();
  const deleteMessage = useDeleteMessage();

  const handleEdit = useCallback(
    async (messageId: string, nextBody: string) => {
      try {
        await updateMessage.mutateAsync({ body: nextBody, messageId });
        toast(t("message.edit.done"), { tone: "success" });
      } catch (error) {
        toast(t("message.edit.failed"), { description: errorDescription(error), tone: "danger" });
        throw error;
      }
    },
    [updateMessage, t],
  );

  const handleDelete = useCallback(
    async (messageId: string) => {
      try {
        await deleteMessage.mutateAsync({ messageId });
        toast(t("message.delete.done"), { tone: "success" });
      } catch (error) {
        toast(t("message.delete.failed"), { description: errorDescription(error), tone: "danger" });
      }
    },
    [deleteMessage, t],
  );

  return { handleDelete, handleEdit, isDeleting: deleteMessage.isPending };
};
