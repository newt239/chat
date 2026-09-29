import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { useChannelMembers } from "#/features/channel/hooks/useChannelMembers";
import { MemberRow } from "#/features/member/components/MemberRow";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { channelViewersAtom } from "#/providers/store/ui";

import type { ChannelMember } from "#/gen/chat/v1/channel_member_service_pb";

type ChannelMemberPanelProps = {
  channelId: string;
};

// 参加者を「いま閲覧中」と「その他のメンバー」に分けて並べる
export const ChannelMemberPanel = ({ channelId }: ChannelMemberPanelProps) => {
  const { t } = useTranslation();
  const { data: members, isLoading, error } = useChannelMembers(channelId);
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

  if (error || members === undefined || members.length === 0) {
    return (
      <p className="m-0 p-4 text-caption text-muted">
        {error ? t("channel.members.loadFailed") : t("channel.members.empty")}
      </p>
    );
  }

  const viewers = members.filter((member) => viewerIds.includes(member.userId));
  const others = members.filter((member) => !viewerIds.includes(member.userId));
  const section = (
    title: string,
    list: ChannelMember[],
    detail: (member: ChannelMember) => string,
  ) => (
    <section aria-label={title} className="flex flex-col">
      <h4 className="m-0 px-2 pt-2.5 pb-1 text-[11.5px] font-semibold text-muted">
        {title} · {list.length}
      </h4>
      <ul className="m-0 flex list-none flex-col p-0">
        {list.map((member) => (
          <li key={member.userId}>
            <MemberRow
              userId={member.userId}
              name={displayName(member.userId, member.displayName)}
              avatarUrl={member.avatarUrl}
              detail={detail(member)}
            />
          </li>
        ))}
      </ul>
    </section>
  );

  return (
    <div className="flex h-full flex-col overflow-y-auto bg-surface p-1.5">
      {section(t("channel.members.viewing"), viewers, () => t("channel.members.viewingNow"))}
      {section(t("channel.members.others"), others, (member) => member.email)}
    </div>
  );
};
