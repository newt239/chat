import { useQuery } from "@connectrpc/connect-query";
import { IconLink } from "@tabler/icons-react";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";
import { ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";

type ChannelLinkBarProps = {
  channelId: string;
};

// チャンネルヘッダーの下に並べる関連リンク。編集は #13 でチャンネル情報から行う
export const ChannelLinkBar = ({ channelId }: ChannelLinkBarProps) => {
  const { t } = useTranslation();
  const { data: links = [] } = useQuery(
    ChannelLinkService.method.listChannelLinks,
    { channelId },
    { select: (res) => res.links },
  );

  if (links.length === 0) {
    return null;
  }

  return (
    <nav
      aria-label={t("shell.channel.links")}
      className="flex h-8 shrink-0 items-center gap-0.5 overflow-x-auto border-b border-border pr-2.5 pl-3.5 [scrollbar-width:none]"
    >
      {links.map((link) => (
        <Link
          key={link.id}
          href={link.url}
          target="_blank"
          rel="noopener noreferrer"
          className={`inline-flex h-6 shrink-0 cursor-pointer items-center gap-1.5 rounded-[6px] px-2 text-[12.5px] whitespace-nowrap text-text no-underline data-hovered:bg-hover [&_svg]:size-3.5 [&_svg]:text-muted ${focusRing}`}
        >
          <IconLink aria-hidden />
          {link.title}
        </Link>
      ))}
    </nav>
  );
};
