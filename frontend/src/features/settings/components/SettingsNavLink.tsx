import { createLink } from "@tanstack/react-router";
import { Link as AriaLink } from "react-aria-components";

import { focusRing, withBaseClassName } from "#/components/ui/styles/styles";

import type { LinkProps } from "react-aria-components";

// 設定ページの左の項目。今の項目には TanStack Router が aria-current="page" を付ける
export const SettingsNavLink = createLink(({ className, ...props }: LinkProps) => (
  <AriaLink
    {...props}
    className={withBaseClassName(
      className,
      `flex h-8 items-center gap-2 rounded-md px-2.5 text-[13.5px] text-muted no-underline data-hovered:bg-hover data-hovered:text-text aria-[current=page]:bg-accent-soft aria-[current=page]:font-semibold aria-[current=page]:text-accent-text [&_svg]:size-4 ${focusRing}`,
    )}
  />
));
