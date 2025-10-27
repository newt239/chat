const MAX_FILE_SIZE = 1024 * 1024 * 1024; // 1GB

type ValidationResult = { valid: true } | { valid: false; error: string };

export const validateFile = (file: File): ValidationResult => {
  if (file.size > MAX_FILE_SIZE) {
    return {
      valid: false,
      error: `ファイルサイズが上限 (1GB) を超えています: ${formatFileSize(file.size)}`,
    };
  }

  if (file.size === 0) {
    return {
      valid: false,
      error: "ファイルが空です",
    };
  }

  return { valid: true };
};

export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 B";

  const units = ["B", "KB", "MB", "GB"];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${units[i]}`;
};
