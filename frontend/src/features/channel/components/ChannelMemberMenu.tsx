import { useState } from "react";

import { IconCheck, IconDots, IconUserMinus } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { useChannelMemberActions } from "#/features/channel/hooks/useChannelMemberActions";
import { ChannelRole } from "#/gen/chat/v1/channel_member_service_pb";

import type { ChannelMember } from "#/gen/chat/v1/channel_member_service_pb";

const ROLES = [ChannelRole.MEMBER, ChannelRole.ADMIN];

type ChannelMemberMenuProps = {
  workspaceId: string;
  channelId: string;
  member: ChannelMember;
  name: string;
};

export const ChannelMemberMenu = ({
  workspaceId,
  channelId,
  member,
  name,
}: ChannelMemberMenuProps) => {
  const { t } = useTranslation();
  const { remove, updateRole } = useChannelMemberActions(workspaceId);
  const [isRemoving, setIsRemoving] = useState(false);

  return (
    <>
      <Menu
        trigger={
          <IconButton label={t("channel.members.menu", { name })}>
            <IconDots />
          </IconButton>
        }
      >
        {ROLES.map((role) => (
          <MenuItem
            key={role}
            icon={member.role === role ? <IconCheck aria-hidden /> : <span className="size-4" />}
            onAction={() => {
              updateRole.mutate({ channelId, role, userId: member.userId });
            }}
          >
            {t(role === ChannelRole.ADMIN ? "channel.roles.admin" : "channel.roles.member")}
          </MenuItem>
        ))}
        <MenuSeparator />
        <MenuItem
          tone="danger"
          icon={<IconUserMinus aria-hidden />}
          onAction={() => {
            setIsRemoving(true);
          }}
        >
          {t("channel.members.remove")}
        </MenuItem>
      </Menu>
      <AlertDialog
        isOpen={isRemoving}
        onOpenChange={setIsRemoving}
        title={t("channel.members.removeTitle", { name })}
        confirmLabel={t("channel.members.remove")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          remove.mutate(
            { channelId, userId: member.userId },
            {
              onSettled: () => {
                setIsRemoving(false);
              },
            },
          );
        }}
      >
        <p className="m-0">{t("channel.members.removeBody")}</p>
      </AlertDialog>
    </>
  );
};
