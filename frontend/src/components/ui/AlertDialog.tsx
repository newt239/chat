import type { ReactNode } from "react";

import { Heading } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "./Button";
import { DialogFrame } from "./DialogFrame";

type AlertDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  children: ReactNode;
  confirmLabel: string;
  onConfirm: () => void;
  tone?: "default" | "danger";
  isPending?: boolean;
};

// 確認用のダイアログ。モバイルでも全画面にせず中央に出し、外側のクリックでは閉じない
export const AlertDialog = ({
  isOpen,
  onOpenChange,
  title,
  children,
  confirmLabel,
  onConfirm,
  tone = "default",
  isPending,
}: AlertDialogProps) => {
  const { t } = useTranslation();
  return (
    <DialogFrame
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      layout="center"
      role="alertdialog"
      className="max-w-[440px]"
    >
      <Heading slot="title" className="m-0 px-[18px] pt-4 text-base font-bold">
        {title}
      </Heading>
      <div className="flex flex-col gap-3 px-[18px] pt-2.5 pb-1 text-[13.5px] text-muted">
        {children}
      </div>
      <footer className="flex justify-end gap-2 px-[18px] pt-3.5 pb-4">
        <Button
          variant="secondary"
          onPress={() => {
            onOpenChange(false);
          }}
        >
          {t("common.cancel")}
        </Button>
        <Button
          variant={tone === "danger" ? "danger" : "primary"}
          isPending={isPending}
          onPress={onConfirm}
        >
          {confirmLabel}
        </Button>
      </footer>
    </DialogFrame>
  );
};
