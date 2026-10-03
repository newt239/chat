import { useRef, useState } from "react";
import type { MouseEvent, ReactNode } from "react";

import { Menu as AriaMenu, Popover } from "react-aria-components";

import { cn, overlayStyles } from "#/components/ui/styles/styles";

type Point = { x: number; y: number };

type ContextMenuProps = {
  // 右クリック（キーボードのコンテキストメニューキーを含む）の対象
  children: ReactNode;
  // MenuItem / MenuSeparator / MenuSection。Menu と同じ定義を渡せる
  menu: ReactNode;
  "aria-label": string;
  className?: string;
};

const pointOf = (event: MouseEvent<HTMLDivElement>) => {
  // キーボードから開いた場合は座標が 0 になるため、対象の左上に出す
  if (event.clientX === 0 && event.clientY === 0) {
    const rect = event.currentTarget.getBoundingClientRect();
    return { x: rect.left, y: rect.top };
  }
  return { x: event.clientX, y: event.clientY };
};

export const ContextMenu = ({ children, menu, className, ...props }: ContextMenuProps) => {
  const anchorRef = useRef<HTMLSpanElement>(null);
  const [point, setPoint] = useState<Point | null>(null);
  const close = () => {
    setPoint(null);
  };

  return (
    <div
      className={className}
      onContextMenu={(event) => {
        event.preventDefault();
        setPoint(pointOf(event));
      }}
    >
      {children}
      <span
        ref={anchorRef}
        aria-hidden
        className="pointer-events-none fixed size-0"
        style={{ left: point?.x, top: point?.y }}
      />
      <Popover
        triggerRef={anchorRef}
        isOpen={point !== null}
        onOpenChange={close}
        placement="bottom start"
        offset={0}
        className={cn(overlayStyles.popover, "overflow-y-auto")}
      >
        {/* 独自のトリガーから開くため、キーボード操作できるよう先頭の項目にフォーカスする */}
        {/* oxlint-disable-next-line jsx-a11y/no-autofocus */}
        <AriaMenu {...props} autoFocus="first" onClose={close} className={overlayStyles.menu}>
          {menu}
        </AriaMenu>
      </Popover>
    </div>
  );
};
