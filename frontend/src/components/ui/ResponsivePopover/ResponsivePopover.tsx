import type { ComponentProps } from "react";

import { Heading } from "react-aria-components";

import { DialogFrame } from "#/components/ui/Dialog/DialogFrame";
import { Popover } from "#/components/ui/Popover/Popover";
import { cn } from "#/components/ui/styles/styles";
import { useIsMobile } from "#/hooks/useMediaQuery";

type ResponsivePopoverProps = ComponentProps<typeof Popover> & {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

// デスクトップではトリガーの横に、モバイルでは親指の届く下からシートで出す
export const ResponsivePopover = ({
  isOpen,
  onOpenChange,
  className,
  children,
  "aria-label": ariaLabel,
  ...props
}: ResponsivePopoverProps) => {
  const isMobile = useIsMobile();
  return isMobile ? (
    <DialogFrame
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      layout="bottom"
      role="dialog"
      className={cn("bg-raised pt-2", className)}
    >
      <Heading slot="title" className="sr-only">
        {ariaLabel}
      </Heading>
      {children}
    </DialogFrame>
  ) : (
    <Popover
      {...props}
      aria-label={ariaLabel}
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      className={className}
    >
      {children}
    </Popover>
  );
};
