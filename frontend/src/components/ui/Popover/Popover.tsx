import type { ReactNode } from "react";

import { Popover as AriaPopover, Dialog } from "react-aria-components";

import { cn, overlayStyles } from "#/components/ui/styles/styles";

import type { PopoverProps as AriaPopoverProps } from "react-aria-components";

type PopoverProps = Omit<AriaPopoverProps, "children" | "className"> & {
  children: ReactNode;
  "aria-label": string;
  className?: string;
};

// DialogTrigger の中にトリガーと並べて置く。メニューや選択肢には Menu / Select を使う
export const Popover = ({
  children,
  className,
  "aria-label": ariaLabel,
  ...props
}: PopoverProps) => (
  <AriaPopover offset={6} {...props} className={cn(overlayStyles.popover, className)}>
    <Dialog aria-label={ariaLabel} className="outline-none">
      {children}
    </Dialog>
  </AriaPopover>
);
