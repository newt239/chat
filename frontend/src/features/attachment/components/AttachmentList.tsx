import { AttachmentListItem } from "./AttachmentListItem";

import type { PendingAttachment } from "../api/types";

type AttachmentListProps = {
  attachments: PendingAttachment[];
  onRemove: (index: number) => void;
};

export const AttachmentList = ({ attachments, onRemove }: AttachmentListProps) => (
  <div className="flex flex-wrap gap-1.5 px-2.5 pt-2">
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
