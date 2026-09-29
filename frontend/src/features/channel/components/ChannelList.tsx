import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";

import { useChannels } from "../hooks/useChannel";
import { buildChannelTree } from "../utils/channelTree";
import { ChannelTreeItem } from "./ChannelTreeItem";

type ChannelListProps = {
  workspaceId: string;
};

// 参加中のチャンネルと、ツリーをつなぐための未参加の祖先を階層のツリーで並べる
export const ChannelList = ({ workspaceId }: ChannelListProps) => {
  const { t } = useTranslation();
  const { data: channels, isLoading } = useChannels(workspaceId);

  if (isLoading) {
    return <Skeleton className="mx-2 my-1 h-4 w-32 bg-(--nav-hover)" />;
  }

  const tree = buildChannelTree(channels ?? []);
  if (tree.length === 0) {
    return (
      <p className="m-0 px-2 py-1 text-caption text-(--nav-muted)">
        {t("shell.sidebar.noChannels")}
      </p>
    );
  }

  return tree.map((node, index) => (
    <ChannelTreeItem
      key={node.channel.id}
      workspaceId={workspaceId}
      node={node}
      depth={0}
      isLast={index === tree.length - 1}
    />
  ));
};
