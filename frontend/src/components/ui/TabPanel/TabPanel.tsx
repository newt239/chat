import { TabPanel as AriaTabPanel } from "react-aria-components";

import { focusRing, withBaseClassName } from "#/components/ui/styles/styles";

import type { TabPanelProps } from "react-aria-components";

export const TabPanel = ({ className, ...props }: TabPanelProps) => (
  <AriaTabPanel
    {...props}
    className={withBaseClassName(className, `min-h-0 flex-1 ${focusRing}`)}
  />
);
