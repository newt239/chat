// 前後が英数字のとき（10:30:45 など）は絵文字として扱わない
export const CUSTOM_EMOJI_IN_TEXT = /(?<!\w):(?<name>[a-z0-9_-]{1,32}):(?!\w)/gu;

const CUSTOM_EMOJI_ONLY = /^:(?<name>[a-z0-9_-]{1,32}):$/u;

// リアクションの値が :name: ならその名前を返す
export const customEmojiName = (emoji: string) =>
  CUSTOM_EMOJI_ONLY.exec(emoji)?.groups?.name ?? null;

export const toCustomEmojiValue = (name: string) => `:${name}:`;
