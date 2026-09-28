import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { useChannels } from "../hooks/useChannel";
import { relativePath } from "../utils/channelTree";

type ChannelChipProps = {
  workspaceId: string;
  // 集約表示している親チャンネルのパス
  parentName: string;
  channelId: string;
};

// 集約表示でメッセージの投稿先を示すチップ。押すとそのチャンネルを開く
export const ChannelChip = ({ workspaceId, parentName, channelId }: ChannelChipProps) => {
  const { t } = useTranslation();
  const { data: channels } = useChannels(workspaceId);
  const listed = channels?.find((channel) => channel.id === channelId);
  // 未参加の公開チャンネルは一覧にないため個別に取得する
  const { data: fetched } = useQuery(
    ChannelService.method.getChannel,
    channels !== undefined && listed === undefined ? { channelId } : skipToken,
    { select: (res) => res.channel },
  );
  const name = (listed ?? fetched)?.name;
  if (name === undefined) {
    return null;
  }

  const label = relativePath(parentName, name);
  return (
    <Link
      to="/app/$workspaceId/$channelId"
      params={{ channelId, workspaceId }}
      aria-label={t("channel.aggregate.open", { name: label })}
      className="rounded-[5px] border border-border px-[5px] text-[11.5px] leading-[17px] font-medium text-muted no-underline data-hovered:border-accent data-hovered:text-accent-text"
    >
      # {label}
    </Link>
  );
};
