import { TabList as AriaTabList } from "react-aria-components";

import { withBaseClassName } from "#/components/ui/styles/styles";

import type { TabListProps } from "react-aria-components";

export const TabList = <T extends object>({ className, ...props }: TabListProps<T>) => (
  <AriaTabList
    {...props}
    className={withBaseClassName(
      className,
      "flex shrink-0 gap-4.5 overflow-x-auto border-b border-border px-4.5 [scrollbar-width:none]",
    )}
  />
);
