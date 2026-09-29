import { fromDate, getLocalTimeZone } from "@internationalized/date";
import { DateField, DateInput, DateSegment, FieldError, Label } from "react-aria-components";

import { cn, fieldStyles } from "./styles";

type DateTimeFieldProps = {
  label: string;
  value: Date;
  onChange: (value: Date) => void;
  // 指定するとフィールドを不正な状態として表示する
  errorMessage?: string;
};

// 端末のタイムゾーンで日付と時刻（分まで）を入力する
export const DateTimeField = ({ label, value, onChange, errorMessage }: DateTimeFieldProps) => (
  <DateField
    value={fromDate(value, getLocalTimeZone())}
    onChange={(next) => {
      if (next !== null) {
        onChange(next.toDate());
      }
    }}
    granularity="minute"
    hideTimeZone
    isInvalid={errorMessage !== undefined}
    className={fieldStyles.root}
  >
    <Label className={fieldStyles.label}>{label}</Label>
    <DateInput
      className={cn(
        fieldStyles.input,
        "flex items-center data-focus-within:border-accent data-focus-within:ring-3 data-focus-within:ring-accent-soft",
      )}
    >
      {(segment) => (
        <DateSegment
          segment={segment}
          className="rounded-sm px-px tabular-nums outline-none data-focused:bg-accent data-focused:text-accent-fg data-placeholder:text-subtle data-[type=literal]:px-0"
        />
      )}
    </DateInput>
    <FieldError className={fieldStyles.error}>{errorMessage}</FieldError>
  </DateField>
);
