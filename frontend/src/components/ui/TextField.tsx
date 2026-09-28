import { TextField as AriaTextField, FieldError, Input, Label, Text } from "react-aria-components";

import { fieldStyles, withBaseClassName } from "./styles";

import type { TextFieldProps as AriaTextFieldProps } from "react-aria-components";

type TextFieldProps = Omit<AriaTextFieldProps, "children"> & {
  label: string;
  description?: string;
  // 指定するとフィールドを不正な状態として表示する
  errorMessage?: string;
  placeholder?: string;
};

export const TextField = ({
  label,
  description,
  errorMessage,
  placeholder,
  className,
  isInvalid,
  ...props
}: TextFieldProps) => (
  <AriaTextField
    {...props}
    isInvalid={isInvalid || Boolean(errorMessage)}
    className={withBaseClassName(className, fieldStyles.root)}
  >
    <Label className={fieldStyles.label}>{label}</Label>
    <Input placeholder={placeholder} className={fieldStyles.input} />
    {description && (
      <Text slot="description" className={fieldStyles.description}>
        {description}
      </Text>
    )}
    <FieldError className={fieldStyles.error}>{errorMessage}</FieldError>
  </AriaTextField>
);
