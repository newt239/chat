import { mediaKindOf } from "./mediaKind";

export type MediaSize = {
  width?: number;
  height?: number;
  durationSeconds?: number;
};

const TIMEOUT_MS = 5_000;

const loadElement = <T extends HTMLImageElement | HTMLMediaElement>(
  element: T,
  eventName: "load" | "loadedmetadata",
  src: string,
) =>
  new Promise<T>((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("timeout"));
    }, TIMEOUT_MS);
    element.addEventListener(eventName, () => {
      clearTimeout(timer);
      resolve(element);
    });
    element.addEventListener("error", () => {
      clearTimeout(timer);
      reject(new Error("failed to load media"));
    });
    element.src = src;
  });

const positive = (value: number) =>
  Number.isFinite(value) && value > 0 ? Math.round(value) : undefined;

const finite = (value: number) => (Number.isFinite(value) ? value : undefined);

// レイアウトを予約するための寸法と再生時間。読み込めないファイルは空で返し、アップロードは続ける
export const measureMedia = async (file: File): Promise<MediaSize> => {
  const kind = mediaKindOf(file.type);
  if (kind === "file") {
    return {};
  }
  const src = URL.createObjectURL(file);
  try {
    if (kind === "image") {
      const image = await loadElement(document.createElement("img"), "load", src);
      return { height: positive(image.naturalHeight), width: positive(image.naturalWidth) };
    }
    if (kind === "audio") {
      const audio = await loadElement(document.createElement("audio"), "loadedmetadata", src);
      return { durationSeconds: finite(audio.duration) };
    }
    const video = await loadElement(document.createElement("video"), "loadedmetadata", src);
    return {
      durationSeconds: finite(video.duration),
      height: positive(video.videoHeight),
      width: positive(video.videoWidth),
    };
  } catch {
    return {};
  } finally {
    URL.revokeObjectURL(src);
  }
};
