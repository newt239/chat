import { AttachmentListItem } from "./AttachmentListItem";

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
          key={`${attachment.file.name}-${attachment.file.lastModified}`}
          attachment={attachment}
          onRemove={() => {
            onRemove(index);
          }}
        />
      ))}
    </div>
  );
};
