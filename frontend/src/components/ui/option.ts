import type { Key } from "react-aria-components";

export type Option<T extends string> = {
  value: T;
  label: string;
};

// React Aria の Key から選択肢を引き直すことで、型アサーションなしに T を取り出す
export const findOption = <T extends string>(options: readonly Option<T>[], key: Key | null) =>
  options.find((option) => option.value === key);
