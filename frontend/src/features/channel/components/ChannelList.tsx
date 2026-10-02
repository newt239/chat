import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { preferencesAtom } from "#/providers/store/preferences";

import { useChannels } from "../hooks/useChannel";
import { useChannelCategories } from "../hooks/useChannelCategories";
import { buildChannelTree, categoryOfChannel, sortChannelsByActivity } from "../utils/channelTree";
import { ChannelRow } from "./ChannelRow";
import { ChannelTreeItem } from "./ChannelTreeItem";

type ChannelListProps = {
  workspaceId: string;
  // null は自分で作ったカテゴリに入っていないチャンネル
  categoryId: string | null;
};

// カテゴリに入るチャンネルを、既定では階層のツリーで、新しいメッセージ順では階層を分けて並べる
export const ChannelList = ({ workspaceId, categoryId }: ChannelListProps) => {
  const { t } = useTranslation();
  const { channelSortOrder } = useAtomValue(preferencesAtom);
  const { data: channels, isLoading } = useChannels(workspaceId);
  const { data: categories } = useChannelCategories(workspaceId);

  if (isLoading) {
    return <Skeleton className="mx-2 my-1 h-4 w-32 bg-(--nav-hover)" />;
  }

  const all = channels ?? [];
  const categoryByChannel = new Map(
    (categories ?? []).flatMap((category) =>
      category.channelIds.map((channelId) => [channelId, category.id] as const),
    ),
  );
  const inCategory = all.filter(
    (channel) => categoryOfChannel(channel, all, categoryByChannel) === categoryId,
  );
  const sorted = channelSortOrder === "recentActivity" ? sortChannelsByActivity(inCategory) : [];
  const tree = channelSortOrder === "recentActivity" ? [] : buildChannelTree(inCategory);

  if (sorted.length === 0 && tree.length === 0) {
    return (
      categoryId === null && (
        <p className="m-0 px-2 py-1 text-caption text-(--nav-muted)">
          {t("shell.sidebar.noChannels")}
        </p>
      )
    );
  }

  return (
    <>
      {sorted.map((channel) => (
        <ChannelRow key={channel.id} workspaceId={workspaceId} channel={channel} />
      ))}
      {tree.map((node, index) => (
        <ChannelTreeItem
          key={node.channel.id}
          workspaceId={workspaceId}
          node={node}
          depth={0}
          isLast={index === tree.length - 1}
        />
      ))}
    </>
  );
};
