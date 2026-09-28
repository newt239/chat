import type { ja } from "./locales/ja";

type Widen<T> = { [K in keyof T]: T[K] extends string ? string : Widen<T[K]> };

// 日本語辞書と同じキーを持つことを他言語の辞書に強制する
export type Messages = Widen<typeof ja>;
