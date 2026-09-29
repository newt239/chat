import { useState } from "react";

import { create } from "@bufbuild/protobuf";
import { IconLoader2 } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { TextField } from "#/components/ui/TextField/TextField";
import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";

import { useCurrentPosition } from "../hooks/useCurrentPosition";
import { formatCoordinates } from "../utils/externalMapUrl";
import { LocationMap } from "./LocationMap";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type LocationSharePickerProps = {
  onConfirm: (location: MessageLocation) => void;
  onCancel: () => void;
};

// 現在地を取得して地図で確かめてから、ラベルを添えて共有する
export const LocationSharePicker = ({ onConfirm, onCancel }: LocationSharePickerProps) => {
  const { t } = useTranslation();
  const { state, locate } = useCurrentPosition();
  const [label, setLabel] = useState("");

  return (
    <>
      <div className="grid h-[220px] place-items-center overflow-hidden rounded-lg border border-border bg-sunken text-caption text-muted">
        {state.status === "ready" ? (
          <LocationMap location={state.location} className="size-full" />
        ) : state.status === "locating" ? (
          <span className="inline-flex items-center gap-1.5 [&_svg]:size-4">
            <IconLoader2 aria-hidden className="animate-spin motion-reduce:animate-none" />
            {t("location.share.locating")}
          </span>
        ) : (
          <span className="flex flex-col items-center gap-2 px-6 text-center">
            <span role="alert" className="text-danger">
              {t(`location.share.${state.reason}`)}
            </span>
            <Button variant="secondary" size="sm" onPress={locate}>
              {t("location.share.retry")}
            </Button>
          </span>
        )}
      </div>
      {state.status === "ready" && (
        <p className="m-0 font-mono text-[11.5px] text-subtle tabular-nums">
          {formatCoordinates(state.location)}
          {state.location.accuracyMeters !== undefined &&
            ` · ${t("location.accuracy", { meters: Math.round(state.location.accuracyMeters) })}`}
        </p>
      )}
      <TextField
        label={t("location.share.label")}
        placeholder={t("location.share.labelPlaceholder")}
        value={label}
        onChange={setLabel}
        maxLength={100}
      />
      <div className="flex justify-end gap-2 pt-1 pb-[max(16px,env(safe-area-inset-bottom))]">
        <Button variant="secondary" onPress={onCancel}>
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
          }}
        >
          {t("location.share.confirm")}
        </Button>
      </div>
    </>
  );
};
