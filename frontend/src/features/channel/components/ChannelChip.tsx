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
  const name = useChannelById(workspaceId, channelId)?.name;
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
