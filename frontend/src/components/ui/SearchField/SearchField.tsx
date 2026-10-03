import type { Ref } from "react";

import { IconSearch } from "@tabler/icons-react";
import { SearchField as AriaSearchField, Input } from "react-aria-components";

import { withBaseClassName } from "#/components/ui/styles/styles";

import type { SearchFieldProps as AriaSearchFieldProps } from "react-aria-components";

type SearchFieldProps = Omit<AriaSearchFieldProps, "children" | "aria-label"> & {
  // aria-label に使う。placeholder を省くとプレースホルダーにも使う
  label: string;
  placeholder?: string;
  inputRef?: Ref<HTMLInputElement>;
};

// 虫眼鏡のアイコン付きの検索欄。文字の大きさや枠は className で上書きする
export const SearchField = ({
  label,
  placeholder,
  inputRef,
  className,
  ...props
}: SearchFieldProps) => (
  <AriaSearchField
    {...props}
    aria-label={label}
    className={withBaseClassName(
      className,
      "flex h-8.5 min-w-0 items-center gap-2 rounded-md border border-border-strong bg-surface px-2.5 font-sans text-body-sm text-muted data-focus-within:border-accent data-focus-within:ring-3 data-focus-within:ring-accent-soft max-md:h-11",
    )}
  >
    <IconSearch aria-hidden className="size-4 shrink-0" />
    <Input
      ref={inputRef}
      placeholder={placeholder ?? label}
      className="h-full min-w-0 flex-1 border-0 bg-transparent text-text outline-none [font-size:inherit] placeholder:text-subtle [&::-webkit-search-cancel-button]:hidden"
    />
  </AriaSearchField>
);
