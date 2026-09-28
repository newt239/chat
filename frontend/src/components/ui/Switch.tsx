import type { ReactNode } from "react";

import { SwitchButton, SwitchField } from "react-aria-components";

import { withBaseClassName } from "./styles";

import type { SwitchFieldProps } from "react-aria-components";

type SwitchProps = Omit<SwitchFieldProps, "children"> & {
  children: ReactNode;
};

export const Switch = ({ className, children, ...props }: SwitchProps) => (
  <SwitchField {...props} className={withBaseClassName(className, "flex font-sans")}>
    <SwitchButton className="group flex cursor-pointer items-center gap-2 text-[13px] text-text data-disabled:cursor-default data-disabled:text-subtle">
      <span className="relative h-[22px] w-9 shrink-0 rounded-full bg-border-strong transition-colors group-data-disabled:opacity-50 group-data-focus-visible:outline-2 group-data-focus-visible:outline-offset-1 group-data-focus-visible:outline-focus group-data-focus-visible:outline-solid group-data-selected:bg-accent">
        <span className="absolute top-[3px] left-[3px] size-4 rounded-full bg-white shadow-sm transition-transform group-data-selected:translate-x-3.5 motion-reduce:transition-none" />
      </span>
      {children}
    </SwitchButton>
  </SwitchField>
);
