const MAX_FILE_SIZE = 1024 * 1024 * 1024; // 1GB

// 問題がなければ null、あれば理由を返す。文言は呼び出し側で辞書から取る
export const validateFile = (file: File) => {
  if (file.size > MAX_FILE_SIZE) {
    return "tooLarge";
  }
  return file.size === 0 ? "empty" : null;
};
