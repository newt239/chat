const MAX_FILE_SIZE = 1024 * 1024 * 1024; // 1GB

export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB"];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${(bytes / k ** i).toFixed(2)} ${units[i]}`;
};

// 問題がなければ null、あれば理由を返す。文言は呼び出し側で辞書から取る
export const validateFile = (file: File) => {
  if (file.size > MAX_FILE_SIZE) {
    return "tooLarge";
  }
  return file.size === 0 ? "empty" : null;
};
