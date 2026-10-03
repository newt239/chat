import { useId, useState } from "react";

import { IconPhoto } from "@tabler/icons-react";
import { FileTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Button } from "#/components/ui/Button/Button";
import { fieldStyles } from "#/components/ui/styles/styles";

import { ImageCropDialog } from "./ImageCropDialog";
import { useImageUpload } from "./useImageUpload";

import type { ImagePurpose } from "#/gen/chat/v1/image_service_pb";

const IMAGE_TYPES = ["image/png", "image/jpeg", "image/webp", "image/gif"];

type IconImageFieldProps = {
  label: string;
  // 画像がないときの頭文字に使う
  name: string;
  // 画像の URL。空文字なら未設定
  value: string;
  onChange: (url: string) => void;
  purpose: ImagePurpose;
  workspaceId: string | null;
  // stacked は画像を幅いっぱいに出し、その下にボタンを並べる
  layout?: "row" | "stacked";
};

// アイコン画像を選んで切り抜き、アップロードした URL を返す。設定済みならリセットできる
export const IconImageField = ({
  label,
  name,
  value,
  onChange,
  purpose,
  workspaceId,
  layout = "row",
}: IconImageFieldProps) => {
  const { t } = useTranslation();
  const [cropSrc, setCropSrc] = useState<string | null>(null);
  const upload = useImageUpload(purpose, workspaceId);
  const labelId = useId();

  const closeCrop = () => {
    if (cropSrc !== null) {
      URL.revokeObjectURL(cropSrc);
    }
    setCropSrc(null);
  };

  return (
    <div role="group" aria-labelledby={labelId} className="flex flex-col gap-1.5">
      <span id={labelId} className={fieldStyles.label}>
        {label}
      </span>
      {layout === "stacked" && <Avatar name={name} src={value || null} size="fill" />}
      <div className="flex flex-wrap items-center gap-3">
        {layout === "row" && <Avatar name={name} src={value || null} size={56} />}
        <FileTrigger
          acceptedFileTypes={IMAGE_TYPES}
          onSelect={(files) => {
            const file = files?.[0];
            if (file !== undefined) {
              upload.reset();
              setCropSrc(URL.createObjectURL(file));
            }
          }}
        >
          <Button variant="secondary" size="sm">
            <IconPhoto aria-hidden className="size-4" />
            {value === "" ? t("ui.iconImage.select") : t("ui.iconImage.change")}
          </Button>
        </FileTrigger>
        {value !== "" && (
          <Button
            variant="ghost"
            size="sm"
            onPress={() => {
              onChange("");
            }}
          >
            {t("ui.iconImage.reset")}
          </Button>
        )}
      </div>
      {upload.isError && (
        <p role="alert" className="m-0 text-caption text-danger">
          {t("ui.iconImage.uploadFailed")}
        </p>
      )}
      <ImageCropDialog
        key={cropSrc}
        src={cropSrc}
        isPending={upload.isPending}
        onCancel={closeCrop}
        onCrop={(image) => {
          upload.mutate(image, {
            onSettled: closeCrop,
            onSuccess: onChange,
          });
        }}
      />
    </div>
  );
};
