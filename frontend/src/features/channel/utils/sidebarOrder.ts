import { buildChannelTree, categoryOfChannel, sortChannelsByActivity } from "./channelTree";

import type { ChannelTreeNode } from "./channelTree";

import type { ChannelCategory } from "#/gen/chat/v1/channel_category_service_pb";
import type { Channel } from "#/gen/chat/v1/channel_service_pb";
import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";
import type { Preferences } from "#/providers/store/preferences";

const flatten = (nodes: ChannelTreeNode[]): Channel[] =>
  nodes.flatMap((node) => [node.channel, ...flatten(node.children)]);

// ミュート中のチャンネルはメンションがあるときだけ未読とみなす
const channelItem = (channel: Channel) => ({
  id: channel.id,
  isUnread: channel.mentionCount > 0 || (!channel.isMuted && channel.unreadCount > 0),
});

const dmItem = (dm: DirectMessage) => ({ id: dm.id, isUnread: !dm.isMuted && dm.unreadCount > 0 });

type SidebarOrderInput = {
  channels: readonly Channel[];
  dms: readonly DirectMessage[];
  categories: readonly ChannelCategory[];
  order: Preferences["channelSortOrder"];
};

// サイドバーの上から順（スター → カテゴリ → チャンネル → DM）に会話を並べ、未読かを添える。2 か所に出る会話は先の位置だけ残す
export const sidebarOrder = ({ channels, dms, categories, order }: SidebarOrderInput) => {
  const inCategory = (categoryId: string | null) => {
    const listed = channels.filter(
      (channel) => categoryOfChannel(channel, channels, categories) === categoryId,
    );
    return order === "recentActivity"
      ? sortChannelsByActivity(listed)
      : flatten(buildChannelTree(listed));
  };
  const all = [
    ...channels
      .filter((channel) => channel.isStarred && channel.isMember)
      .map((channel) => channelItem(channel)),
    ...dms.filter((dm) => dm.isStarred).map((dm) => dmItem(dm)),
    ...[...categories.flatMap((category) => inCategory(category.id)), ...inCategory(null)].map(
      (channel) => channelItem(channel),
    ),
    ...dms.map((dm) => dmItem(dm)),
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
