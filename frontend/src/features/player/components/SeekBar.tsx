import { Slider, SliderThumb, SliderTrack } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn } from "#/components/ui/styles/styles";

import { Waveform } from "./Waveform";

type SeekBarProps = {
  position: number;
  duration: number;
  onSeek: (position: number) => void;
  // 細いバー（サイドバー用の配色あり）か、音声の波形
  track: "bar" | "sidebar" | { waveformSeed: string };
};

const barTones = {
  bar: "bg-border-strong",
  sidebar: "bg-side-muted/35",
};

// 再生位置のスライダー。矢印キーでも動かせる
export const SeekBar = ({ position, duration, onSeek, track }: SeekBarProps) => {
  const { t } = useTranslation();
  const max = Math.max(duration, 1);
  const progress = Math.min(position / max, 1);

  return (
    <Slider
      aria-label={t("attachment.player.seek")}
      minValue={0}
      maxValue={max}
      step={1}
      value={Math.min(position, max)}
      onChange={onSeek}
      className="min-w-10 flex-1"
    >
      <SliderTrack className="relative flex h-4 cursor-pointer items-center">
        {typeof track === "object" ? (
          <Waveform seed={track.waveformSeed} progress={progress} />
        ) : (
          <span className={cn("relative h-1 w-full overflow-hidden rounded-xs", barTones[track])}>
            <span
              className="absolute inset-y-0 left-0 bg-accent"
              style={{ width: `${progress * 100}%` }}
            />
          </span>
        )}
        <SliderThumb className="top-1/2 size-2.5 rounded-full bg-accent opacity-0 outline-none data-dragging:opacity-100 data-focus-visible:opacity-100 data-focus-visible:ring-2 data-focus-visible:ring-focus" />
      </SliderTrack>
    </Slider>
  );
};
