import { useState } from "react";
import type { ReactNode } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { DateTimeField } from "#/components/ui/DateTimeField/DateTimeField";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { useDateFormat } from "#/hooks/useDateFormat";

type ScheduleDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  initialDate: Date;
  onConfirm: (scheduledAt: Date) => void;
  isPending: boolean;
  // 日時の上に出す欄（予約の編集で本文を直すときなど）
  children: ReactNode;
};

// 送信日時を選ぶダイアログ。過去の日時で確定しようとしたらエラーを出す
export const ScheduleDialog = ({
  isOpen,
  onOpenChange,
  title,
  initialDate,
  onConfirm,
  isPending,
  children,
}: ScheduleDialogProps) => {
  const { t } = useTranslation();
  const { timeZone } = useDateFormat();
  const [scheduledAt, setScheduledAt] = useState(initialDate);
  const [isPast, setIsPast] = useState(false);

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title={title}
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
          <Button
            isPending={isPending}
            onPress={() => {
              if (scheduledAt.getTime() <= Date.now()) {
                setIsPast(true);
                return;
              }
              onConfirm(scheduledAt);
            }}
          >
            {t("schedule.dialog.confirm")}
          </Button>
        </>
      }
    >
      {children}
      <DateTimeField
        label={t("schedule.dialog.field")}
        timeZone={timeZone}
        value={scheduledAt}
        onChange={(next) => {
          setScheduledAt(next);
          setIsPast(false);
        }}
        errorMessage={isPast ? t("schedule.dialog.past") : undefined}
      />
    </Dialog>
  );
};
