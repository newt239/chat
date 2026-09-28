import type { ReactNode } from "react";

import { IconCheck, IconMinus } from "@tabler/icons-react";
import { CheckboxButton, CheckboxField } from "react-aria-components";

import { withBaseClassName } from "./styles";

import type { CheckboxFieldProps } from "react-aria-components";

type CheckboxProps = Omit<CheckboxFieldProps, "children"> & {
  children: ReactNode;
};

export const Checkbox = ({ className, children, ...props }: CheckboxProps) => (
  <CheckboxField {...props} className={withBaseClassName(className, "flex font-sans")}>
    <CheckboxButton className="group flex cursor-pointer items-center gap-2 py-0.5 text-[13px] text-text data-disabled:cursor-default data-disabled:text-subtle">
      {({ isSelected, isIndeterminate }) => (
        <>
          <span className="grid size-[15px] shrink-0 place-items-center rounded-sm border border-border-strong bg-surface text-accent-fg transition-colors group-data-focus-visible:outline-2 group-data-focus-visible:outline-offset-1 group-data-focus-visible:outline-focus group-data-focus-visible:outline-solid group-data-indeterminate:border-accent group-data-indeterminate:bg-accent group-data-invalid:border-danger group-data-selected:border-accent group-data-selected:bg-accent [&_svg]:size-3">
            {isIndeterminate ? (
              <IconMinus aria-hidden stroke={3} />
            ) : (
              isSelected && <IconCheck aria-hidden stroke={3} />
            )}
          </span>
          {children}
        </>
      )}
    </CheckboxButton>
  </CheckboxField>
);
