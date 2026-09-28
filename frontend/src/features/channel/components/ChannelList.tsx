import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton";

import { useChannels } from "../hooks/useChannel";
import { ChannelRow } from "./ChannelRow";

type ChannelListProps = {
  workspaceId: string;
};

// 参加中のチャンネルを名前順に並べる。階層のツリー表示は #13 でこの一覧を置き換える
export const ChannelList = ({ workspaceId }: ChannelListProps) => {
  const { t } = useTranslation();
  const { data: channels, isLoading } = useChannels(workspaceId);

  if (isLoading) {
    return <Skeleton className="mx-2 my-1 h-4 w-32 bg-(--nav-hover)" />;
  }

  const joined = (channels ?? [])
    .filter((channel) => channel.isMember)
    .toSorted((a, b) => a.name.localeCompare(b.name));

  if (joined.length === 0) {
    return (
      <p className="m-0 px-2 py-1 text-caption text-(--nav-muted)">
        {t("shell.sidebar.noChannels")}
      </p>
    );
  }

  return joined.map((channel) => (
    <ChannelRow key={channel.id} workspaceId={workspaceId} channel={channel} />
  ));
};
