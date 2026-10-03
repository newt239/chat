import { useTranslation } from "react-i18next";

import { markdownClassName } from "../utils/markdown/className";
import { renderMarkdown } from "../utils/markdown/renderer";

type MessagePreviewProps = {
  content: string;
};

export const MessagePreview = ({ content }: MessagePreviewProps) => {
  const { t } = useTranslation();

  return (
    <div className="max-h-45 min-h-15 overflow-y-auto px-3 pt-2.25 pb-0.5">
      {content ? (
        <div className={markdownClassName}>{renderMarkdown(content, [])}</div>
      ) : (
        <p className="m-0 text-body text-subtle">{t("message.composer.previewEmpty")}</p>
      )}
    </div>
  );
};
