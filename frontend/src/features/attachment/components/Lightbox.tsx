import { useEffect } from "react";

import { formatDateTime } from "@chat/i18n";
import { IconChevronLeft, IconChevronRight, IconDownload, IconX } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { AnimatePresence, motion } from "motion/react";
import { Button, Dialog, Modal, ModalOverlay } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { cn, focusRing } from "#/components/ui/styles";
import { transitions } from "#/lib/motion";
import { toDate } from "#/lib/timestamp";
import { useIsMobile } from "#/lib/useMediaQuery";
import { preferencesAtom } from "#/providers/store/preferences";

import { useAttachmentUrl } from "../api/client";
import { AttachmentImage } from "./AttachmentImage";

import type { Message, MessageAttachment } from "#/gen/chat/v1/message_pb";

type LightboxProps = {
  images: MessageAttachment[];
  message: Message;
  // null のときは閉じている
  index: number | null;
  onIndexChange: (index: number | null) => void;
};

const MotionModalOverlay = motion.create(ModalOverlay);
// 高さが幅の 2.2 倍を超える画像は、縮めずにスクロールで見せる
const TALL_RATIO = 2.2;
const SWIPE_THRESHOLD_PX = 60;

const controlClassName = `grid size-9 shrink-0 cursor-pointer place-items-center rounded-md text-media-fg data-hovered:bg-media-fg/12 [&_svg]:size-5 ${focusRing}`;

export const Lightbox = ({ images, message, index, onIndexChange }: LightboxProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const isMobile = useIsMobile();
  const image = index === null ? undefined : images[index];
  const { data: url } = useAttachmentUrl(image?.id ?? null);
  const hasMany = images.length > 1;

  const move = (delta: number) => {
    if (index !== null) {
      onIndexChange((index + delta + images.length) % images.length);
    }
  };

  // 左右キーで前後の画像へ移る。Esc で閉じるのは Modal が扱う
  useEffect(() => {
    if (index === null || !hasMany) {
      return undefined;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
        onIndexChange(
          (index + (event.key === "ArrowLeft" ? -1 : 1) + images.length) % images.length,
        );
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [index, hasMany, images.length, onIndexChange]);

  const isTall =
    image?.media?.width !== undefined &&
    image.media.height !== undefined &&
    image.media.height / image.media.width > TALL_RATIO;
  const authorName = message.user?.displayName ?? "";

  return (
    <AnimatePresence>
      {index !== null && image !== undefined && (
        <MotionModalOverlay
          isOpen
          isDismissable
          onOpenChange={(isOpen) => {
            if (!isOpen) {
              onIndexChange(null);
            }
          }}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={transitions.base}
          className="fixed inset-0 z-[300] bg-media/92"
        >
          <Modal className="size-full">
            <Dialog
              aria-label={t("attachment.lightbox.label")}
              className="flex size-full flex-col font-sans text-media-fg outline-none"
            >
              <header className="flex items-center gap-2.5 px-3.5 pt-[max(10px,env(safe-area-inset-top))] pb-2.5 text-[13px]">
                <Avatar name={authorName} src={message.user?.avatarUrl} size={30} />
                <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
                  <span className="truncate">{image.fileName}</span>
                  <small className="truncate text-media-fg/60">
                    {authorName} · {formatDateTime(toDate(message.createdAt), locale)}
                  </small>
                </span>
                {hasMany && (
                  <span className="font-mono text-xs text-media-fg/70 tabular-nums">
                    {t("attachment.lightbox.position", { index: index + 1, total: images.length })}
                  </span>
                )}
                {url !== undefined && (
                  <a
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    aria-label={t("attachment.download")}
                    className={controlClassName}
                  >
                    <IconDownload aria-hidden />
                  </a>
                )}
                <Button
                  aria-label={t("common.close")}
                  onPress={() => {
                    onIndexChange(null);
                  }}
                  className={controlClassName}
                >
                  <IconX aria-hidden />
                </Button>
              </header>

              <div
                className={cn(
                  "flex min-h-0 flex-1 justify-center gap-2 px-2 pb-4 max-md:px-0",
                  isTall ? "items-start overflow-y-auto" : "items-center",
                )}
                onPointerDown={(event) => {
                  // 画像の外側を押したら閉じる
                  if (event.target === event.currentTarget) {
                    onIndexChange(null);
                  }
                }}
              >
                {hasMany && !isMobile && (
                  <Button
                    aria-label={t("attachment.lightbox.previous")}
                    onPress={() => {
                      move(-1);
                    }}
                    className={cn(controlClassName, "self-center")}
                  >
                    <IconChevronLeft aria-hidden />
                  </Button>
                )}
                <motion.div
                  key={image.id}
                  initial={{ opacity: 0, scale: 0.97 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={transitions.spring}
                  drag={hasMany && isMobile ? "x" : false}
                  dragSnapToOrigin
                  onDragEnd={(_, info) => {
                    if (Math.abs(info.offset.x) > SWIPE_THRESHOLD_PX) {
                      move(info.offset.x < 0 ? 1 : -1);
                    }
                  }}
                  className={cn(
                    "flex min-w-0 justify-center",
                    isTall
                      ? "w-[min(420px,100%)]"
                      : "max-h-full max-w-[calc(100%-100px)] max-md:max-w-full",
                  )}
                >
                  <AttachmentImage
                    attachmentId={image.id}
                    alt={image.fileName}
                    className={cn(
                      "rounded-sm object-contain",
                      isTall ? "h-auto w-full" : "max-h-[calc(100dvh-140px)] max-w-full",
                    )}
                  />
                </motion.div>
                {hasMany && !isMobile && (
                  <Button
                    aria-label={t("attachment.lightbox.next")}
                    onPress={() => {
                      move(1);
                    }}
                    className={cn(controlClassName, "self-center")}
                  >
                    <IconChevronRight aria-hidden />
                  </Button>
                )}
              </div>

              {hasMany && isMobile && (
                <div className="flex justify-center gap-1.5 pb-[max(24px,env(safe-area-inset-bottom))]">
                  {images.map((item, itemIndex) => (
                    <Button
                      key={item.id}
                      aria-label={t("attachment.lightbox.page", { index: itemIndex + 1 })}
                      onPress={() => {
                        onIndexChange(itemIndex);
                      }}
                      className={cn(
                        "size-[7px] cursor-pointer rounded-full",
                        itemIndex === index ? "bg-media-fg" : "bg-media-fg/35",
                      )}
                    />
                  ))}
                </div>
              )}
              {isTall && (
                <p className="m-0 pb-2.5 text-center text-xs text-media-fg/60">
                  {t("attachment.lightbox.tallHint")}
                </p>
              )}
            </Dialog>
          </Modal>
        </MotionModalOverlay>
      )}
    </AnimatePresence>
  );
};
