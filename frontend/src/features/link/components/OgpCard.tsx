import { IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";

import type { OgpData } from "#/gen/chat/v1/message_pb";

type OgpCardProps = {
  url: string;
  ogp: OgpData;
  // 投稿前のプレビューだけが外せる
  onRemove?: () => void;
};

export const OgpCard = ({ url, ogp, onRemove }: OgpCardProps) => {
  const { t } = useTranslation();

  return (
    <div className="relative w-105 max-w-full overflow-hidden rounded-lg border border-border bg-surface font-sans">
      {onRemove && (
        <IconButton
          label={t("link.remove")}
          onPress={onRemove}
          className="absolute top-1.5 right-1.5 z-1 size-6 border border-border bg-raised [&_svg]:size-3.5"
        >
          <IconX />
        </IconButton>
      )}
      <div className="flex flex-col gap-0.5 px-3 pt-2 pb-2.5">
        <span className="truncate text-caption text-muted">
          {ogp.siteName || new URL(url).hostname}
        </span>
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-body-sm leading-normal font-semibold text-accent-text no-underline"
        >
          {ogp.title || url}
        </a>
        {ogp.description && (
          <p className="m-0 line-clamp-2 text-label font-normal text-muted">{ogp.description}</p>
        )}
      </div>
      {ogp.imageUrl && (
        <img
          src={ogp.imageUrl}
          alt=""
          className="block aspect-1200/630 w-full border-t border-border object-cover"
        />
      )}
    </div>
  );
};
