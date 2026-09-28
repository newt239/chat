import { createLink } from "@tanstack/react-router";
import { Link as AriaLink } from "react-aria-components";

import { focusRing, withBaseClassName } from "#/components/ui/styles";

import { navItemClassName } from "../utils/navTone";

import type { LinkProps } from "react-aria-components";

// サイドバーとモバイルの一覧の行。配色は親が --nav-* で決める（サイドバーは side 系、モバイルは surface 系）
export const NavLink = createLink(({ className, ...props }: LinkProps) => (
  <AriaLink
    {...props}
    className={withBaseClassName(className, `${navItemClassName} ${focusRing}`)}
  />
));
