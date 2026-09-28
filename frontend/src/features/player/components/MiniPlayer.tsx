import { IconMusic, IconX } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { AnimatePresence, motion } from "motion/react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { lastSegment } from "#/features/channel/utils/channelPath";
import { transitions } from "#/lib/motion";
import { useIsMobile } from "#/lib/useMediaQuery";

import { usePlayerState } from "../hooks/usePlayerState";
import { mediaPlayer } from "../mediaPlayer";
import { formatDuration } from "../utils/formatDuration";
import { PlayPauseButton } from "./PlayPauseButton";
import { SeekBar } from "./SeekBar";
import { SpeedButton } from "./SpeedButton";
import { VideoSlot } from "./VideoSlot";

type MiniPlayerProps = {
  // sidebar: デスクトップのサイドバー下部、mobile: 画面下部の帯
  variant: "sidebar" | "mobile";
};

const variants = {
  mobile: {
    button: "size-[38px] rounded-md text-text data-hovered:bg-hover [&_svg]:size-[17px]",
    root: "relative shrink-0 border-t border-border bg-raised text-text",
    source: "text-xs text-muted",
    title: "text-[13.5px] text-text",
  },
  sidebar: {
    button: "size-7 rounded-md text-side-strong data-hovered:bg-side-active/40 [&_svg]:size-3.5",
    root: "mx-2 mb-1.5 shrink-0 overflow-hidden rounded-[10px] border border-side-hover bg-side-hover text-side-fg",
    source: "text-[11px] text-side-muted",
    title: "text-[12.5px] text-side-strong",
  },
};

// 再生中の添付が画面に見えていないときに出す。元のメッセージへの移動もここから行う
export const MiniPlayer = ({ variant }: MiniPlayerProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { track, isPlaying, isInlineVisible, position, duration, rate } = usePlayerState();
  const { data: channels } = useChannels(track?.workspaceId ?? null);
  const styles = variants[variant];
  // 共有の <video> の枠が 2 つできないよう、画面幅に合う方だけを描画する
  const isShown = useIsMobile() === (variant === "mobile");

  const channel = channels?.find(({ id }) => id === track?.channelId);
  const isVideo = track?.kind === "video";

  const handleJump = () => {
    if (track === null || mediaPlayer.scrollToSource()) {
      return;
    }
    const params = { channelId: track.channelId, workspaceId: track.workspaceId };
    // スレッドの返信はスレッドを開いてから返信へ移動する
    void (track.parentId === undefined
      ? navigate({
          params,
          search: { message: track.messageId },
          to: "/app/$workspaceId/$channelId",
        })
      : navigate({
          params: { ...params, messageId: track.parentId },
          search: { message: track.messageId },
          to: "/app/$workspaceId/$channelId/thread/$messageId",
        }));
  };

  const handleTogglePlay = () => {
    if (isPlaying) {
      mediaPlayer.pause();
    } else {
      void mediaPlayer.resume();
    }
  };

  const { stop: handleStop, seek: handleSeek, cycleRate: handleCycleRate } = mediaPlayer;
  const thumbnail = isVideo ? <VideoSlot kind="mini" /> : <IconMusic aria-hidden />;

  return (
    <AnimatePresence>
      {isShown && track !== null && !isInlineVisible && (
        <motion.section
          aria-label={t("attachment.player.label")}
          initial={{ opacity: 0, y: variant === "mobile" ? 24 : 16 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: variant === "mobile" ? 24 : 16 }}
          transition={transitions.spring}
          className={cn("font-sans", styles.root)}
        >
          {variant === "mobile" && (
            <span
              className="absolute -top-px left-0 h-0.5 bg-accent"
              style={{ width: `${Math.min(position / Math.max(duration, 1), 1) * 100}%` }}
            />
          )}
          {variant === "sidebar" && isVideo && (
            <div className="relative aspect-video bg-media">{thumbnail}</div>
          )}
          <div
            className={cn(
              "flex items-center",
              variant === "mobile" ? "gap-2 py-2 pr-1.5 pl-3" : "gap-0.5 pt-1.5 pr-1 pl-2.5",
            )}
          >
            {variant === "mobile" ? (
              <span className="relative grid h-8 w-14 shrink-0 place-items-center overflow-hidden rounded-md bg-sunken text-muted [&_svg]:size-4">
                {thumbnail}
              </span>
            ) : (
              !isVideo && (
                <span className="mr-1.5 grid size-[26px] shrink-0 place-items-center rounded-[7px] bg-accent text-accent-fg [&_svg]:size-[15px]">
                  <IconMusic aria-hidden />
                </span>
              )
            )}
            <div className="flex min-w-0 flex-1 flex-col leading-[1.3]">
              <b className={cn("truncate font-semibold", styles.title)}>{track.fileName}</b>
              <Button
                onPress={handleJump}
                className={cn(
                  "cursor-pointer truncate rounded-sm text-left data-hovered:underline",
                  styles.source,
                  focusRing,
                )}
              >
                <span aria-hidden>
                  {channel ? `#${lastSegment(channel.name)} · ` : ""}
                  {track.authorName}
                </span>
                <span className="sr-only">{t("attachment.player.jump")}</span>
              </Button>
            </div>
            <PlayPauseButton
              isPlaying={isPlaying}
              onPress={handleTogglePlay}
              className={styles.button}
            />
            <Button
              aria-label={t("attachment.player.close")}
              onPress={handleStop}
              className={cn(
                "grid shrink-0 cursor-pointer place-items-center",
                styles.button,
                focusRing,
              )}
            >
              <IconX aria-hidden />
            </Button>
          </div>
          {variant === "sidebar" && (
            <div className="flex items-center gap-2 pr-2 pb-1.5 pl-2.5 font-mono text-[10.5px] text-side-muted tabular-nums">
              <SeekBar
                position={position}
                duration={duration}
                onSeek={handleSeek}
                track="sidebar"
              />
              <span>{formatDuration(position)}</span>
              <SpeedButton rate={rate} onPress={handleCycleRate} />
            </div>
          )}
        </motion.section>
      )}
    </AnimatePresence>
  );
};
