import { PlayPauseButton } from "#/features/player/components/PlayPauseButton";
import { SeekBar } from "#/features/player/components/SeekBar";
import { SpeedButton } from "#/features/player/components/SpeedButton";
import { formatDuration } from "#/features/player/utils/formatDuration";

import { useMediaControls } from "../hooks/useMediaControls";

import type { Message, MessageAttachment } from "#/gen/chat/v1/message_pb";

type AudioAttachmentProps = {
  attachment: MessageAttachment;
  message: Message;
};

export const AudioAttachment = ({ attachment, message }: AudioAttachmentProps) => {
  const {
    duration,
    handleCycleRate,
    handleSeek,
    handleToggle,
    inlineRef,
    isPlaying,
    position,
    rate,
  } = useMediaControls(attachment, message, "audio");

  return (
    <div
      ref={inlineRef}
      className="flex w-100 max-w-full items-center gap-2.5 rounded-xl border border-border bg-surface py-1.75 pr-2.5 pl-1.75 font-sans text-muted"
    >
      <PlayPauseButton
        isPlaying={isPlaying}
        onPress={handleToggle}
        className="size-8.5 rounded-full bg-accent text-accent-fg data-hovered:bg-accent-hover [&_svg]:size-3.75"
      />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <b className="truncate text-label font-semibold text-text">{attachment.fileName}</b>
        <SeekBar
          position={position}
          duration={duration}
          onSeek={handleSeek}
          track={{ waveformSeed: attachment.id }}
        />
      </div>
      <span className="font-mono text-caption whitespace-nowrap tabular-nums">
        {formatDuration(position)} / {formatDuration(duration)}
      </span>
      <SpeedButton rate={rate} onPress={handleCycleRate} />
    </div>
  );
};
