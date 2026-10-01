import { useState } from "react";

import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { IconPlus, IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Checkbox } from "#/components/ui/Checkbox/Checkbox";
import { DateTimeField } from "#/components/ui/DateTimeField/DateTimeField";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { SegmentedControl } from "#/components/ui/SegmentedControl/SegmentedControl";
import { Switch } from "#/components/ui/Switch/Switch";
import { TextField } from "#/components/ui/TextField/TextField";
import { PollInputSchema, PollMode } from "#/gen/chat/v1/message_pb";

import type { PollInput } from "#/gen/chat/v1/message_pb";

type PollComposerDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  onConfirm: (poll: PollInput) => void;
};

type DateOption = { at: Date; allDay: boolean };

const MAX_OPTIONS = 20;
const modes = ["text", "date"] as const;

// 今日から days 日後の hour 時
const daysLater = (days: number, hour: number) => {
  const date = new Date();
  date.setDate(date.getDate() + days);
  date.setHours(hour, 0, 0, 0);
  return date;
};

/** 投票を組み立てる。日程調整では日時の候補を並べる */
export const PollComposerDialog = ({
  isOpen,
  onOpenChange,
  onConfirm,
}: PollComposerDialogProps) => {
  const { t } = useTranslation();
  const [mode, setMode] = useState<(typeof modes)[number]>("text");
  const [question, setQuestion] = useState("");
  const [labels, setLabels] = useState(["", ""]);
  const [dates, setDates] = useState<DateOption[]>([
    { allDay: false, at: daysLater(1, 19) },
    { allDay: false, at: daysLater(2, 19) },
  ]);
  const [allowMultiple, setAllowMultiple] = useState(false);
  const [anonymous, setAnonymous] = useState(false);
  const [hasDeadline, setHasDeadline] = useState(false);
  const [deadline, setDeadline] = useState(() => daysLater(1, 18));
  const [isSubmitted, setIsSubmitted] = useState(false);

  const filledLabels = labels.map((label) => label.trim()).filter((label) => label !== "");
  const optionCount = mode === "text" ? labels.length : dates.length;
  const isQuestionMissing = question.trim() === "";
  const isOptionsShort = mode === "text" && filledLabels.length < 2;
  const isDeadlinePast = hasDeadline && deadline <= new Date();

  const submit = () => {
    setIsSubmitted(true);
    if (isQuestionMissing || isOptionsShort || isDeadlinePast) {
      return;
    }
    onConfirm(
      create(PollInputSchema, {
        allowMultiple,
        anonymous,
        closesAt: hasDeadline ? timestampFromDate(deadline) : undefined,
        mode: mode === "text" ? PollMode.TEXT : PollMode.DATE,
        options:
          mode === "text"
            ? filledLabels.map((label) => ({ label }))
            : dates.map(({ at, allDay }) => ({ allDay, startsAt: timestampFromDate(at) })),
        question: question.trim(),
      }),
    );
    onOpenChange(false);
  };

  const addOption = () => {
    if (mode === "text") {
      setLabels((current) => [...current, ""]);
    } else {
      setDates((current) => [...current, { allDay: false, at: daysLater(current.length + 1, 19) }]);
    }
  };
  const removeOption = (index: number) => {
    if (mode === "text") {
      setLabels((current) => current.filter((_, i) => i !== index));
    } else {
      setDates((current) => current.filter((_, i) => i !== index));
    }
  };

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title={t("poll.create")}
      size="md"
      footer={
        <>
          <Button
            variant="secondary"
            onPress={() => {
              onOpenChange(false);
            }}
          >
            {t("common.cancel")}
          </Button>
          <Button onPress={submit}>{t("poll.submit")}</Button>
        </>
      }
    >
      <SegmentedControl
        label={t("poll.mode")}
        options={modes.map((value) => ({
          label: t(value === "text" ? "poll.textMode" : "poll.dateMode"),
          value,
        }))}
        value={mode}
        onChange={setMode}
      />
      <TextField
        label={t("poll.question")}
        value={question}
        onChange={setQuestion}
        isRequired
        maxLength={300}
        errorMessage={isSubmitted && isQuestionMissing ? t("poll.questionRequired") : undefined}
      />
      <div className="flex flex-col gap-2">
        {mode === "text"
          ? labels.map((label, index) => (
              // 選択肢は並び替えないため、位置を key にする
              // oxlint-disable-next-line react/no-array-index-key
              <div key={index} className="flex items-end gap-1.5">
                <TextField
                  className="flex-1"
                  label={t("poll.option", { number: index + 1 })}
                  value={label}
                  maxLength={100}
                  onChange={(next) => {
                    setLabels((current) => current.map((value, i) => (i === index ? next : value)));
                  }}
                />
                {optionCount > 2 && (
                  <IconButton
                    label={t("poll.removeOption", { number: index + 1 })}
                    onPress={() => {
                      removeOption(index);
                    }}
                  >
                    <IconX />
                  </IconButton>
                )}
              </div>
            ))
          : dates.map((option, index) => (
              // oxlint-disable-next-line react/no-array-index-key
              <div key={index} className="flex items-end gap-2">
                <DateTimeField
                  label={t("poll.optionDate", { number: index + 1 })}
                  value={option.at}
                  onChange={(at) => {
                    setDates((current) =>
                      current.map((value, i) => (i === index ? { ...value, at } : value)),
                    );
                  }}
                />
                <Checkbox
                  className="mb-2"
                  isSelected={option.allDay}
                  onChange={(allDay) => {
                    setDates((current) =>
                      current.map((value, i) => (i === index ? { ...value, allDay } : value)),
                    );
                  }}
                >
                  {t("poll.allDay")}
                </Checkbox>
                {optionCount > 2 && (
                  <IconButton
                    label={t("poll.removeOption", { number: index + 1 })}
                    onPress={() => {
                      removeOption(index);
                    }}
                  >
                    <IconX />
                  </IconButton>
                )}
              </div>
            ))}
        {isSubmitted && isOptionsShort && (
          <p className="m-0 text-caption text-danger">{t("poll.optionsRequired")}</p>
        )}
        {optionCount < MAX_OPTIONS && (
          <Button variant="ghost" size="sm" className="self-start" onPress={addOption}>
            <IconPlus aria-hidden />
            {t("poll.addOption")}
          </Button>
        )}
      </div>
      <Switch isSelected={allowMultiple} onChange={setAllowMultiple}>
        {t("poll.allowMultiple")}
      </Switch>
      <Switch isSelected={anonymous} onChange={setAnonymous}>
        {t("poll.anonymous")}
      </Switch>
      <Switch isSelected={hasDeadline} onChange={setHasDeadline}>
        {t("poll.deadline")}
      </Switch>
      {hasDeadline && (
        <DateTimeField
          label={t("poll.deadlineLabel")}
          value={deadline}
          onChange={setDeadline}
          errorMessage={isSubmitted && isDeadlinePast ? t("poll.deadlinePast") : undefined}
        />
      )}
    </Dialog>
  );
};
