import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Link } from "#/components/ui/Link/Link";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/message/hooks/useMentionDirectory";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";

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
  const { toText } = useMentionDirectory();
  const { formatDateTime } = useDateFormat();
  const name = useDisplayName()(message.userId, message.user?.displayName ?? "");

  return (
    <Link
      to="/app/$workspaceId/$channelId"
      params={{ channelId: message.channelId, workspaceId }}
      search={{ message: message.id }}
      className="flex gap-2.5 rounded-lg px-2 py-1.5 font-sans text-text no-underline data-hovered:bg-hover"
    >
      <Avatar name={name} src={message.user?.avatarUrl} size={28} />
      <span className="flex min-w-0 flex-1 flex-col gap-0.5 leading-snug">
        <span className="flex items-baseline gap-2">
          <b className="truncate text-body-sm font-semibold">{name}</b>
          <span className="shrink-0 font-mono text-caption text-subtle tabular-nums">
            {formatDateTime(toDate(markedAt))}
          </span>
        </span>
        <span className="line-clamp-3 text-body-sm whitespace-pre-wrap text-muted">
          {toText(message.body)}
        </span>
      </span>
    </Link>
  );
};
