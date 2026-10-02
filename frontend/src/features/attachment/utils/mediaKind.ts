// ブラウザで表示・再生できる種類だけをメディアとして扱い、それ以外はファイルのカードにする
const PLAYABLE =
  /^(?:image\/(?:png|jpeg|gif|webp|avif|svg\+xml)|video\/(?:mp4|webm|ogg|quicktime)|audio\/(?:mpeg|mp4|aac|ogg|wav|x-wav|webm|flac|x-m4a))$/;

export const mediaKindOf = (mimeType: string) => {
  if (!PLAYABLE.test(mimeType)) {
    return "file";
  }
  if (mimeType.startsWith("image/")) {
    return "image";
  }
  return mimeType.startsWith("video/") ? "video" : "audio";
};
