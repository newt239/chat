import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Link } from "#/components/ui/Link/Link";
import { lastSegment } from "#/features/channel/utils/channelPath";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";
import { messageLocation } from "#/lib/messageLocation";
import { toDate } from "#/lib/timestamp";

import type { MessageLink } from "#/gen/chat/v1/message_pb";

type MessageQuoteCardProps = {
  link: MessageLink;
};

// 同じワークスペースのメッセージへのリンクを引用カードにする。閲覧できないメッセージは何も出さない
export const MessageQuoteCard = ({ link }: MessageQuoteCardProps) => {
  const { t } = useTranslation();
  const { toExcerpt } = useMentionDirectory();
  const displayName = useDisplayName();
  const { formatDateTime } = useDateFormat();
  const { workspaceId } = useParams({ strict: false });
  // WebSocket で届いたメッセージには引用が含まれないため、あとから取得する
  const { data: fetched } = useQuery(
    MessageService.method.getMessagePreview,
    link.messagePreview === undefined && link.linkedMessageId !== undefined
      ? { messageId: link.linkedMessageId }
      : skipToken,
    { retry: false, select: (res) => res.preview },
  );
  const preview = link.messagePreview ?? fetched;

  if (preview === undefined || workspaceId === undefined) {
    return null;
  }
  const name = displayName(preview.user?.id ?? "", preview.user?.displayName ?? "");

  return (
    <div className="flex w-130 max-w-full flex-col gap-1 rounded-lg border border-border bg-surface px-3 py-2 font-sans">
      <div className="flex min-w-0 items-center gap-1.5 text-label font-normal">
        <Avatar name={name} src={preview.user?.avatarUrl} size={18} />
        <b className="shrink-0 font-bold">{name}</b>
        <span className="truncate text-muted">
          #{lastSegment(preview.channelName)} · {formatDateTime(toDate(preview.createdAt))}
        </span>
      </div>
      <p className="m-0 line-clamp-3 text-body-sm leading-normal">
        {toExcerpt(preview.bodyExcerpt)}
      </p>
      <div className="flex justify-end">
        <Link
          {...messageLocation({
            channelId: preview.channelId,
            messageId: preview.messageId,
            parentId: preview.parentId,
            workspaceId,
          })}
          className="rounded-md px-1.5 py-0.5 text-xs font-semibold no-underline data-hovered:bg-hover"
        >
          {t("link.quote.show")}
        </Link>
      </div>
    </div>
  );
};
