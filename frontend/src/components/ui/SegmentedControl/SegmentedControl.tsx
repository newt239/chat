import { ToggleButton, ToggleButtonGroup } from "react-aria-components";

import { findOption } from "#/components/ui/option/option";
import { focusRing } from "#/components/ui/styles/styles";

import type { Option } from "#/components/ui/option/option";

type SegmentedControlProps<T extends string> = {
  label: string;
  options: readonly Option<T>[];
  value: T;
  onChange: (value: T) => void;
};

// 少数の選択肢から 1 つを選ぶ。モードや言語の切り替えに使う
export const SegmentedControl = <T extends string>({
  label,
  options,
  value,
  onChange,
}: SegmentedControlProps<T>) => (
  <ToggleButtonGroup
    aria-label={label}
    selectionMode="single"
    disallowEmptySelection
    selectedKeys={[value]}
    onSelectionChange={(keys) => {
      const option = findOption(options, [...keys][0] ?? null);
      if (option) {
        onChange(option.value);
      }
    }}
    className="inline-flex flex-wrap gap-0.5 rounded-md border border-border bg-sunken p-0.5 font-sans"
  >
    {options.map((option) => (
      <ToggleButton
        key={option.value}
        id={option.value}
        className={`min-h-[26px] cursor-pointer rounded-[6px] px-2.5 text-[12.5px] text-muted data-hovered:text-text data-selected:bg-surface data-selected:font-semibold data-selected:text-text data-selected:shadow-sm ${focusRing}`}
      >
        {option.label}
      </ToggleButton>
    ))}
  </ToggleButtonGroup>
);
