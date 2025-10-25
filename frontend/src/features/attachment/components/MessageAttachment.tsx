import { useDownloadUrl } from "../api/client";
import { formatFileSize } from "../utils/validator";

import type { Attachment } from "../api/types";

type MessageAttachmentProps = {
  attachment: Attachment;
};

export const MessageAttachment = ({ attachment }: MessageAttachmentProps) => {
  const downloadMutation = useDownloadUrl();

  const handleDownload = async () => {
    try {
      const data = await downloadMutation.mutateAsync(attachment.id);
      window.open(data.downloadUrl, "_blank", "noopener,noreferrer");
    } catch (error) {
      console.error("ダウンロードに失敗しました:", error);
    }
  };

  return (
    <div className="inline-flex items-center gap-2 p-3 bg-gray-50 border border-gray-200 rounded-lg max-w-sm">
      <div className="flex-shrink-0">
        <FileIcon mimeType={attachment.mimeType} />
      </div>
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-gray-900 truncate">
          {attachment.fileName}
        </div>
        <div className="text-xs text-gray-500">{formatFileSize(attachment.sizeBytes)}</div>
      </div>
      <button
        type="button"
        onClick={handleDownload}
        disabled={downloadMutation.isPending}
        className="flex-shrink-0 p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded disabled:opacity-50"
        aria-label="ダウンロード"
      >
        {downloadMutation.isPending ? (
          <svg
            className="w-5 h-5 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        ) : (
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
            />
          </svg>
        )}
      </button>
    </div>
  );
};

type FileIconProps = {
  mimeType: string;
};

// eslint-disable-next-line react/no-multi-comp
const FileIcon = ({ mimeType }: FileIconProps) => {
  const getIconColor = () => {
    if (mimeType.startsWith("image/")) return "text-purple-500";
    if (mimeType.startsWith("video/")) return "text-red-500";
    if (mimeType.startsWith("audio/")) return "text-green-500";
    if (mimeType.includes("pdf")) return "text-red-600";
    if (
      mimeType.includes("document") ||
      mimeType.includes("word") ||
      mimeType.includes("text")
    )
      return "text-blue-500";
    if (mimeType.includes("spreadsheet") || mimeType.includes("excel"))
      return "text-green-600";
    if (mimeType.includes("presentation") || mimeType.includes("powerpoint"))
      return "text-orange-500";
    return "text-gray-500";
  };

  return (
    <svg
      className={`w-8 h-8 ${getIconColor()}`}
      fill="currentColor"
      viewBox="0 0 24 24"
    >
      <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z" />
      <path d="M14 2v6h6" />
    </svg>
  );
};
