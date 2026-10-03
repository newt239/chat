import type { ReactElement, ReactNode } from "react";

import { Tooltip as AriaTooltip, TooltipTrigger } from "react-aria-components";

type TooltipProps = {
  content: ReactNode;
  // React Aria の Button などフォーカスできる要素
  children: ReactElement;
};

// ホバーは補助。ツールチップだけに情報を置かず、タップでも届く場所に同じ情報を置く
export const Tooltip = ({ content, children }: TooltipProps) => (
  <TooltipTrigger delay={300} closeDelay={0}>
    {children}
    <AriaTooltip
      offset={6}
      className="max-w-60 rounded-md bg-text px-2 py-1 font-sans text-caption text-surface tabular-nums shadow-md data-entering:animate-fade-in data-exiting:animate-fade-out motion-reduce:animate-none"
    >
      {content}
    </AriaTooltip>
  </TooltipTrigger>
);
