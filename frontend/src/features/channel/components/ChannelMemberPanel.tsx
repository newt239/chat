import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { channelViewersAtom } from "#/features/channel/atoms";
import { ChannelMemberManager } from "#/features/channel/components/ChannelMemberManager";
import { ChannelMemberMenu } from "#/features/channel/components/ChannelMemberMenu";
import { useChannelMembers } from "#/features/channel/hooks/useChannelMembers";
import { useDMs } from "#/features/channel/hooks/useDM";
import { MemberRow } from "#/features/member/components/MemberRow";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";

import type { ChannelMember } from "#/gen/chat/v1/channel_member_service_pb";

type ChannelMemberPanelProps = {
  workspaceId: string;
  channelId: string;
};

// 参加者を「いま閲覧中」と「その他のメンバー」に分けて並べる
export const ChannelMemberPanel = ({ workspaceId, channelId }: ChannelMemberPanelProps) => {
  const { t } = useTranslation();
  const { data: members, isLoading, error } = useChannelMembers(channelId);
  const { data: dms } = useDMs(workspaceId);
  // DM には招待やロールがないので、管理の操作はチャンネルだけに出す
  const canManage = dms !== undefined && !dms.some((dm) => dm.id === channelId);
  const viewerIds = useAtomValue(channelViewersAtom)[channelId] ?? [];
  const displayName = useDisplayName();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-3">
        {[0, 1, 2].map((index) => (
          <div key={index} className="flex items-center gap-2.5">
            <Skeleton className="size-8 rounded-md" />
            <Skeleton className="h-4 w-32" />
          </div>
        ))}
      </div>
    );
  }

  if (error || members === undefined) {
    return <p className="m-0 p-4 text-caption text-muted">{t("channel.members.loadFailed")}</p>;
  }

  const viewers = members.filter((member) => viewerIds.includes(member.userId));
  const others = members.filter((member) => !viewerIds.includes(member.userId));
  const renderSection = (
    title: string,
    list: ChannelMember[],
    detail: (member: ChannelMember) => string,
  ) => (
    <section aria-label={title} className="flex flex-col">
      <h4 className="m-0 px-2 pt-2.5 pb-1 text-caption font-semibold text-muted">
        {title} · {list.length}
      </h4>
      <ul className="m-0 flex list-none flex-col p-0">
        {list.map((member) => (
          <li key={member.userId} className="flex items-center gap-1">
            <div className="min-w-0 flex-1">
              <MemberRow
                userId={member.userId}
                name={displayName(member.userId, member.displayName)}
                avatarUrl={member.avatarUrl}
                detail={detail(member)}
              />
            </div>
            {canManage && (
              <ChannelMemberMenu
                workspaceId={workspaceId}
                channelId={channelId}
                member={member}
                name={displayName(member.userId, member.displayName)}
              />
            )}
          </li>
        ))}
      </ul>
    </section>
  );

  return (
    <div className="flex min-h-full flex-col bg-surface p-1.5">
      {members.length === 0 ? (
        <p className="m-0 p-2.5 text-caption text-muted">{t("channel.members.empty")}</p>
      ) : (
        <>
          {renderSection(t("channel.members.viewing"), viewers, () =>
            t("channel.members.viewingNow"),
          )}
          {renderSection(t("channel.members.others"), others, (member) => member.email)}
        </>
      )}
      {canManage && (
        <ChannelMemberManager channelId={channelId} workspaceId={workspaceId} members={members} />
      )}
    </div>
  );
};
