import { describe, expect, test } from "vite-plus/test";

import { pickRecordingMimeType, recordingFile } from "./recordingFormat";

describe("pickRecordingMimeType", () => {
  test("webm に対応していればそれを使う", () => {
    expect(pickRecordingMimeType(() => true)).toBe("audio/webm;codecs=opus");
  });

  test("iOS Safari のように mp4 だけなら mp4 を使う", () => {
    expect(pickRecordingMimeType((type) => type === "audio/mp4")).toBe("audio/mp4");
  });

  test("どれにも対応していなければブラウザの既定にする", () => {
    expect(pickRecordingMimeType(() => false)).toBe("");
  });
});

describe("recordingFile", () => {
  const recordedAt = new Date(2026, 8, 29, 9, 5, 3);

  test("コーデックの指定を落とし、MIME に合う拡張子を付ける", () => {
    const file = recordingFile(new Blob(["a"], { type: "audio/webm;codecs=opus" }), recordedAt);
    expect(file.type).toBe("audio/webm");
    expect(file.name).toBe("voice-20260929-090503.webm");
  });

  test("mp4 は m4a として保存する", () => {
    const file = recordingFile(new Blob(["a"], { type: "audio/mp4" }), recordedAt);
    expect(file.type).toBe("audio/mp4");
    expect(file.name).toBe("voice-20260929-090503.m4a");
  });
});
