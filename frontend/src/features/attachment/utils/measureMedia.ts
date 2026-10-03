import { canvasToBlob } from "#/lib/canvasToBlob";

import { mediaKindOf } from "./mediaKind";

type MediaSize = {
  width?: number;
  height?: number;
  durationSeconds?: number;
  // 動画の再生前に表示する画像。切り出せなければ付けない
  thumbnail?: { blob: Blob; width: number; height: number };
};

const TIMEOUT_MS = 5_000;
// インライン表示（最大 420px）の 2 倍程度あれば高密度の画面でも粗く見えない
const THUMBNAIL_MAX_EDGE = 960;
const JPEG_QUALITY = 0.8;
// 冒頭は暗転していることが多いため、少し進めた位置を使う
const SEEK_SECONDS = 1;

// 読み込みやシークが終わらない形式もあるため、待つ時間に上限を設ける
const waitForEvent = (
  element: HTMLImageElement | HTMLMediaElement,
  eventName: "load" | "loadedmetadata" | "seeked",
  start: () => void,
) =>
  new Promise<void>((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("timeout"));
    }, TIMEOUT_MS);
    element.addEventListener(
      eventName,
      () => {
        clearTimeout(timer);
        resolve();
      },
      { once: true },
    );
    element.addEventListener(
      "error",
      () => {
        clearTimeout(timer);
        reject(new Error("failed to load media"));
      },
      { once: true },
    );
    start();
  });

const positive = (value: number) =>
  Number.isFinite(value) && value > 0 ? Math.round(value) : undefined;

const finite = (value: number) => (Number.isFinite(value) ? value : undefined);

// メタデータを読み込んだ <video> から 1 フレームを JPEG として切り出す
const captureVideoFrame = async (video: HTMLVideoElement, duration: number) => {
  await waitForEvent(video, "seeked", () => {
    video.currentTime = Math.min(SEEK_SECONDS, duration / 2);
  });
  const scale = Math.min(1, THUMBNAIL_MAX_EDGE / Math.max(video.videoWidth, video.videoHeight));
  const width = Math.max(1, Math.round(video.videoWidth * scale));
  const height = Math.max(1, Math.round(video.videoHeight * scale));
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;
  const context = canvas.getContext("2d");
  if (context === null) {
    throw new Error("canvas is not supported");
  }
  context.drawImage(video, 0, 0, width, height);
  return { blob: await canvasToBlob(canvas, "image/jpeg", JPEG_QUALITY), height, width };
};

// レイアウトを予約するための寸法と再生時間。読み込めないファイルは空で返し、アップロードは続ける
export const measureMedia = async (file: File): Promise<MediaSize> => {
  const kind = mediaKindOf(file.type);
  if (kind === "file") {
    return {};
  }
  const src = URL.createObjectURL(file);
  try {
    if (kind === "image") {
      const image = document.createElement("img");
      await waitForEvent(image, "load", () => {
        image.src = src;
      });
      return { height: positive(image.naturalHeight), width: positive(image.naturalWidth) };
    }
    if (kind === "audio") {
      const audio = document.createElement("audio");
      await waitForEvent(audio, "loadedmetadata", () => {
        audio.src = src;
      });
      return { durationSeconds: finite(audio.duration) };
    }
    const video = document.createElement("video");
    video.muted = true;
    video.playsInline = true;
    video.preload = "auto";
    await waitForEvent(video, "loadedmetadata", () => {
      video.src = src;
    });
    const size = {
      durationSeconds: finite(video.duration),
      height: positive(video.videoHeight),
      width: positive(video.videoWidth),
    };
    if (size.width === undefined || size.height === undefined) {
      return size;
    }
    const thumbnail = await captureVideoFrame(video, size.durationSeconds ?? 0).catch(
      () => undefined,
    );
    return { ...size, thumbnail };
  } catch {
    return {};
  } finally {
    URL.revokeObjectURL(src);
  }
};
