import { IconPlayerPlayFilled } from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles/styles";
import { PlayPauseButton } from "#/features/player/components/PlayPauseButton";
import { SeekBar } from "#/features/player/components/SeekBar";
import { SpeedButton } from "#/features/player/components/SpeedButton";
import { VideoSlot } from "#/features/player/components/VideoSlot";
import { formatDuration } from "#/features/player/utils/formatDuration";

import { useMediaControls } from "../hooks/useMediaControls";
import { AttachmentImage } from "./AttachmentImage";

import type { Message, MessageAttachment } from "#/gen/chat/v1/message_pb";

type VideoAttachmentProps = {
  attachment: MessageAttachment;
  message: Message;
};

const MAX_WIDTH = 420;
const MAX_HEIGHT = 320;

export const VideoAttachment = ({ attachment, message }: VideoAttachmentProps) => {
  const { t } = useTranslation();
  const {
    duration,
    handleCycleRate,
    handleSeek,
    handleToggle,
    inlineRef,
    isActive,
    isPlaying,
    position,
    rate,
  } = useMediaControls(attachment, message, "video");
  const { width, height, thumbnail } = attachment.media ?? {};
  const ratio = width && height ? width / height : 16 / 9;

  return (
    <div
      ref={inlineRef}
      className="max-w-full overflow-hidden rounded-[10px] border border-border bg-surface font-sans text-muted"
      style={{ width: Math.min(MAX_WIDTH, Math.round(MAX_HEIGHT * ratio)) }}
    >
      <div className="relative w-full bg-media" style={{ aspectRatio: ratio }}>
        {isActive ? (
          <VideoSlot kind="inline" />
        ) : (
          thumbnail && (
            <AttachmentImage
              attachmentId={attachment.id}
              thumbnail
              alt=""
              className="absolute inset-0 size-full object-cover"
            />
          )
        )}
        <Button
          aria-label={
            isPlaying
              ? t("attachment.player.pause")
              : t("attachment.player.playFile", { name: attachment.fileName })
          }
          onPress={handleToggle}
          className={`absolute inset-0 grid cursor-pointer place-items-center ${focusRing}`}
        >
          {!isActive && (
            <>
              <span className="absolute inset-x-0 bottom-0 h-12 bg-linear-to-t from-media/72 to-transparent" />
              <span className="absolute bottom-2.5 left-3 max-w-[calc(100%-24px)] truncate text-[13px] text-media-fg">
                {attachment.fileName}
              </span>
            </>
          )}
          {!isPlaying && (
            <span className="grid h-10 w-14 place-items-center rounded-xl bg-media/72 text-media-fg [&_svg]:size-5">
              <IconPlayerPlayFilled aria-hidden />
            </span>
          )}
        </Button>
      </div>
      <div className="flex items-center gap-2 border-t border-border py-[5px] pr-2.5 pl-1.5">
        <PlayPauseButton
          isPlaying={isPlaying}
          onPress={handleToggle}
          className="size-7 rounded-md text-text data-hovered:bg-hover [&_svg]:size-3.5"
        />
        <SeekBar position={position} duration={duration} onSeek={handleSeek} track="bar" />
        <span className="font-mono text-[11px] whitespace-nowrap tabular-nums">
          {formatDuration(position)} / {formatDuration(duration)}
        </span>
        <SpeedButton rate={rate} onPress={handleCycleRate} />
      </div>
    </div>
  );
};
