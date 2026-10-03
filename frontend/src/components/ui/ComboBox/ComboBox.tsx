import { IconSelector } from "@tabler/icons-react";
import {
  ComboBox as AriaComboBox,
  Button,
  Input,
  Label,
  ListBox,
  ListBoxItem,
  Popover,
  Text,
} from "react-aria-components";
import { useTranslation } from "react-i18next";

import { findOption } from "#/components/ui/option/option";
import { cn, fieldStyles, overlayStyles } from "#/components/ui/styles/styles";

import type { Option } from "#/components/ui/option/option";

type ComboBoxProps<T extends string> = {
  label: string;
  options: readonly Option<T>[];
  value: T | null;
  onChange: (value: T | null) => void;
  description?: string;
  placeholder?: string;
  isDisabled?: boolean;
  className?: string;
};

// 入力で絞り込める Select。候補の絞り込みは React Aria に任せる
export const ComboBox = <T extends string>({
  label,
  options,
  value,
  onChange,
  description,
  placeholder,
  isDisabled,
  className,
}: ComboBoxProps<T>) => {
  const { t } = useTranslation();
  return (
    <AriaComboBox
      defaultItems={options}
      value={value}
      onChange={(key) => {
        onChange(findOption(options, key)?.value ?? null);
      }}
      isDisabled={isDisabled}
      menuTrigger="focus"
      className={cn(fieldStyles.root, className)}
    >
      <Label className={fieldStyles.label}>{label}</Label>
      <div className="relative">
        <Input placeholder={placeholder} className={cn(fieldStyles.input, "pr-8 max-md:pr-10")} />
        <Button
          aria-label={t("ui.comboBox.showSuggestions")}
          className="absolute inset-y-0 right-1 my-auto grid size-7 cursor-pointer max-md:size-9 place-items-center rounded-sm text-muted outline-none data-hovered:text-text"
        >
          <IconSelector aria-hidden className="size-4" />
        </Button>
      </div>
      {description && (
        <Text slot="description" className={fieldStyles.description}>
          {description}
        </Text>
      )}
      <Popover offset={4} className={cn(overlayStyles.popover, "w-(--trigger-width) p-1")}>
        <ListBox
          className="max-h-72 overflow-y-auto outline-none"
          renderEmptyState={() => (
            <p className="m-0 px-2.5 py-1.5 text-body-sm text-muted">{t("ui.comboBox.empty")}</p>
          )}
        >
          {(option: Option<T>) => (
            <ListBoxItem
              id={option.value}
              textValue={option.label}
              className={overlayStyles.listItem}
            >
              {option.label}
            </ListBoxItem>
          )}
        </ListBox>
      </Popover>
    </AriaComboBox>
  );
};
