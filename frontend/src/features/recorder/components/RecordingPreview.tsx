import { useRef, useState } from "react";

import { PlayPauseButton } from "#/features/player/components/PlayPauseButton";
import { Waveform } from "#/features/player/components/Waveform";
import { formatDuration } from "#/features/player/utils/formatDuration";

type RecordingPreviewProps = {
  url: string;
  durationSeconds: number;
};

// 送る前に録音を聴き直す。アプリ全体のプレイヤーは使わず、その場の audio で鳴らす
export const RecordingPreview = ({ url, durationSeconds }: RecordingPreviewProps) => {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [position, setPosition] = useState(0);

  return (
    <span className="flex min-w-0 flex-1 items-center gap-2">
      {/* oxlint-disable-next-line jsx-a11y/media-has-caption -- 自分の録音を聴き直すだけで字幕はない */}
      <audio
        ref={audioRef}
        src={url}
        preload="metadata"
        className="hidden"
        onPlay={() => {
          setIsPlaying(true);
        }}
        onPause={() => {
          setIsPlaying(false);
        }}
        onTimeUpdate={(event) => {
          setPosition(event.currentTarget.currentTime);
        }}
        onEnded={() => {
          setPosition(0);
        }}
      />
      <PlayPauseButton
        isPlaying={isPlaying}
        onPress={() => {
          const audio = audioRef.current;
          if (audio === null) {
            return;
          }
          if (audio.paused) {
            void audio.play();
          } else {
            audio.pause();
          }
        }}
        className="size-7 rounded-full bg-accent max-md:size-11 text-accent-fg data-hovered:bg-accent-hover [&_svg]:size-3.5"
      />
      <span className="min-w-0 flex-1">
        <Waveform seed={url} progress={durationSeconds > 0 ? position / durationSeconds : 0} />
      </span>
      <span className="font-mono text-[11px] whitespace-nowrap text-muted tabular-nums">
        {formatDuration(position)} / {formatDuration(durationSeconds)}
      </span>
    </span>
  );
};
