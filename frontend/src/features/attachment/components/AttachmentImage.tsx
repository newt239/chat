import { IconPhotoOff } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { cn } from "#/components/ui/styles/styles";

import { useAttachmentUrl } from "../hooks/useAttachmentUrl";

type AttachmentImageProps = {
  attachmentId: string;
  // true なら動画のサムネイルを表示する
  thumbnail: boolean;
  alt: string;
  className: string;
};

// 署名付き URL を取得して表示する。取得中は枠だけを出してレイアウトを保つ
export const AttachmentImage = ({
  attachmentId,
  thumbnail,
  alt,
  className,
}: AttachmentImageProps) => {
  const { t } = useTranslation();
  const { data: url, isError } = useAttachmentUrl(attachmentId, thumbnail);

  if (isError) {
    return (
      <span
        role="img"
        aria-label={t("attachment.loadFailed")}
        className={cn("grid place-items-center text-subtle [&_svg]:size-6", className)}
      >
        <IconPhotoOff aria-hidden />
      </span>
    );
  }
  if (url === undefined) {
    return <span aria-busy className={cn("block animate-pulse bg-sunken", className)} />;
  }
  return <img src={url} alt={alt} draggable={false} className={className} />;
};
