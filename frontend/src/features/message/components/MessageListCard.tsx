import type { ReactNode } from "react";

import { formatDateTime } from "@chat/i18n";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link";
import { useConversationLabel } from "#/features/channel/hooks/useConversationLabel";
import { toDate } from "#/lib/timestamp";
import { preferencesAtom } from "#/providers/store/preferences";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageListCardProps = {
  workspaceId: string;
  message: Message;
  children: ReactNode;
};

const linkClassName =
  "shrink-0 rounded-sm px-2 py-0.5 text-xs font-semibold text-accent-text no-underline data-hovered:bg-hover";

// 検索結果・スレッド一覧・メンション一覧のカード。見出しに会話の名前・日時・元の場所へのリンクを出す
export const MessageListCard = ({ workspaceId, message, children }: MessageListCardProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const label = useConversationLabel(workspaceId)(message.channelId);
  const { parentId } = message;

  return (
    <article className="rounded-[10px] border border-border bg-surface font-sans text-text">
      <header className="flex items-center gap-1.5 rounded-t-[10px] border-b border-border bg-sunken py-1.5 pr-2 pl-3 text-xs text-muted">
        {label && <b className="max-w-[50%] truncate font-semibold text-text">{label}</b>}
        <span className="min-w-0 flex-1 truncate">
          {formatDateTime(toDate(message.createdAt), locale)}
          {parentId !== undefined && ` · ${t("search.inThread")}`}
        </span>
        {parentId === undefined ? (
          <Link
            to="/app/$workspaceId/$channelId"
            params={{ channelId: message.channelId, workspaceId }}
            search={{ message: message.id }}
            className={linkClassName}
          >
            {t("search.showInChannel")}
          </Link>
        ) : (
          <Link
            to="/app/$workspaceId/$channelId/thread/$messageId"
            params={{ channelId: message.channelId, messageId: parentId, workspaceId }}
            search={{ message: message.id }}
            className={linkClassName}
          >
            {t("search.showInThread")}
          </Link>
        )}
      </header>
      {children}
    </article>
  );
};
