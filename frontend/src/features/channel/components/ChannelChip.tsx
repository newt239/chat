import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";

import { useChannelById } from "../hooks/useChannelById";
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
  const name = useChannelById(workspaceId, channelId).channel?.name;
  if (name === undefined) {
    return null;
  }

  const label = relativePath(parentName, name);
  return (
    <Link
      to="/app/$workspaceId/$channelId"
      params={{ channelId, workspaceId }}
      aria-label={t("channel.aggregate.open", { name: label })}
      className="rounded-sm border border-border px-1.25 text-caption leading-4.25 font-medium text-muted no-underline data-hovered:border-accent data-hovered:text-accent-text"
    >
      # {label}
    </Link>
  );
};
