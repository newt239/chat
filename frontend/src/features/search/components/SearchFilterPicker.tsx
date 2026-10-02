import { IconCheck, IconChevronDown } from "@tabler/icons-react";
import {
  Autocomplete,
  Button,
  Menu,
  MenuItem,
  MenuTrigger,
  Popover,
  useFilter,
} from "react-aria-components";
import { useTranslation } from "react-i18next";

import { SearchField } from "#/components/ui/SearchField/SearchField";
import { cn, overlayStyles } from "#/components/ui/styles/styles";
import { chipClassName } from "#/features/search/utils/chipClassName";

import type { Option } from "#/components/ui/option/option";

type SearchFilterPickerProps<T extends string> = {
  label: string;
  // 選択中の値の表示。未選択なら null
  summary: string | null;
  options: readonly Option<T>[];
  selected: readonly T[];
  onToggle: (value: T) => void;
  isSearchable: boolean;
  isInvalid: boolean;
};

// チップを押すと選択肢を開く。選ぶたびに onToggle を呼ぶ
export const SearchFilterPicker = <T extends string>({
  label,
  summary,
  options,
  selected,
  onToggle,
  isSearchable,
  isInvalid,
}: SearchFilterPickerProps<T>) => {
  const { t } = useTranslation();
  const { contains } = useFilter({ sensitivity: "base" });

  const menu = (
    <Menu
      aria-label={label}
      items={options}
      selectionMode="multiple"
      selectedKeys={selected}
      onSelectionChange={(keys) => {
        // 選択の差分から押された選択肢を探す
        const option =
          keys === "all"
            ? undefined
            : options.find(({ value }) => keys.has(value) !== selected.includes(value));
        if (option) {
          onToggle(option.value);
        }
      }}
      renderEmptyState={() => (
        <p className="m-0 px-2.5 py-1.5 text-[13px] text-muted">{t("ui.comboBox.empty")}</p>
      )}
      className="max-h-72 overflow-y-auto p-1 outline-none"
    >
      {(option) => (
        <MenuItem id={option.value} textValue={option.label} className={overlayStyles.listItem}>
          {({ isSelected }) => (
            <>
              <span className="flex-1 truncate">{option.label}</span>
              {isSelected && <IconCheck aria-hidden />}
            </>
          )}
        </MenuItem>
      )}
    </Menu>
  );

  return (
    <MenuTrigger>
      <Button data-active={summary !== null} data-invalid={isInvalid} className={chipClassName}>
        {summary === null ? label : `${label}: ${summary}`}
        <IconChevronDown aria-hidden />
      </Button>
      <Popover
        offset={4}
        placement="bottom start"
        className={cn(overlayStyles.popover, "flex w-64 flex-col")}
      >
        {isSearchable ? (
          <Autocomplete filter={contains}>
            <SearchField
              label={t("search.filters.find")}
              autoFocus
              className="rounded-none border-0 border-b border-border data-focus-within:ring-0"
            />
            {menu}
          </Autocomplete>
        ) : (
          menu
        )}
      </Popover>
    </MenuTrigger>
  );
};
