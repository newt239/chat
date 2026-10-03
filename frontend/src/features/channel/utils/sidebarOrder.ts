import { buildChannelTree, categoryOfChannel, sortChannelsByActivity } from "./channelTree";

import type { ChannelTreeNode } from "./channelTree";

import type { ChannelCategory } from "#/gen/chat/v1/channel_category_service_pb";
import type { Channel } from "#/gen/chat/v1/channel_service_pb";
import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";
import type { Preferences } from "#/providers/store/preferences";

const flatten = (nodes: ChannelTreeNode[]): Channel[] =>
  nodes.flatMap((node) => [node.channel, ...flatten(node.children)]);

// サイドバーの上から順（スター → カテゴリ → チャンネル → DM）に会話を並べ、未読かを添える。2 か所に出る会話は先の位置だけ残す
// ミュート中のチャンネルはメンションがあるときだけ未読とみなす
export const sidebarOrder = (
  channels: readonly Channel[],
  dms: readonly DirectMessage[],
  categories: readonly ChannelCategory[],
  order: Preferences["channelSortOrder"],
) => {
  const inCategory = (categoryId: string | null) => {
    const listed = channels.filter(
      (channel) => categoryOfChannel(channel, channels, categories) === categoryId,
    );
    return order === "recentActivity"
      ? sortChannelsByActivity(listed)
      : flatten(buildChannelTree(listed));
  };
  const toItem = (channel: Channel) => ({
    id: channel.id,
    isUnread: channel.mentionCount > 0 || (!channel.isMuted && channel.unreadCount > 0),
  });
  const toDMItem = (dm: DirectMessage) => ({
    id: dm.id,
    isUnread: !dm.isMuted && dm.unreadCount > 0,
  });
  const all = [
    ...channels.filter((channel) => channel.isStarred && channel.isMember).map(toItem),
    ...dms.filter((dm) => dm.isStarred).map(toDMItem),
    ...categories.flatMap((category) => inCategory(category.id)).map(toItem),
    ...inCategory(null).map(toItem),
    ...dms.map(toDMItem),
  ];
  return all.filter((item, index) => all.findIndex(({ id }) => id === item.id) === index);
};

// 現在の会話より後ろにある未読の会話の ID。末尾まで無ければ先頭から探す
export const nextUnreadId = (
  items: ReturnType<typeof sidebarOrder>,
  currentId: string | undefined,
) => {
  const index = items.findIndex((item) => item.id === currentId);
  return [...items.slice(index + 1), ...items.slice(0, Math.max(index, 0))].find(
    (item) => item.isUnread,
  )?.id;
};
