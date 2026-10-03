import type { ReactNode } from "react";

import { IconInfoCircle } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { Trans, useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { SystemMessageKind } from "#/gen/chat/v1/message_pb";
import { toDate } from "#/lib/timestamp";

import { messageLocation } from "../utils/messageLocation";
import { MessageTime } from "./MessageTime";

import type { SystemMessage } from "#/gen/chat/v1/message_pb";

import type { JsonValue } from "@bufbuild/protobuf";

type SystemMessageItemProps = {
  message: SystemMessage;
};

const textOf = (value: JsonValue | undefined) => (typeof value === "string" ? value : "");

export const SystemMessageItem = ({ message }: SystemMessageItemProps) => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ strict: false });
  const displayName = useDisplayName();
  const payload = message.payload ?? {};
  const nameOf = (userId: string) => displayName(userId, userId);
  const actor = nameOf(message.actorId ?? "");
  const user = nameOf(textOf(payload.userId));
  const from = textOf(payload.from);
  const to = textOf(payload.to);

  const pinnedLink =
    workspaceId === undefined ? (
      <span />
    ) : (
      <Link
        {...messageLocation({
          channelId: message.channelId,
          messageId: textOf(payload.messageId),
          parentId: textOf(payload.parentId) || undefined,
          workspaceId,
        })}
      />
    );

  const texts: Partial<Record<SystemMessageKind, ReactNode>> = {
    [SystemMessageKind.MEMBER_JOINED]: t("message.system.memberJoined", { user }),
    [SystemMessageKind.MEMBER_ADDED]: t("message.system.memberAdded", {
      by: actor,
      user,
    }),
    [SystemMessageKind.CHANNEL_PRIVACY_CHANGED]: t("message.system.privacyChanged", {
      from: from || "public",
      to: to || "public",
    }),
    [SystemMessageKind.CHANNEL_NAME_CHANGED]: t("message.system.nameChanged", { from, to }),
    [SystemMessageKind.CHANNEL_DESCRIPTION_CHANGED]: t("message.system.descriptionChanged"),
    [SystemMessageKind.MESSAGE_PINNED]: (
      <Trans
        i18nKey="message.system.messagePinned"
        values={{ user: actor }}
        components={{ target: pinnedLink }}
      />
    ),
  };
  const createdAt = toDate(message.createdAt);

  return (
    <div className="flex items-center gap-2.5 px-4.5 py-0.75 font-sans text-label font-normal text-muted">
      <span className="grid w-8 shrink-0 place-items-center text-subtle [&_svg]:size-3.5">
        <IconInfoCircle aria-hidden />
      </span>
      <span className="min-w-0">{texts[message.kind] ?? t("message.system.unspecified")}</span>
      <MessageTime date={createdAt} />
    </div>
  );
};
