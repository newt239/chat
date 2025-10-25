import { formatFileSize } from "../utils/validator";

import type { PendingAttachment } from "../api/types";

type AttachmentListProps = {
  attachments: PendingAttachment[];
  onRemove: (index: number) => void;
};

export const AttachmentList = ({ attachments, onRemove }: AttachmentListProps) => {
  if (attachments.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-col gap-2 p-2 border-t border-gray-200">
      {attachments.map((attachment, index) => (
        <AttachmentListItem
          key={`${attachment.file.name}-${index}`}
          attachment={attachment}
          onRemove={() => onRemove(index)}
        />
      ))}
    </div>
  );
};

type AttachmentListItemProps = {
  attachment: PendingAttachment;
  onRemove: () => void;
};

// eslint-disable-next-line react/no-multi-comp
const AttachmentListItem = ({ attachment, onRemove }: AttachmentListItemProps) => {
  const { file, state } = attachment;

  return (
    <div className="flex items-center gap-3 p-2 bg-gray-50 rounded border border-gray-200">
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-gray-900 truncate">{file.name}</div>
        <div className="text-xs text-gray-500">
          {formatFileSize(file.size)}
          {state.status === "uploading" && ` - ${state.progress}%`}
          {state.status === "error" && ` - エラー: ${state.error}`}
          {state.status === "completed" && " - 完了"}
        </div>
        {state.status === "uploading" && (
          <div className="mt-1 w-full bg-gray-200 rounded-full h-1.5">
            <div
              className="bg-blue-600 h-1.5 rounded-full transition-all duration-300"
              style={{ width: `${state.progress}%` }}
            />
          </div>
        )}
      </div>
      <button
        type="button"
        onClick={onRemove}
        disabled={state.status === "uploading" || state.status === "presigning"}
        className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="削除"
      >
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
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>
  );
};
