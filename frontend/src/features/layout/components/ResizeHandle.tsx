import { useRef, useState } from "react";

import { cn } from "#/components/ui/styles/styles";

type ResizeHandleProps = {
  label: string;
  // パネルのどちらの端に置くか。left なら左へ動かすほど広がる
  edge: "left" | "right";
  value: number;
  minValue: number;
  maxValue: number;
  // ダブルクリックと Enter で戻す幅
  defaultValue: number;
  onChange: (value: number) => void;
};

const keyStep = 16;

// パネルの端に置く幅変更のつまみ。ドラッグと ← → キーで動かす
export const ResizeHandle = ({
  label,
  edge,
  value,
  minValue,
  maxValue,
  defaultValue,
  onChange,
}: ResizeHandleProps) => {
  const start = useRef({ value, x: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const direction = edge === "right" ? 1 : -1;
  const change = (next: number) => {
    onChange(Math.min(maxValue, Math.max(minValue, Math.round(next))));
  };
  const keyActions: Record<string, () => void> = {
    ArrowLeft: () => {
      change(value - keyStep);
    },
    ArrowRight: () => {
      change(value + keyStep);
    },
    End: () => {
      change(maxValue);
    },
    Enter: () => {
      change(defaultValue);
    },
    Home: () => {
      change(minValue);
    },
  };

  return (
    <div
      role="separator"
      aria-label={label}
      aria-orientation="vertical"
      aria-valuenow={value}
      aria-valuemin={minValue}
      aria-valuemax={maxValue}
      tabIndex={0}
      data-dragging={isDragging || undefined}
      onPointerDown={(event) => {
        event.currentTarget.setPointerCapture(event.pointerId);
        start.current = { value, x: event.clientX };
        setIsDragging(true);
      }}
      onPointerMove={(event) => {
        if (isDragging) {
          change(start.current.value + (event.clientX - start.current.x) * direction);
        }
      }}
      onPointerUp={() => {
        setIsDragging(false);
      }}
      onPointerCancel={() => {
        setIsDragging(false);
      }}
      onDoubleClick={() => {
        change(defaultValue);
      }}
      onKeyDown={(event) => {
        const action = keyActions[event.key];
        if (action) {
          event.preventDefault();
          action();
        }
      }}
      className={cn(
        "group absolute inset-y-0 z-10 flex w-1.5 cursor-col-resize touch-none outline-none",
        edge === "right" ? "right-0 justify-end" : "left-0 justify-start",
      )}
    >
      <span className="w-0.5 bg-transparent transition-colors group-hover:bg-accent group-focus-visible:bg-focus group-data-dragging:bg-accent" />
    </div>
  );
};
