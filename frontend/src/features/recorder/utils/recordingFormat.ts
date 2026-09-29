// iOS Safari は webm で録音できず mp4（AAC）になるため、対応しているものを先頭から選ぶ
const candidates = ["audio/webm;codecs=opus", "audio/mp4", "audio/aac", "audio/ogg;codecs=opus"];

const extensions = new Map([
  ["audio/aac", "aac"],
  ["audio/mp4", "m4a"],
  ["audio/ogg", "ogg"],
  ["audio/webm", "webm"],
]);

// どれにも対応していなければ空文字（ブラウザの既定）にする
export const pickRecordingMimeType = (isSupported: (mimeType: string) => boolean) =>
  candidates.find((type) => isSupported(type)) ?? "";

const pad = (value: number) => String(value).padStart(2, "0");

// 添付として送れるよう、コーデックの指定を落とした MIME と拡張子のファイルにする
export const recordingFile = (blob: Blob, recordedAt: Date) => {
  const mimeType = blob.type.replace(/;.*$/, "").trim() || "audio/webm";
  const stamp = `${recordedAt.getFullYear()}${pad(recordedAt.getMonth() + 1)}${pad(recordedAt.getDate())}-${pad(recordedAt.getHours())}${pad(recordedAt.getMinutes())}${pad(recordedAt.getSeconds())}`;
  return new File([blob], `voice-${stamp}.${extensions.get(mimeType) ?? "webm"}`, {
    type: mimeType,
  });
};
