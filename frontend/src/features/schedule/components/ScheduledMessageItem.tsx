import { useState } from "react";

import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { formatDateTime } from "@chat/i18n";
import {
  IconDotsVertical,
  IconEdit,
  IconMapPin,
  IconPaperclip,
  IconSend,
  IconTrash,
} from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Badge } from "#/components/ui/Badge";
import { IconButton } from "#/components/ui/IconButton";
import { LinkButton } from "#/components/ui/LinkButton";
import { Menu } from "#/components/ui/Menu";
import { MenuItem } from "#/components/ui/MenuItem";
import { TextArea } from "#/components/ui/TextArea";
import { ScheduledMessageStatus } from "#/gen/chat/v1/scheduled_message_service_pb";
import { toDate } from "#/lib/timestamp";
import { preferencesAtom } from "#/providers/store/preferences";

import { useScheduledMessageActions } from "../hooks/useScheduledMessages";
import { ScheduleDialog } from "./ScheduleDialog";

import type { ScheduledMessage } from "#/gen/chat/v1/scheduled_message_service_pb";

type ScheduledMessageItemProps = {
  workspaceId: string;
  message: ScheduledMessage;
  // 会話の名前（#チャンネル名や DM の相手）
  label: string;
};

// 予約の一覧の行。予約中と失敗したものは編集・今すぐ送信・削除でき、送信済みは投稿へ移動できる
export const ScheduledMessageItem = ({
  workspaceId,
  message,
  label,
}: ScheduledMessageItemProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const { remove, sendNow, update } = useScheduledMessageActions();
  const [body, setBody] = useState(message.body);
  const [isEditing, setIsEditing] = useState(false);
  const { status, channelId, parentId, sentMessageId } = message;
  const isSent = status === ScheduledMessageStatus.SENT;
  const isEditable =
    status === ScheduledMessageStatus.SCHEDULED || status === ScheduledMessageStatus.FAILED;
  const time = formatDateTime(toDate(isSent ? message.updatedAt : message.scheduledAt), locale);

  return (
    <article className="flex items-start gap-2 rounded-[10px] border border-border bg-surface py-2.5 pr-2 pl-3 font-sans text-text">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <header className="flex min-w-0 flex-wrap items-center gap-1.5 text-xs text-muted">
          {label && <b className="truncate font-semibold text-text">{label}</b>}
          <span>{t(isSent ? "schedule.list.sentAt" : "schedule.list.scheduledFor", { time })}</span>
          {status === ScheduledMessageStatus.SENDING && (
            <Badge tone="accent">{t("schedule.list.status.sending")}</Badge>
          )}
          {status === ScheduledMessageStatus.FAILED && (
            <Badge tone="accent" className="bg-danger text-danger-fg">
              {t("schedule.list.status.failed")}
            </Badge>
          )}
        </header>
        {message.body !== "" && (
          <p className="m-0 line-clamp-3 text-body break-words whitespace-pre-wrap">
            {message.body}
          </p>
        )}
        {(message.location !== undefined || message.attachmentIds.length > 0) && (
          <span className="flex flex-wrap gap-3 text-caption text-muted [&_svg]:size-3.5">
            {message.location && (
              <span className="inline-flex items-center gap-1">
                <IconMapPin aria-hidden />
                {message.location.label ?? t("location.card.title")}
              </span>
            )}
            {message.attachmentIds.length > 0 && (
              <span className="inline-flex items-center gap-1">
                <IconPaperclip aria-hidden />
                {t("schedule.list.attachments", { count: message.attachmentIds.length })}
              </span>
            )}
          </span>
        )}
        {message.failureReason !== undefined && (
          <p className="m-0 text-caption text-danger">{message.failureReason}</p>
        )}
      </div>
      {isSent &&
        sentMessageId !== undefined &&
        (parentId === undefined ? (
          <LinkButton
            variant="secondary"
            size="sm"
            to="/app/$workspaceId/$channelId"
            params={{ channelId, workspaceId }}
            search={{ message: sentMessageId }}
          >
            {t("schedule.list.showMessage")}
          </LinkButton>
        ) : (
          <LinkButton
            variant="secondary"
            size="sm"
            to="/app/$workspaceId/$channelId/thread/$messageId"
            params={{ channelId, messageId: parentId, workspaceId }}
            search={{ message: sentMessageId }}
          >
            {t("schedule.list.showMessage")}
          </LinkButton>
        ))}
      {status !== ScheduledMessageStatus.SENDING && (
        <Menu
          trigger={
            <IconButton label={t("schedule.list.actions")} className="size-7 [&_svg]:size-4">
              <IconDotsVertical />
            </IconButton>
          }
        >
          {isEditable && (
            <MenuItem
              icon={<IconEdit aria-hidden />}
              onAction={() => {
                setBody(message.body);
                setIsEditing(true);
              }}
            >
              {t("schedule.list.edit")}
            </MenuItem>
          )}
          {isEditable && (
            <MenuItem
              icon={<IconSend aria-hidden />}
              onAction={() => {
                sendNow.mutate({ id: message.id });
              }}
            >
              {t("schedule.list.sendNow")}
            </MenuItem>
          )}
          <MenuItem
            icon={<IconTrash aria-hidden />}
            tone="danger"
            onAction={() => {
              remove.mutate({ id: message.id });
            }}
          >
            {t("schedule.list.delete")}
          </MenuItem>
        </Menu>
      )}
      {isEditing && (
        <ScheduleDialog
          isOpen
          onOpenChange={setIsEditing}
          title={t("schedule.list.editTitle")}
          initialDate={toDate(message.scheduledAt)}
          isPending={update.isPending}
          onConfirm={(scheduledAt) => {
            update.mutate(
              { body, id: message.id, scheduledAt: timestampFromDate(scheduledAt) },
              {
                onSuccess: () => {
                  setIsEditing(false);
                },
              },
            );
          }}
        >
          <TextArea label={t("schedule.list.body")} value={body} onChange={setBody} rows={4} />
        </ScheduleDialog>
      )}
    </article>
  );
};
