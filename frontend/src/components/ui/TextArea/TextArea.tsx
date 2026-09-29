import {
  TextArea as AriaTextArea,
  TextField as AriaTextField,
  FieldError,
  Label,
  Text,
} from "react-aria-components";

import { cn, fieldStyles, withBaseClassName } from "#/components/ui/styles/styles";

import type { TextFieldProps as AriaTextFieldProps } from "react-aria-components";

type TextAreaProps = Omit<AriaTextFieldProps, "children"> & {
  label: string;
  description?: string;
  errorMessage?: string;
  placeholder?: string;
  rows?: number;
};

export const TextArea = ({
  label,
  description,
  errorMessage,
  placeholder,
  rows = 3,
  className,
  isInvalid,
  ...props
}: TextAreaProps) => (
  <AriaTextField
    {...props}
    isInvalid={isInvalid || Boolean(errorMessage)}
    className={withBaseClassName(className, fieldStyles.root)}
  >
    <Label className={fieldStyles.label}>{label}</Label>
    <AriaTextArea
      rows={rows}
      placeholder={placeholder}
      className={cn(fieldStyles.input, "h-auto resize-y px-2.5 py-1.5 leading-normal")}
    />
    {description && (
      <Text slot="description" className={fieldStyles.description}>
        {description}
      </Text>
    )}
    <FieldError className={fieldStyles.error}>{errorMessage}</FieldError>
  </AriaTextField>
);
