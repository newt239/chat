import { formatDateTime } from "@chat/i18n";
import { useAtomValue } from "jotai";

import { Avatar } from "#/components/ui/Avatar";
import { Link } from "#/components/ui/Link";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { toDate } from "#/lib/timestamp";
import { preferencesAtom } from "#/providers/store/preferences";

import type { Message } from "#/gen/chat/v1/message_pb";

import type { Timestamp } from "@bufbuild/protobuf/wkt";

type MessageLinkCardProps = {
  message: Message;
  workspaceId: string;
  // ピン留め・ブックマークした日時
  markedAt: Timestamp | undefined;
};

// ピン留めやブックマークの一覧で、元のメッセージへ移動するカード
export const MessageLinkCard = ({ message, workspaceId, markedAt }: MessageLinkCardProps) => {
  const { locale } = useAtomValue(preferencesAtom);
  const name = useDisplayName()(message.userId, message.user?.displayName ?? "");

  return (
    <Link
      to="/app/$workspaceId/$channelId"
      params={{ channelId: message.channelId, workspaceId }}
      search={{ message: message.id }}
      className="flex gap-2.5 rounded-lg px-2 py-1.5 font-sans text-text no-underline data-hovered:bg-hover"
    >
      <Avatar name={name} src={message.user?.avatarUrl} size={28} />
      <span className="flex min-w-0 flex-1 flex-col gap-0.5 leading-[1.35]">
        <span className="flex items-baseline gap-2">
          <b className="truncate text-[13.5px] font-semibold">{name}</b>
          <span className="shrink-0 font-mono text-[11px] text-subtle tabular-nums">
            {formatDateTime(toDate(markedAt), locale)}
          </span>
        </span>
        <span className="line-clamp-3 text-[13px] whitespace-pre-wrap text-muted">
          {message.body}
        </span>
      </span>
    </Link>
  );
};
