import { IconBrandX } from "@tabler/icons-react";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles/styles";

import type { OgpData, XPost } from "#/gen/chat/v1/message_pb";

type XPostCardProps = {
  url: string;
  ogp: OgpData;
  post: XPost;
};

// 画像のない投稿では X が投稿者のアイコンを og:image に入れるため、card の種類で出し分ける
export const XPostCard = ({ url, ogp, post }: XPostCardProps) => {
  const { t } = useTranslation();
  const hasMedia = ogp.cardType === "summary_large_image";

  return (
    <Link
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      aria-label={t("link.xPost.open")}
      className={`block w-[min(420px,100%)] overflow-hidden rounded-[10px] border border-border bg-surface font-sans no-underline data-hovered:bg-hover ${focusRing}`}
    >
      <div className="flex flex-col gap-2 px-3 pt-2.5 pb-3">
        <div className="flex items-center gap-2">
          {!hasMedia && ogp.imageUrl && (
            <img src={ogp.imageUrl} alt="" className="size-9 shrink-0 rounded-full object-cover" />
          )}
          <div className="flex min-w-0 flex-1 flex-col">
            <span className="truncate text-[13.5px] font-semibold text-text">
              {post.authorName}
            </span>
            <span className="truncate text-[12px] text-muted">@{post.authorHandle}</span>
          </div>
          <IconBrandX aria-hidden className="size-4 shrink-0 text-text" />
        </div>
        {ogp.description && (
          <p className="m-0 line-clamp-6 text-[13.5px] leading-[1.5] whitespace-pre-wrap text-text">
            {ogp.description}
          </p>
        )}
      </div>
      {hasMedia && ogp.imageUrl && (
        <img
          src={ogp.imageUrl}
          alt=""
          className="block max-h-80 w-full border-t border-border object-cover"
        />
      )}
    </Link>
  );
};
