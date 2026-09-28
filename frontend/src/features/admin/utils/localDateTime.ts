const pad = (value: number) => String(value).padStart(2, "0");

// <input type="datetime-local"> の形式（ブラウザのタイムゾーン、分まで）にする
export const toLocalDateTime = (date: Date) =>
  `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
