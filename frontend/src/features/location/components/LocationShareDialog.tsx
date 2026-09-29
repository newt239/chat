import { useTranslation } from "react-i18next";

import { Dialog } from "#/components/ui/Dialog";

import { LocationSharePicker } from "./LocationSharePicker";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type LocationShareDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  onConfirm: (location: MessageLocation) => void;
};

// 開くたびに現在地を取り直す（中身は開いている間だけ描画される）
export const LocationShareDialog = ({
  isOpen,
  onOpenChange,
  onConfirm,
}: LocationShareDialogProps) => {
  const { t } = useTranslation();
  const close = () => {
    onOpenChange(false);
  };
  return (
    <Dialog isOpen={isOpen} onOpenChange={onOpenChange} title={t("location.share.title")} size="md">
      <LocationSharePicker
        onCancel={close}
        onConfirm={(location) => {
          onConfirm(location);
          close();
        }}
      />
    </Dialog>
  );
};
