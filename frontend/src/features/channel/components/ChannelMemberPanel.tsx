import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton";
import { useChannelMembers } from "#/features/channel/hooks/useChannelMembers";
import { MemberRow } from "#/features/member/components/MemberRow";

type ChannelMemberPanelProps = {
  channelId: string;
};

export const ChannelMemberPanel = ({ channelId }: ChannelMemberPanelProps) => {
  const { t } = useTranslation();
  const { data: members, isLoading, error } = useChannelMembers(channelId);

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

  return (
    <ul className="m-0 flex h-full list-none flex-col overflow-y-auto bg-surface p-1.5">
      {members.map((member) => (
        <li key={member.userId}>
          <MemberRow
            userId={member.userId}
            name={member.displayName}
            avatarUrl={member.avatarUrl}
            detail={member.email}
          />
        </li>
      ))}
    </ul>
  );
};
