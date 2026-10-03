import { FieldError, Group, Input, Label, Text, TextField } from "react-aria-components";

import { cn, fieldStyles } from "#/components/ui/styles/styles";

type ChannelNameFieldProps = {
  label: string;
  // 入力欄の前に固定で表示する文字列（"#" や "#dev/"）
  prefix: string;
  value: string;
  onChange: (value: string) => void;
  description: string;
  errorMessage: string | null;
  placeholder: string;
};

// 英字は小文字にそろえて受け取る
export const ChannelNameField = ({
  label,
  prefix,
  value,
  onChange,
  description,
  errorMessage,
  placeholder,
}: ChannelNameFieldProps) => (
  <TextField
    value={value}
    onChange={(next) => {
      onChange(next.toLowerCase());
    }}
    isInvalid={errorMessage !== null}
    autoComplete="off"
    spellCheck="false"
    className={fieldStyles.root}
  >
    <Label className={fieldStyles.label}>{label}</Label>
    <Group
      className={cn(
        fieldStyles.input,
        "flex items-center px-0 data-focus-within:border-accent data-focus-within:ring-3 data-focus-within:ring-accent-soft data-invalid:border-danger",
      )}
    >
      <span aria-hidden className="shrink-0 pl-2.5 font-mono text-subtle">
        {prefix}
      </span>
      <Input
        placeholder={placeholder}
        className="h-full min-w-0 flex-1 border-0 bg-transparent pr-2.5 pl-0.5 font-sans text-body-sm text-text outline-none placeholder:text-subtle"
      />
    </Group>
    {errorMessage === null ? (
      <Text slot="description" className={fieldStyles.description}>
        {description}
      </Text>
    ) : (
      <FieldError className={fieldStyles.error}>{errorMessage}</FieldError>
    )}
  </TextField>
);
