import { useState } from "react";

import { IconCheck, IconDots, IconUserMinus } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Badge } from "#/components/ui/Badge/Badge";
import { Button } from "#/components/ui/Button/Button";
import { ComboBox } from "#/components/ui/ComboBox/ComboBox";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { useChannelMemberActions } from "#/features/channel/hooks/useChannelMemberActions";
import { useChannelMembers } from "#/features/channel/hooks/useChannelMembers";
import { channelRoleKeys } from "#/features/channel/utils/channelRole";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMembers } from "#/features/member/hooks/useMembers";
import { ChannelRole } from "#/gen/chat/v1/channel_member_service_pb";
import { myUserIdAtom } from "#/providers/store/auth";

const ROLES = [ChannelRole.MEMBER, ChannelRole.ADMIN];

type ChannelMemberManagerProps = {
  channelId: string;
  workspaceId: string;
};

export const ChannelMemberManager = ({ channelId, workspaceId }: ChannelMemberManagerProps) => {
  const { t } = useTranslation();
  const myId = useAtomValue(myUserIdAtom);
  const { data: channelMembers } = useChannelMembers(channelId);
  const { data: workspaceMembers } = useMembers(workspaceId);
  const { invite, join, leave, remove, updateRole } = useChannelMemberActions(workspaceId);
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const displayName = useDisplayName();

  const memberIds = new Set(channelMembers?.map((member) => member.userId));
  const isJoined = myId !== null && memberIds.has(myId);
  const inviteOptions = (workspaceMembers ?? [])
    .filter((member) => !memberIds.has(member.userId))
    .map((member) => ({ label: member.nickname ?? member.displayName, value: member.userId }));
  const failedAction = [remove, updateRole, leave, join].find((action) => action.isError);

  return (
    <section className="flex flex-col gap-2 border-b border-border px-4 py-3">
      <h4 className="m-0 flex items-center justify-between gap-2 text-xs font-semibold text-muted">
        {t("channel.members.title", { count: channelMembers?.length ?? 0 })}
        {isJoined ? (
          <Button
            size="sm"
            variant="ghost"
            className="text-danger"
            isPending={leave.isPending}
            onPress={() => {
              leave.mutate({ channelId });
            }}
          >
            {t("channel.members.leave")}
          </Button>
        ) : (
          <Button
            size="sm"
            variant="secondary"
            isPending={join.isPending}
            onPress={() => {
              join.mutate({ channelId });
            }}
          >
            {t("channel.members.join")}
          </Button>
        )}
      </h4>

      <ul className="m-0 -mx-2 flex list-none flex-col p-0">
        {channelMembers?.map((member) => {
          const name = displayName(member.userId, member.displayName);
          return (
            <li
              key={member.userId}
              className="flex items-center gap-2.5 rounded-md px-2 py-1 text-body-sm"
            >
              <Avatar name={member.displayName} src={member.avatarUrl} size={28} />
              <span className="min-w-0 flex-1 truncate">{name}</span>
              {member.role === ChannelRole.ADMIN && (
                <Badge tone="tag">{t("channel.roles.admin")}</Badge>
              )}
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
                    icon={
                      member.role === role ? <IconCheck aria-hidden /> : <span className="size-4" />
                    }
                    onAction={() => {
                      updateRole.mutate({ channelId, role, userId: member.userId });
                    }}
                  >
                    {t(channelRoleKeys[role])}
                  </MenuItem>
                ))}
                <MenuSeparator />
                <MenuItem
                  tone="danger"
                  icon={<IconUserMinus aria-hidden />}
                  onAction={() => {
                    remove.mutate({ channelId, userId: member.userId });
                  }}
                >
                  {t("channel.members.remove")}
                </MenuItem>
              </Menu>
            </li>
          );
        })}
      </ul>
      {channelMembers?.length === 0 && (
        <p className="m-0 text-caption text-muted">{t("channel.members.empty")}</p>
      )}

      <Form
        className="flex items-end gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          if (selectedUserId !== null) {
            invite.mutate({ channelId, role: ChannelRole.MEMBER, userId: selectedUserId });
            setSelectedUserId(null);
          }
        }}
      >
        <ComboBox
          label={t("channel.members.invite")}
          placeholder={t("channel.members.invitePlaceholder")}
          options={inviteOptions}
          value={selectedUserId}
          onChange={setSelectedUserId}
          className="flex-1"
        />
        <Button type="submit" isDisabled={selectedUserId === null} isPending={invite.isPending}>
          {t("channel.members.inviteSubmit")}
        </Button>
      </Form>

      {invite.isError && <p className="m-0 text-caption text-danger">{invite.error.message}</p>}
      {failedAction && (
        <p className="m-0 text-caption text-danger">{t("channel.members.actionFailed")}</p>
      )}
    </section>
  );
};
