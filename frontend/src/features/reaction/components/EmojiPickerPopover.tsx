import { useState } from "react";
import type { ReactElement } from "react";

import { DialogTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Popover } from "#/components/ui/Popover";

import { EmojiPicker } from "./EmojiPicker";

type EmojiPickerPopoverProps = {
  // React Aria の Button（IconButton など）
  trigger: ReactElement;
  onSelect: (emoji: string) => void;
  onOpenChange?: (isOpen: boolean) => void;
};

export const EmojiPickerPopover = ({
  trigger,
  onSelect,
  onOpenChange,
}: EmojiPickerPopoverProps) => {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const changeOpen = (next: boolean) => {
    setIsOpen(next);
    onOpenChange?.(next);
  };

  return (
    <DialogTrigger isOpen={isOpen} onOpenChange={changeOpen}>
      {trigger}
      <Popover aria-label={t("reaction.add")} placement="bottom end" className="overflow-hidden">
        <EmojiPicker
          onEmojiSelect={(emoji) => {
            onSelect(emoji);
            changeOpen(false);
          }}
        />
      </Popover>
    </DialogTrigger>
  );
};
