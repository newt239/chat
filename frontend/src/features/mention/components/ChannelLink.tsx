import type { ReactNode } from "react";

import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { cn } from "#/components/ui/styles/styles";

import { useMentionDirectory } from "../hooks/useMentionDirectory";

const chipClassName =
  "inline cursor-pointer rounded-sm bg-accent-soft px-0.75 font-semibold text-accent-text no-underline";

type ChannelLinkProps = {
  "data-channel": string;
  children?: ReactNode;
};

/** 本文に ID で埋め込んだチャンネルを今の名前で出す。見えないチャンネルは名前を伏せる */
export const ChannelLink = ({ "data-channel": channelId }: ChannelLinkProps) => {
  const { t } = useTranslation();
  const directory = useMentionDirectory();
  const channel = directory.channel(channelId);

  if (directory.workspaceId === null || channel === undefined) {
    return (
      <span className={cn(chipClassName, "cursor-default")}>
        #{t("message.mention.unknownChannel")}
      </span>
    );
  }

  return (
    <Link
      to="/app/$workspaceId/$channelId"
      params={{ channelId: channel.id, workspaceId: directory.workspaceId }}
      className={chipClassName}
    >
      #{channel.name}
    </Link>
  );
};
