import { IconCheck, IconChevronDown } from "@tabler/icons-react";
import {
  Select as AriaSelect,
  Button,
  Label,
  ListBox,
  ListBoxItem,
  Popover,
  SelectValue,
  Text,
} from "react-aria-components";

import { findOption } from "#/components/ui/option/option";
import { cn, fieldStyles, focusRing, overlayStyles } from "#/components/ui/styles/styles";

import type { Option } from "#/components/ui/option/option";

type SelectProps<T extends string> = {
  label: string;
  options: readonly Option<T>[];
  value: T;
  onChange: (value: T) => void;
  description?: string;
  // 表の中に置くため、見出しを出さない小さな形にする
  isCompact?: boolean;
  isDisabled?: boolean;
  className?: string;
};

export const Select = <T extends string>({
  label,
  options,
  value,
  onChange,
  description,
  isCompact,
  isDisabled,
  className,
}: SelectProps<T>) => (
  <AriaSelect
    aria-label={isCompact ? label : undefined}
    value={value}
    onChange={(key) => {
      const option = findOption(options, key);
      if (option) {
        onChange(option.value);
      }
    }}
    isDisabled={isDisabled}
    className={cn(fieldStyles.root, className)}
  >
    {!isCompact && <Label className={fieldStyles.label}>{label}</Label>}
    <Button
      className={cn(
        focusRing,
        "flex cursor-pointer items-center text-left",
        isCompact
          ? "h-7 min-w-24 gap-1.5 rounded-md border border-border-strong bg-surface px-2 text-label font-normal text-text data-disabled:cursor-default data-disabled:bg-sunken data-disabled:text-subtle"
          : cn(fieldStyles.input, "gap-2 data-pressed:border-accent"),
      )}
    >
      <SelectValue className="flex-1 truncate" />
      <IconChevronDown
        aria-hidden
        className={cn("shrink-0 text-muted", isCompact ? "size-3.5" : "size-4")}
      />
    </Button>
    {description && (
      <Text slot="description" className={fieldStyles.description}>
        {description}
      </Text>
    )}
    <Popover offset={4} className={cn(overlayStyles.popover, "min-w-(--trigger-width) p-1")}>
      <ListBox items={options} className="max-h-72 overflow-y-auto outline-none">
        {(option) => (
          <ListBoxItem
            id={option.value}
            textValue={option.label}
            className={overlayStyles.listItem}
          >
            {({ isSelected }) => (
              <>
                <span className="flex-1 truncate">{option.label}</span>
                {isSelected && <IconCheck aria-hidden />}
              </>
            )}
          </ListBoxItem>
        )}
      </ListBox>
    </Popover>
  </AriaSelect>
);
