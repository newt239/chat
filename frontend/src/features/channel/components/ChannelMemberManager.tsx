import { useState } from "react";

import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { useChannelMemberActions } from "#/features/channel/hooks/useChannelMemberActions";
import { MemberPickerForm } from "#/features/member/components/MemberPickerForm";
import { ChannelRole } from "#/gen/chat/v1/channel_member_service_pb";
import { myUserIdAtom } from "#/providers/store/auth";

import type { ChannelMember } from "#/gen/chat/v1/channel_member_service_pb";

type ChannelMemberManagerProps = {
  channelId: string;
  workspaceId: string;
  members: ChannelMember[];
};

// メンバーパネルの下に置く、参加・退出と招待の操作
export const ChannelMemberManager = ({
  channelId,
  workspaceId,
  members,
}: ChannelMemberManagerProps) => {
  const { t } = useTranslation();
  const myId = useAtomValue(myUserIdAtom);
  const { invite, join, leave } = useChannelMemberActions(workspaceId);
  const [isLeaving, setIsLeaving] = useState(false);

  const memberIds = new Set(members.map((member) => member.userId));
  const isJoined = myId !== null && memberIds.has(myId);

  return (
    <section className="flex flex-col gap-2 border-t border-border px-2 pt-3 pb-2">
      <MemberPickerForm
        workspaceId={workspaceId}
        memberIds={memberIds}
        label={t("channel.members.invite")}
        placeholder={t("channel.members.invitePlaceholder")}
        submitLabel={t("channel.members.inviteSubmit")}
        isPending={invite.isPending}
        onSubmit={(userId) => {
          invite.mutate({ channelId, role: ChannelRole.MEMBER, userId });
        }}
      />
      {invite.isError && <p className="m-0 text-caption text-danger">{invite.error.message}</p>}
      {isJoined ? (
        <Button
          size="sm"
          variant="ghost"
          className="self-start text-danger"
          onPress={() => {
            setIsLeaving(true);
          }}
        >
          {t("channel.members.leave")}
        </Button>
      ) : (
        <Button
          size="sm"
          variant="secondary"
          className="self-start"
          isPending={join.isPending}
          onPress={() => {
            join.mutate({ channelId });
          }}
        >
          {t("channel.members.join")}
        </Button>
      )}
      <AlertDialog
        isOpen={isLeaving}
        onOpenChange={setIsLeaving}
        title={t("channel.members.leaveTitle")}
        confirmLabel={t("channel.members.leave")}
        tone="danger"
        isPending={leave.isPending}
        onConfirm={() => {
          leave.mutate(
            { channelId },
            {
              onSettled: () => {
                setIsLeaving(false);
              },
            },
          );
        }}
      >
        <p className="m-0">{t("channel.members.leaveBody")}</p>
      </AlertDialog>
    </section>
  );
};
