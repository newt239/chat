import { IconPlayerPlayFilled } from "@tabler/icons-react";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles/styles";
import { formatDuration } from "#/lib/formatDuration";

import type { OgpData, YouTubeVideo } from "#/gen/chat/v1/message_pb";

type YouTubeCardProps = {
  url: string;
  ogp: OgpData;
  video: YouTubeVideo;
};

// 埋め込みの iframe は読み込まず、サムネイルを押したら YouTube を新しいタブで開く
export const YouTubeCard = ({ url, ogp, video }: YouTubeCardProps) => {
  const { t } = useTranslation();
  const thumbnail = ogp.imageUrl ?? `https://i.ytimg.com/vi/${video.videoId}/hqdefault.jpg`;

  return (
    <div className="w-105 max-w-full overflow-hidden rounded-lg border border-border bg-surface font-sans">
      <Link
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        aria-label={t("link.youtube.play")}
        className={`relative block aspect-video w-full bg-media ${focusRing}`}
      >
        <img src={thumbnail} alt="" className="block size-full object-cover" />
        <span className="absolute inset-0 m-auto grid h-10 w-14 place-items-center rounded-xl bg-media/72 text-media-fg [&_svg]:size-5">
          <IconPlayerPlayFilled aria-hidden />
        </span>
        {video.durationSeconds !== undefined && (
          <span className="absolute right-2 bottom-2 rounded-sm bg-media/75 px-1.25 py-px font-mono text-caption font-medium text-media-fg">
            {formatDuration(video.durationSeconds)}
          </span>
        )}
      </Link>
      <div className="flex flex-col gap-0.5 px-3 pt-2 pb-2.5">
        <span className="flex items-center gap-1.5 text-caption text-muted">
          <i aria-hidden className="block size-3.5 shrink-0 rounded-sm bg-danger" />
          <span className="truncate">
            {[t("link.youtube.site"), video.channelName].filter(Boolean).join(" · ")}
          </span>
        </span>
        <Link
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className={`cursor-pointer rounded-sm text-body-sm leading-normal font-semibold text-accent-text no-underline data-hovered:underline ${focusRing}`}
        >
          {ogp.title ?? url}
        </Link>
      </div>
    </div>
  );
};
