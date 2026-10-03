import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

type DMAvatarProps = {
  dm: DirectMessage;
  size: number;
};

export const DMAvatar = ({ dm, size }: DMAvatarProps) => {
  const { t } = useTranslation();
  const [member] = dm.members;
  if (dm.type !== DirectMessageType.GROUP_DM && member !== undefined) {
    return <Avatar name={member.displayName} src={member.avatarUrl} size={size} />;
  }
  // グループ DM は顔を並べる代わりに自分を含めた人数を表示する
  const count = dm.members.length + 1;
  return (
    <span
      role="img"
      aria-label={t("ui.avatar.groupMembers", { count })}
      className="inline-grid shrink-0 place-items-center rounded-[28%] border border-border bg-sunken font-mono leading-none font-bold text-muted select-none"
      style={{ fontSize: Math.max(9, Math.round(size * 0.42)), height: size, width: size }}
    >
      <span aria-hidden>{count}</span>
    </span>
  );
};
