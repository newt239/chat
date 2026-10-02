type AttachmentUploadState =
  | { status: "idle" }
  | { status: "validating" }
  | { status: "presigning" }
  | { status: "uploading"; progress: number }
  | { status: "completed"; attachmentId: string }
  | { status: "error"; error: string };

export type PendingAttachment = {
  id: string;
  file: File;
  state: AttachmentUploadState;
};
