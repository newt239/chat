import { cn } from "#/components/ui/styles/styles";

import { waveformBars } from "../utils/waveformBars";

type WaveformProps = {
  seed: string;
  progress: number;
};

// 実際の音量ではなく、添付ごとに決まる見た目だけの波形
export const Waveform = ({ seed, progress }: WaveformProps) => {
  const bars = waveformBars(seed);
  return (
    <span aria-hidden className="flex h-5.5 w-full items-center gap-0.5">
      {bars.map((height, index) => (
        <i
          // oxlint-disable-next-line react/no-array-index-key -- 並びが変わらない固定長の配列
          key={index}
          className={cn(
            "min-h-0.75 flex-1 rounded-xs",
            index / bars.length < progress ? "bg-accent" : "bg-border-strong",
          )}
          style={{ height: `${height * 100}%` }}
        />
      ))}
    </span>
  );
};
