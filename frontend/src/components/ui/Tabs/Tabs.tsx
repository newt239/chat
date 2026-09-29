import { useId } from "react";

import { LayoutGroup } from "motion/react";
import { Tabs as AriaTabs } from "react-aria-components";

import { withBaseClassName } from "#/components/ui/styles/styles";

import type { TabsProps } from "react-aria-components";

// LayoutGroup で下線の layoutId をタブ群ごとに分け、複数のタブが同時にあっても干渉させない
export const Tabs = ({ className, ...props }: TabsProps) => (
  <LayoutGroup id={useId()}>
    <AriaTabs
      {...props}
      className={withBaseClassName(className, "flex min-h-0 flex-col font-sans")}
    />
  </LayoutGroup>
);
