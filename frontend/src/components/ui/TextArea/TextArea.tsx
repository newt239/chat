import { TextArea as AriaTextArea, TextField as AriaTextField, Label } from "react-aria-components";

import { cn, fieldStyles, withBaseClassName } from "#/components/ui/styles/styles";

import type { TextFieldProps as AriaTextFieldProps } from "react-aria-components";

type TextAreaProps = Omit<AriaTextFieldProps, "children"> & {
  label: string;
  placeholder?: string;
  rows?: number;
};

export const TextArea = ({ label, placeholder, rows = 3, className, ...props }: TextAreaProps) => (
  <AriaTextField {...props} className={withBaseClassName(className, fieldStyles.root)}>
    <Label className={fieldStyles.label}>{label}</Label>
    <AriaTextArea
      rows={rows}
      placeholder={placeholder}
      className={cn(fieldStyles.input, "h-auto resize-y px-2.5 py-1.5 leading-normal")}
    />
  </AriaTextField>
);
