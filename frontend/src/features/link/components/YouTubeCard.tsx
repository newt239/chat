import { IconPlayerPlayFilled } from "@tabler/icons-react";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";
import { formatDuration } from "#/features/player/utils/formatDuration";

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
    <div className="w-[min(420px,100%)] overflow-hidden rounded-[10px] border border-border bg-surface font-sans">
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
          <span className="absolute right-2 bottom-2 rounded-[4px] bg-media/75 px-[5px] py-px font-mono text-[11px] font-medium text-media-fg">
            {formatDuration(video.durationSeconds)}
          </span>
        )}
      </Link>
      <div className="flex flex-col gap-0.5 px-3 pt-2 pb-2.5">
        <span className="flex items-center gap-1.5 text-[11.5px] text-muted">
          <i aria-hidden className="block size-3.5 shrink-0 rounded-[3px] bg-danger" />
          <span className="truncate">
            {[t("link.youtube.site"), video.channelName].filter(Boolean).join(" · ")}
          </span>
        </span>
        <Link
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className={`cursor-pointer rounded-sm text-[13.5px] leading-[1.45] font-semibold text-accent-text no-underline data-hovered:underline ${focusRing}`}
        >
          {ogp.title ?? url}
        </Link>
      </div>
    </div>
  );
};
