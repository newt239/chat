import { IconInfoCircle } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMembers } from "#/features/member/hooks/useMembers";
import { SystemMessageKind } from "#/gen/chat/v1/message_pb";
import { toDate } from "#/lib/timestamp";

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
  const { data: members } = useMembers(workspaceId ?? null);
  const displayName = useDisplayName();
  const payload = message.payload ?? {};
  const nameOf = (key: string) => {
    const userId = textOf(payload[key]);
    return displayName(
      userId,
      members?.find((member) => member.userId === userId)?.displayName ?? userId,
    );
  };
  const from = textOf(payload.from);
  const to = textOf(payload.to);

  const texts: Record<SystemMessageKind, () => string> = {
    [SystemMessageKind.UNSPECIFIED]: () => t("message.system.unspecified"),
    [SystemMessageKind.MEMBER_JOINED]: () =>
      t("message.system.memberJoined", { user: nameOf("userId") }),
    [SystemMessageKind.MEMBER_ADDED]: () =>
      t("message.system.memberAdded", { by: nameOf("addedBy"), user: nameOf("userId") }),
    [SystemMessageKind.MEMBER_REMOVED]: () =>
      t("message.system.memberRemoved", { user: nameOf("userId") }),
    [SystemMessageKind.MEMBER_LEFT]: () =>
      t("message.system.memberLeft", { user: nameOf("userId") }),
    [SystemMessageKind.CHANNEL_PRIVACY_CHANGED]: () =>
      t("message.system.privacyChanged", { from: from || "public", to: to || "public" }),
    [SystemMessageKind.CHANNEL_NAME_CHANGED]: () => t("message.system.nameChanged", { from, to }),
    [SystemMessageKind.CHANNEL_DESCRIPTION_CHANGED]: () => t("message.system.descriptionChanged"),
    [SystemMessageKind.MESSAGE_PINNED]: () =>
      t("message.system.messagePinned", { user: nameOf("pinnedBy") }),
  };
  const createdAt = toDate(message.createdAt);

  return (
    <div className="flex items-center gap-2.5 px-[18px] py-[3px] font-sans text-[12.5px] text-muted">
      <span className="grid w-8 shrink-0 place-items-center text-subtle [&_svg]:size-3.5">
        <IconInfoCircle aria-hidden />
      </span>
      <span className="min-w-0">{texts[message.kind]()}</span>
      <MessageTime date={createdAt} />
    </div>
  );
};
