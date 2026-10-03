import { useState } from "react";

import { create } from "@bufbuild/protobuf";
import { IconLoader2 } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { TextField } from "#/components/ui/TextField/TextField";
import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";

import { useCurrentPosition } from "../hooks/useCurrentPosition";
import { formatCoordinates } from "../utils/location";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type LocationShareDialogProps = {
  onClose: () => void;
  onConfirm: (location: MessageLocation) => void;
};

// 開いている間だけ描画し、開くたびに現在地を取り直す
export const LocationShareDialog = ({ onClose, onConfirm }: LocationShareDialogProps) => {
  const { t } = useTranslation();
  const { state, locate } = useCurrentPosition();
  const [label, setLabel] = useState("");

  return (
    <Dialog
      isOpen
      onOpenChange={onClose}
      title={t("location.share.title")}
      size="md"
      footer={
        <>
          <Button variant="secondary" onPress={onClose}>
            {t("common.cancel")}
          </Button>
          <Button
            isDisabled={state.status !== "ready"}
            onPress={() => {
              if (state.status !== "ready") {
                return;
              }
              const { accuracyMeters, latitude, longitude } = state.location;
              onConfirm(
                create(MessageLocationSchema, {
                  accuracyMeters,
                  label: label.trim() || undefined,
                  latitude,
                  longitude,
                }),
              );
              onClose();
            }}
          >
            {t("location.share.confirm")}
          </Button>
        </>
      }
    >
      {state.status === "ready" ? (
        <p className="m-0 font-mono text-caption text-subtle tabular-nums">
          {formatCoordinates(state.location, t)}
        </p>
      ) : state.status === "locating" ? (
        <span className="inline-flex items-center gap-1.5 text-caption text-muted [&_svg]:size-4">
          <IconLoader2 aria-hidden className="animate-spin motion-reduce:animate-none" />
          {t("location.share.locating")}
        </span>
      ) : (
        <span className="flex flex-wrap items-center gap-2 text-caption">
          <span role="alert" className="text-danger">
            {t(`location.share.${state.reason}`)}
          </span>
          <Button variant="secondary" size="sm" onPress={locate}>
            {t("location.share.retry")}
          </Button>
        </span>
      )}
      <TextField
        label={t("location.share.label")}
        placeholder={t("location.share.labelPlaceholder")}
        value={label}
        onChange={setLabel}
        maxLength={100}
      />
    </Dialog>
  );
};
