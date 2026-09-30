export const EMOJI_IMAGE_TYPES = ["image/png", "image/gif", "image/jpeg", "image/webp"];
const EMOJI_MAX_BYTES = 256 * 1024;
const EMOJI_MAX_SIZE = 128;

export class EmojiImageError extends Error {
  public readonly reason: "type" | "size";

  public constructor(reason: "type" | "size") {
    super(reason);
    this.name = "EmojiImageError";
    this.reason = reason;
  }
}

const toPngBlob = (canvas: HTMLCanvasElement) =>
  new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) {
        resolve(blob);
      } else {
        reject(new EmojiImageError("type"));
      }
    }, "image/png");
  });

// 128px 四方に収まるよう縮めて PNG にする。GIF はアニメーションを残すためそのまま使う
export const prepareEmojiImage = async (file: File): Promise<Blob> => {
  if (!EMOJI_IMAGE_TYPES.includes(file.type)) {
    throw new EmojiImageError("type");
  }
  if (file.type === "image/gif") {
    if (file.size > EMOJI_MAX_BYTES) {
      throw new EmojiImageError("size");
    }
    return file;
  }

  const bitmap = await createImageBitmap(file);
  const scale = Math.min(1, EMOJI_MAX_SIZE / Math.max(bitmap.width, bitmap.height));
  const canvas = document.createElement("canvas");
  canvas.width = Math.max(1, Math.round(bitmap.width * scale));
  canvas.height = Math.max(1, Math.round(bitmap.height * scale));
  canvas.getContext("2d")?.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
  bitmap.close();

  const blob = await toPngBlob(canvas);
  if (blob.size > EMOJI_MAX_BYTES) {
    throw new EmojiImageError("size");
  }
  return blob;
};
