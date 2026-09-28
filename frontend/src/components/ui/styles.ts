import { colorTokenNames, radius, shadow, typography } from "@chat/design-tokens";
import { composeRenderProps } from "react-aria-components";
import { extendTailwindMerge } from "tailwind-merge";

// 独自トークン（text-caption と text-muted など）を別グループとして扱わせる
export const cn = extendTailwindMerge({
  extend: {
    theme: {
      color: [...colorTokenNames],
      radius: Object.keys(radius),
      shadow: Object.keys(shadow),
      text: Object.keys(typography),
    },
  },
});

// React Aria の className（文字列 or 状態を受け取る関数）に既定のクラスを合成する
export const withBaseClassName = <T>(
  className: string | ((values: T) => string) | undefined,
  base: string,
) => composeRenderProps(className, (value) => cn(base, value));

export const focusRing =
  "outline-none data-focus-visible:outline-2 data-focus-visible:outline-offset-1 data-focus-visible:outline-focus data-focus-visible:outline-solid";

export const fieldStyles = {
  description: "text-xs text-muted",
  error: "text-xs text-danger",
  input:
    "h-[34px] w-full min-w-0 rounded-md border border-border-strong bg-surface px-2.5 font-sans text-[13.5px] text-text outline-none placeholder:text-subtle data-disabled:bg-sunken data-disabled:text-subtle data-focused:border-accent data-focused:ring-3 data-focused:ring-accent-soft data-invalid:border-danger",
  label: "text-label text-text",
  root: "flex flex-col gap-[5px] font-sans",
};

export const overlayStyles = {
  listItem:
    "flex min-h-[30px] cursor-default items-center gap-2.5 rounded-[6px] px-2.5 text-[13.5px] text-text no-underline outline-none data-disabled:text-subtle data-focused:bg-accent data-focused:text-accent-fg [&_svg]:size-4 [&_svg]:shrink-0",
  menu: "flex min-w-[220px] flex-col p-1 outline-none",
  popover:
    "rounded-lg border border-border bg-raised font-sans text-text shadow-lg outline-none data-entering:animate-pop-in data-exiting:animate-pop-out motion-reduce:animate-none",
};
