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
  isDisabled?: boolean;
  className?: string;
};

export const Select = <T extends string>({
  label,
  options,
  value,
  onChange,
  description,
  isDisabled,
  className,
}: SelectProps<T>) => (
  <AriaSelect
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
    <Label className={fieldStyles.label}>{label}</Label>
    <Button
      className={cn(
        fieldStyles.input,
        focusRing,
        "flex cursor-pointer items-center gap-2 text-left data-pressed:border-accent",
      )}
    >
      <SelectValue className="flex-1 truncate" />
      <IconChevronDown aria-hidden className="size-4 shrink-0 text-muted" />
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
