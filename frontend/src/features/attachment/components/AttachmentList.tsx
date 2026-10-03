import { AttachmentListItem } from "./AttachmentListItem";

import type { PendingAttachment } from "../api/types";

type AttachmentListProps = {
  attachments: PendingAttachment[];
  onRemove: (id: string) => void;
};

export const AttachmentList = ({ attachments, onRemove }: AttachmentListProps) => (
  <div className="flex flex-wrap gap-1.5 px-2.5 pt-2">
    {attachments.map((attachment) => (
      <AttachmentListItem
        key={attachment.id}
        attachment={attachment}
        onRemove={() => {
          onRemove(attachment.id);
        }}
      />
    ))}
  </div>
);
