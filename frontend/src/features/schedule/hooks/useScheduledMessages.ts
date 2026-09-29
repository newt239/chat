import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { formatDateTime } from "@chat/i18n";
import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";
import { ScheduledMessageService } from "#/gen/chat/v1/scheduled_message_service_pb";
import { preferencesAtom } from "#/providers/store/preferences";

import type { ComposerContent } from "#/features/message/utils/composerContent";

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
  const { locale } = useAtomValue(preferencesAtom);
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
          toast(t("schedule.scheduled", { time: formatDateTime(scheduledAt, locale) }), {
            tone: "success",
          });
          void invalidate();
        },
      },
    );
  };
};
