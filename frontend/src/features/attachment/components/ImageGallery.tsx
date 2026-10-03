import { IconArrowsDiagonal } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { closeDialog, openDialog } from "#/features/layout/utils/overlaySearch";
import { workspaceRoute } from "#/features/layout/utils/workspaceRoute";
import { useOwnsMessageOverlay } from "#/features/message/hooks/useOwnsMessageOverlay";

import { imageBox } from "../utils/imageBox";
import { AttachmentImage } from "./AttachmentImage";
import { Lightbox } from "./Lightbox";

import type { Message, MessageAttachment } from "#/gen/chat/v1/message_pb";

type ImageGalleryProps = {
  images: MessageAttachment[];
  message: Message;
};

const GRID_LIMIT = 4;

// 表示する枚数（2〜4）ごとの並べ方
const gridClassNames = new Map([
  [2, "h-50 grid-cols-2"],
  [3, "h-65 grid-cols-[2fr_1fr] grid-rows-2 [&>:first-child]:row-span-2"],
  [4, "h-75 grid-cols-2 grid-rows-2"],
]);

const tileClassName = `relative block min-h-0 cursor-zoom-in overflow-hidden bg-sunken ${focusRing} [&_img]:transition-transform [&_img]:motion-reduce:transition-none data-hovered:[&_img]:scale-102`;

const imageClassName = "block size-full object-cover";

export const ImageGallery = ({ images, message }: ImageGalleryProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const ownsOverlay = useOwnsMessageOverlay(message.id);
  // 開いている画像は ?image= の添付 ID で表す
  const imageId = workspaceRoute.useSearch({ select: (search) => search.image });
  const found = ownsOverlay ? images.findIndex((image) => image.id === imageId) : -1;
  const openIndex = found === -1 ? null : found;
  const setOpenIndex = (index: number | null) => {
    const image = index === null ? undefined : images[index];
    void navigate({
      // 前後の画像への移動は履歴に積まない
      replace: openIndex !== null && image !== undefined,
      search: image === undefined ? closeDialog : openDialog({ image: image.id }),
      to: ".",
    });
  };

  const renderTiles = () => {
    const [first] = images;
    if (images.length === 1 && first !== undefined) {
      const box = imageBox(first.media?.width, first.media?.height);
      return (
        <Button
          aria-label={t("attachment.expand", { name: first.fileName })}
          onPress={() => {
            setOpenIndex(0);
          }}
          className={cn(tileClassName, "max-w-full self-start rounded-lg border border-border")}
          style={{ aspectRatio: `${box.width} / ${box.height}`, width: box.width }}
        >
          <AttachmentImage
            thumbnail={false}
            attachmentId={first.id}
            alt={first.fileName}
            className={imageClassName}
          />
          {box.crop !== null && (
            <span className="absolute right-1.5 bottom-1.5 flex items-center gap-1 rounded-full bg-media/60 px-1.75 py-0.5 text-caption font-semibold text-media-fg [&_svg]:size-2.75">
              <IconArrowsDiagonal aria-hidden />
              {t(box.crop === "tall" ? "attachment.crop.tall" : "attachment.crop.wide")}
            </span>
          )}
        </Button>
      );
    }
    const shown = images.slice(0, GRID_LIMIT);
    const hiddenCount = images.length - shown.length;
    return (
      <div
        className={cn(
          "grid w-100 max-w-full gap-0.75 overflow-hidden rounded-lg",
          gridClassNames.get(shown.length),
        )}
      >
        {shown.map((image, index) => (
          <Button
            key={image.id}
            aria-label={t("attachment.expand", { name: image.fileName })}
            onPress={() => {
              setOpenIndex(index);
            }}
            className={tileClassName}
          >
            <AttachmentImage
              thumbnail={false}
              attachmentId={image.id}
              alt={image.fileName}
              className={imageClassName}
            />
            {index === GRID_LIMIT - 1 && hiddenCount > 0 && (
              <span className="absolute inset-0 grid place-items-center bg-media/50 text-xl font-bold text-media-fg">
                +{hiddenCount}
              </span>
            )}
          </Button>
        ))}
      </div>
    );
  };

  return (
    <>
      {renderTiles()}
      <Lightbox images={images} message={message} index={openIndex} onIndexChange={setOpenIndex} />
    </>
  );
};
