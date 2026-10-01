import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { ScheduledMessageService } from "#/gen/chat/v1/scheduled_message_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";

import type { ComposerContent } from "#/features/message/utils/composerContent";

export const useScheduledMessages = (workspaceId: string) =>
  useQuery(
    ScheduledMessageService.method.listScheduledMessages,
    { workspaceId },
    { select: (res) => res.scheduledMessages },
  );

const useInvalidateScheduledMessages = () => {
  const queryClient = useQueryClient();
  return () =>
    queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({ cardinality: "finite", schema: ScheduledMessageService }),
    });
};

type ScheduleTarget = ComposerContent & { channelId: string; parentId: string | undefined };

// 入力欄の内容を予約する。予約できたら onScheduled で入力欄を空にする
export const useScheduleMessage = () => {
  const { t } = useTranslation();
  const { formatDateTime } = useDateFormat();
  const invalidate = useInvalidateScheduledMessages();
  const { mutate } = useMutation(ScheduledMessageService.method.createScheduledMessage);

  return (target: ScheduleTarget, scheduledAt: Date, onScheduled: () => void) => {
    mutate(
      { ...target, scheduledAt: timestampFromDate(scheduledAt) },
      {
        onError: (error) => {
          toast(t("schedule.failed"), { description: error.message, tone: "danger" });
        },
        onSuccess: () => {
          onScheduled();
          toast(t("schedule.scheduled", { time: formatDateTime(scheduledAt) }), {
            tone: "success",
          });
          void invalidate();
        },
      },
    );
  };
};

// 一覧からの編集・今すぐ送信・削除。成功したら一覧を取り直し、結果をトーストで伝える
export const useScheduledMessageActions = () => {
  const { t } = useTranslation();
  const invalidate = useInvalidateScheduledMessages();
  const options = (
    done: "schedule.list.deleted" | "schedule.list.sent" | "schedule.list.updated",
  ) => ({
    onError: (error: Error) => {
      toast(t("schedule.list.actionFailed"), { description: error.message, tone: "danger" });
    },
    onSuccess: () => {
      toast(t(done), { tone: "success" });
      void invalidate();
    },
  });
  return {
    remove: useMutation(
      ScheduledMessageService.method.deleteScheduledMessage,
      options("schedule.list.deleted"),
    ),
    sendNow: useMutation(
      ScheduledMessageService.method.sendScheduledMessageNow,
      options("schedule.list.sent"),
    ),
    update: useMutation(
      ScheduledMessageService.method.updateScheduledMessage,
      options("schedule.list.updated"),
    ),
  };
};
