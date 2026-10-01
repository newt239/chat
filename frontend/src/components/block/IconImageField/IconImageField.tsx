import { useId, useState } from "react";

import { IconPhoto } from "@tabler/icons-react";
import { FileTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { ImageCropDialog } from "#/components/block/ImageCropDialog/ImageCropDialog";
import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Button } from "#/components/ui/Button/Button";
import { fieldStyles } from "#/components/ui/styles/styles";
import { useImageUpload } from "#/hooks/useImageUpload";

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
};

// アイコン画像を選んで切り抜き、アップロードした URL を返す。設定済みならリセットできる
export const IconImageField = ({
  label,
  name,
  value,
  onChange,
  purpose,
  workspaceId,
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
      <div className="flex flex-wrap items-center gap-3">
        <Avatar name={name} src={value || null} size={56} />
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
