import type { components } from "#/lib/api/schema";

export type MessageAttachmentInfo = components["schemas"]["MessageAttachment"];

type AttachmentUploadState =
  | { status: "idle" }
  | { status: "validating" }
  | { status: "presigning" }
  | { status: "uploading"; progress: number }
  | { status: "completed"; attachmentId: string }
  | { status: "error"; error: string };

export type PendingAttachment = {
  file: File;
  state: AttachmentUploadState;
};
