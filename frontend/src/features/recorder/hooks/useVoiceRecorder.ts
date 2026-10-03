import { useEffect, useRef, useState } from "react";

import { pickRecordingMimeType, recordingFile } from "../utils/recordingFormat";

type RecorderState =
  | { status: "starting" }
  | { status: "recording"; elapsedSeconds: number }
  | { status: "recorded"; file: File; url: string; durationSeconds: number }
  | { status: "failed"; reason: "denied" | "unsupported" };

// 長すぎる録音は自動で止める
const MAX_SECONDS = 10 * 60;

// マウントと同時に録音を始め、アンマウントでマイクと試聴用の URL を解放する
export const useVoiceRecorder = () => {
  const [state, setState] = useState<RecorderState>({ status: "starting" });
  const recorderRef = useRef<MediaRecorder | null>(null);

  useEffect(() => {
    let isDisposed = false;
    let stream: MediaStream | null = null;
    let timer: ReturnType<typeof setInterval> | undefined = undefined;
    let url: string | null = null;
    const releaseMicrophone = () => {
      clearInterval(timer);
      for (const track of stream?.getTracks() ?? []) {
        track.stop();
      }
    };

    const start = async () => {
      if (!("MediaRecorder" in globalThis) || !("mediaDevices" in navigator)) {
        setState({ reason: "unsupported", status: "failed" });
        return;
      }
      try {
        stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      } catch {
        setState({ reason: "denied", status: "failed" });
        return;
      }
      if (isDisposed) {
        releaseMicrophone();
        return;
      }
      const mimeType = pickRecordingMimeType((type) => MediaRecorder.isTypeSupported(type));
      const recorder = new MediaRecorder(stream, mimeType === "" ? {} : { mimeType });
      const chunks: Blob[] = [];
      const startedAt = Date.now();
      const elapsed = () => (Date.now() - startedAt) / 1000;
      recorder.addEventListener("dataavailable", (event) => {
        chunks.push(event.data);
      });
      recorder.addEventListener("stop", () => {
        releaseMicrophone();
        if (isDisposed) {
          return;
        }
        const file = recordingFile(
          new Blob(chunks, { type: recorder.mimeType || mimeType }),
          new Date(startedAt),
        );
        url = URL.createObjectURL(file);
        setState({ durationSeconds: elapsed(), file, status: "recorded", url });
      });
      recorder.start();
      recorderRef.current = recorder;
      setState({ elapsedSeconds: 0, status: "recording" });
      timer = setInterval(() => {
        if (elapsed() >= MAX_SECONDS) {
          recorder.stop();
          return;
        }
        setState({ elapsedSeconds: elapsed(), status: "recording" });
      }, 250);
    };
    void start();

    return () => {
      isDisposed = true;
      if (recorderRef.current?.state === "recording") {
        recorderRef.current.stop();
      }
      releaseMicrophone();
      if (url !== null) {
        URL.revokeObjectURL(url);
      }
    };
  }, []);

  const stop = () => {
    if (recorderRef.current?.state === "recording") {
      recorderRef.current.stop();
    }
  };

  return { state, stop };
};
