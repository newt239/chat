import { IconLoader2 } from "@tabler/icons-react";
import { Button as AriaButton, composeRenderProps } from "react-aria-components";

import { buttonClassName } from "./buttonClassName";

import type { ButtonSize, ButtonVariant } from "./buttonClassName";

import type { ButtonProps as AriaButtonProps } from "react-aria-components";

type ButtonProps = AriaButtonProps & {
  variant?: ButtonVariant;
  size?: ButtonSize;
};

export const Button = ({
  variant = "primary",
  size = "md",
  className,
  children,
  ...props
}: ButtonProps) => (
  <AriaButton {...props} className={buttonClassName(variant, size, className)}>
    {composeRenderProps(children, (content, { isPending }) => (
      <>
        {isPending && (
          <IconLoader2 aria-hidden className="animate-spin motion-reduce:animate-none" />
        )}
        {content}
      </>
    ))}
  </AriaButton>
);
