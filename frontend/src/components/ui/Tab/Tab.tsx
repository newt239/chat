import { motion } from "motion/react";
import { Tab as AriaTab, composeRenderProps } from "react-aria-components";

import { focusRing, withBaseClassName } from "#/components/ui/styles/styles";
import { transitions } from "#/lib/motion";

import type { TabProps } from "react-aria-components";

export const Tab = ({ className, children, ...props }: TabProps) => (
  <AriaTab
    {...props}
    className={withBaseClassName(
      className,
      `relative flex h-9 shrink-0 cursor-pointer items-center gap-1.5 whitespace-nowrap text-body-sm text-muted data-disabled:cursor-default data-disabled:text-subtle data-hovered:text-text data-selected:font-semibold data-selected:text-text ${focusRing}`,
    )}
  >
    {composeRenderProps(children, (content, { isSelected }) => (
      <>
        {content}
        {isSelected && (
          <motion.span
            layoutId="tab-underline"
            transition={transitions.spring}
            className="absolute inset-x-0 -bottom-px h-0.5 rounded-xs bg-accent"
          />
        )}
      </>
    ))}
  </AriaTab>
);
