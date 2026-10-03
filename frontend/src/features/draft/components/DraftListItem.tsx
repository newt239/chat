import { IconTrash } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Badge } from "#/components/ui/Badge/Badge";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { LinkButton } from "#/components/ui/LinkButton/LinkButton";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { messageLocation } from "#/features/message/utils/messageLocation";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";

import type { Draft } from "#/gen/chat/v1/draft_service_pb";

type DraftListItemProps = {
  workspaceId: string;
  draft: Draft;
  // 会話の名前（#チャンネル名や DM の相手）
  label: string;
  onDelete: () => void;
};

// 下書きの一覧の行。開くと書きかけの入力欄に戻れる
export const DraftListItem = ({ workspaceId, draft, label, onDelete }: DraftListItemProps) => {
  const { t } = useTranslation();
  const { toText } = useMentionDirectory();
  const { formatDateTime } = useDateFormat();
  const { channelId, parentId } = draft;

  return (
    <article className="flex items-start gap-2 rounded-lg border border-border bg-surface py-2.5 pr-2 pl-3 font-sans text-text">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <header className="flex min-w-0 items-center gap-1.5 text-xs text-muted">
          {label && <b className="truncate font-semibold text-text">{label}</b>}
          {parentId !== undefined && <Badge tone="tag">{t("draft.list.inThread")}</Badge>}
          <span className="truncate">
            {t("draft.list.savedAt", { time: formatDateTime(toDate(draft.updatedAt)) })}
          </span>
        </header>
        <p className="m-0 line-clamp-3 text-body break-words whitespace-pre-wrap">
          {toText(draft.body)}
        </p>
      </div>
      <LinkButton
        variant="secondary"
        size="sm"
        {...messageLocation({ channelId, messageId: undefined, parentId, workspaceId })}
      >
        {t("draft.list.open")}
      </LinkButton>
      <IconButton
        label={t("draft.list.delete")}
        onPress={onDelete}
        className="size-7 [&_svg]:size-4"
      >
        <IconTrash />
      </IconButton>
    </article>
  );
};
