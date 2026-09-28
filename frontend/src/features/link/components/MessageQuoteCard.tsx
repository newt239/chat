import { formatDateTime } from "@chat/i18n";
import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { Link } from "#/components/ui/Link";
import { lastSegment } from "#/features/channel/utils/channelPath";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { toPlainText } from "#/features/message/utils/markdown/plainText";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { toDate } from "#/lib/timestamp";
import { preferencesAtom } from "#/providers/store/preferences";

import type { MessageLink } from "#/gen/chat/v1/message_pb";

const showLinkClassName =
  "rounded-md px-1.5 py-0.5 text-xs font-semibold no-underline data-hovered:bg-hover";

type MessageQuoteCardProps = {
  link: MessageLink;
};

// 同じワークスペースのメッセージへのリンクを引用カードにする。閲覧できないメッセージは何も出さない
export const MessageQuoteCard = ({ link }: MessageQuoteCardProps) => {
  const { t } = useTranslation();
  const displayName = useDisplayName();
  const { locale } = useAtomValue(preferencesAtom);
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
  const excerpt = toPlainText(preview.bodyExcerpt);

  return (
    <div className="flex w-[min(520px,100%)] flex-col gap-1 rounded-[10px] border border-border bg-surface px-3 py-2 font-sans">
      <div className="flex min-w-0 items-center gap-1.5 text-[12.5px]">
        <Avatar name={name} src={preview.user?.avatarUrl} size={18} />
        <b className="shrink-0 font-bold">{name}</b>
        <span className="truncate text-muted">
          #{lastSegment(preview.channelName)} · {formatDateTime(toDate(preview.createdAt), locale)}
        </span>
      </div>
      <p className="m-0 line-clamp-3 text-[13.5px] leading-normal">
        {excerpt || <span className="text-muted">{t("link.quote.attachmentOnly")}</span>}
      </p>
      <div className="flex justify-end">
        {preview.parentId === undefined ? (
          <Link
            to="/app/$workspaceId/$channelId"
            params={{ channelId: preview.channelId, workspaceId }}
            search={{ message: preview.messageId }}
            className={showLinkClassName}
          >
            {t("link.quote.show")}
          </Link>
        ) : (
          <Link
            to="/app/$workspaceId/$channelId/thread/$messageId"
            params={{ channelId: preview.channelId, messageId: preview.parentId, workspaceId }}
            search={{ message: preview.messageId }}
            className={showLinkClassName}
          >
            {t("link.quote.show")}
          </Link>
        )}
      </div>
    </div>
  );
};
