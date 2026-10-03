import { useState } from "react";

import Cropper from "react-easy-crop";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { Slider } from "#/components/ui/Slider/Slider";

import { cropImage } from "./cropImage";

import type { Area } from "react-easy-crop";

type ImageCropDialogProps = {
  // 切り抜く画像の URL。null なら閉じている
  src: string | null;
  isPending: boolean;
  onCancel: () => void;
  onCrop: (image: Blob) => void;
};

// アイコン用に画像を正方形に切り抜く。ドラッグで位置、スライダーで拡大率を決める
export const ImageCropDialog = ({ src, isPending, onCancel, onCrop }: ImageCropDialogProps) => {
  const { t } = useTranslation();
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [area, setArea] = useState<Area | null>(null);

  return (
    <Dialog
      isOpen={src !== null}
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onCancel();
        }
      }}
      title={t("ui.imageCrop.title")}
      footer={
        <>
          <Button variant="secondary" onPress={onCancel}>
            {t("common.cancel")}
          </Button>
          <Button
            isDisabled={area === null}
            isPending={isPending}
            onPress={() => {
              if (src !== null && area !== null) {
                void cropImage(src, area).then(onCrop);
              }
            }}
          >
            {t("ui.imageCrop.apply")}
          </Button>
        </>
      }
    >
      <div className="relative h-72 overflow-hidden rounded-md bg-sunken">
        {src !== null && (
          <Cropper
            image={src}
            crop={crop}
            zoom={zoom}
            aspect={1}
            onCropChange={setCrop}
            onZoomChange={setZoom}
            onCropComplete={(_, pixels) => {
              setArea(pixels);
            }}
          />
        )}
      </div>
      <Slider
        label={t("ui.imageCrop.zoom")}
        value={zoom}
        minValue={1}
        maxValue={3}
        step={0.05}
        onChange={setZoom}
        onChangeEnd={setZoom}
      />
    </Dialog>
  );
};
