import { createLink } from "@tanstack/react-router";
import { Link as AriaLink } from "react-aria-components";

import { focusRing, withBaseClassName } from "#/components/ui/styles/styles";

import type { LinkProps } from "react-aria-components";

// TanStack Router の to / params で型検査される React Aria のリンク
export const Link = createLink(({ className, ...props }: LinkProps) => (
  <AriaLink
    {...props}
    className={withBaseClassName(
      className,
      `cursor-pointer rounded-sm text-accent-text underline decoration-1 underline-offset-2 ${focusRing}`,
    )}
  />
));
