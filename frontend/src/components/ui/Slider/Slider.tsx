import {
  Slider as AriaSlider,
  Label,
  SliderOutput,
  SliderThumb,
  SliderTrack,
} from "react-aria-components";

import { cn, fieldStyles } from "#/components/ui/styles/styles";

type SliderProps = {
  label: string;
  value: number;
  minValue: number;
  maxValue: number;
  step: number;
  onChange: (value: number) => void;
  // ドラッグを終えたときの値。保存など重い処理はこちらで行う
  onChangeEnd: (value: number) => void;
  // 色相のようにトラック自体で値の意味を示すときの背景（CSS の background）
  trackBackground?: string;
  className?: string;
};

export const Slider = ({
  label,
  value,
  minValue,
  maxValue,
  step,
  onChange,
  onChangeEnd,
  trackBackground,
  className,
}: SliderProps) => (
  <AriaSlider
    value={value}
    minValue={minValue}
    maxValue={maxValue}
    step={step}
    onChange={onChange}
    onChangeEnd={onChangeEnd}
    className={cn("grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-1 font-sans", className)}
  >
    <Label className={fieldStyles.label}>{label}</Label>
    <SliderOutput className="text-right font-mono text-xs text-muted tabular-nums" />
    <SliderTrack className="relative col-span-2 h-6 w-full">
      {({ state }) => (
        <>
          <span
            className="absolute inset-x-0 top-1/2 h-2.5 -translate-y-1/2 rounded-full bg-border-strong"
            style={trackBackground ? { background: trackBackground } : undefined}
          />
          {!trackBackground && (
            <span
              className="absolute top-1/2 left-0 h-2.5 -translate-y-1/2 rounded-full bg-accent"
              style={{ width: `${state.getThumbPercent(0) * 100}%` }}
            />
          )}
          <SliderThumb className="top-1/2 size-4.5 cursor-grab rounded-full border-2 border-text bg-surface shadow-md data-dragging:cursor-grabbing data-focus-visible:outline-2 data-focus-visible:outline-offset-2 data-focus-visible:outline-focus data-focus-visible:outline-solid" />
        </>
      )}
    </SliderTrack>
  </AriaSlider>
);
