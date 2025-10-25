import type { components } from "@/lib/api/schema";

export type Attachment = components["schemas"]["Attachment"];
export type PresignRequest = components["schemas"]["PresignRequest"];
export type PresignResponse = components["schemas"]["PresignResponse"];

export type AttachmentUploadState =
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
