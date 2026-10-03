import { canvasToBlob } from "#/lib/canvasToBlob";

export type VideoThumbnail = {
  blob: Blob;
  width: number;
  height: number;
};

// インライン表示（最大 420px）の 2 倍程度あれば高密度の画面でも粗く見えない
const MAX_EDGE = 960;
const JPEG_QUALITY = 0.8;
// 冒頭は暗転していることが多いため、少し進めた位置を使う
const SEEK_SECONDS = 1;

const fitThumbnail = (width: number, height: number) => {
  const scale = Math.min(1, MAX_EDGE / Math.max(width, height));
  return {
    height: Math.max(1, Math.round(height * scale)),
    width: Math.max(1, Math.round(width * scale)),
  };
};

const SEEK_TIMEOUT_MS = 5_000;

// デコードできない形式ではシークが終わらないことがあるため、待つ時間に上限を設ける
const seek = (video: HTMLVideoElement, time: number) =>
  new Promise<void>((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("timeout"));
    }, SEEK_TIMEOUT_MS);
    video.addEventListener(
      "seeked",
      () => {
        clearTimeout(timer);
        resolve();
      },
      { once: true },
    );
    video.currentTime = time;
  });

// メタデータを読み込んだ <video> から 1 フレームを JPEG として切り出す
export const captureVideoFrame = async (video: HTMLVideoElement) => {
  const duration = Number.isFinite(video.duration) ? video.duration : 0;
  await seek(video, Math.min(SEEK_SECONDS, duration / 2));
  const { width, height } = fitThumbnail(video.videoWidth, video.videoHeight);
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
