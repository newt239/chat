import { formatRelativeTime } from "@chat/i18n/format";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { focusRing } from "#/components/ui/styles/styles";
import { usePreferences } from "#/hooks/usePreferences";
import { toDate } from "#/lib/timestamp";

import type { ThreadMetadata } from "#/gen/chat/v1/message_pb";

type ThreadMetadataPreviewProps = {
  metadata: ThreadMetadata;
  onPress: () => void;
};

export const ThreadMetadataPreview = ({ metadata, onPress }: ThreadMetadataPreviewProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const { lastReplyUser, lastReplyAt, replyCount } = metadata;

  return (
    <Button
      onPress={onPress}
      className={`-ml-[3px] inline-flex cursor-pointer items-center gap-[7px] self-start rounded-md border border-transparent py-0.5 pr-2 pl-[3px] font-sans data-hovered:border-border data-hovered:bg-surface ${focusRing}`}
    >
      {lastReplyUser && (
        <Avatar name={lastReplyUser.displayName} src={lastReplyUser.avatarUrl} size={20} />
      )}
      <b className="text-[12.5px] font-semibold text-accent-text">
        {t("message.thread.replies", { count: replyCount })}
      </b>
      {lastReplyAt && (
        <span className="text-[11.5px] text-subtle">
          {t("message.thread.lastReply", {
            time: formatRelativeTime(toDate(lastReplyAt), new Date(), locale),
          })}
        </span>
      )}
    </Button>
  );
};
