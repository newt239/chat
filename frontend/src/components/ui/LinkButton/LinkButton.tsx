import { createLink } from "@tanstack/react-router";
import { Link as AriaLink } from "react-aria-components";

import { buttonClassName } from "#/components/ui/Button/buttonClassName";

import type { ButtonSize, ButtonVariant } from "#/components/ui/Button/buttonClassName";

import type { LinkProps } from "react-aria-components";

type LinkButtonProps = LinkProps & {
  variant?: ButtonVariant;
  size?: ButtonSize;
};

// 見た目がボタンのリンク。遷移先は Link と同じく to / params で指定する
export const LinkButton = createLink(
  ({ variant = "primary", size = "md", className, ...props }: LinkButtonProps) => (
    <AriaLink {...props} className={buttonClassName(variant, size, className)} />
  ),
);
