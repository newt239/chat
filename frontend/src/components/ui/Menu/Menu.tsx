import { useState } from "react";
import type { ReactElement, ReactNode } from "react";

import { Menu as AriaMenu, Heading, MenuTrigger, Popover } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { DialogFrame } from "#/components/ui/Dialog/DialogFrame";
import { cn, overlayStyles } from "#/components/ui/styles/styles";
import { useIsMobile } from "#/hooks/useMediaQuery";

import type { PopoverProps } from "react-aria-components";

type MenuProps = {
  // React Aria の Button（IconButton など）。メニューの名前はトリガーのラベルになる
  trigger: ReactElement;
  // MenuItem / MenuSeparator / MenuSection
  children: ReactNode;
  placement?: PopoverProps["placement"];
  onOpenChange?: (isOpen: boolean) => void;
};

// モバイルでは親指の届く下からシートで出す
export const Menu = ({ trigger, children, placement = "bottom end", onOpenChange }: MenuProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [isOpen, setIsOpen] = useState(false);
  const handleOpenChange = (next: boolean) => {
    setIsOpen(next);
    onOpenChange?.(next);
  };

  return (
    <MenuTrigger isOpen={isOpen} onOpenChange={handleOpenChange}>
      {trigger}
      {isMobile ? (
        <DialogFrame
          isOpen={isOpen}
          onOpenChange={handleOpenChange}
          layout="bottom"
          role="dialog"
          className="bg-raised pt-2"
        >
          <Heading slot="title" className="sr-only">
            {t("ui.menu.title")}
          </Heading>
          <AriaMenu className={cn(overlayStyles.menu, "min-w-0 px-2")}>{children}</AriaMenu>
        </DialogFrame>
      ) : (
        <Popover
          offset={4}
          placement={placement}
          className={cn(overlayStyles.popover, "overflow-y-auto")}
        >
          <AriaMenu className={overlayStyles.menu}>{children}</AriaMenu>
        </Popover>
      )}
    </MenuTrigger>
  );
};
