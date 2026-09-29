import type { ReactNode } from "react";

import { Header, MenuSection as AriaMenuSection } from "react-aria-components";

type MenuSectionProps = {
  title: string;
  children: ReactNode;
};

export const MenuSection = ({ title, children }: MenuSectionProps) => (
  <AriaMenuSection className="flex flex-col">
    <Header className="px-2.5 pt-1.5 pb-1 text-[11.5px] font-semibold text-subtle">{title}</Header>
    {children}
  </AriaMenuSection>
);
