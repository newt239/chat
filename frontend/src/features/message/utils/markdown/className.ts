// 本文の Markdown の見た目。コードブロックは CodeBlock 側で整える
export const markdownClassName = [
  "min-w-0 text-body leading-[1.65] [overflow-wrap:anywhere]",
  "[&_p]:m-0 [&>*]:m-0 [&>*+*]:mt-1.5",
  "[&_h1]:text-base [&_h1]:font-bold [&_h2]:text-[15px] [&_h2]:font-bold [&_h3]:text-[15px] [&_h3]:font-bold",
  "[&_ol]:pl-[22px] [&_ul]:pl-[22px] [&_li+li]:mt-px",
  // チャンネルリンク（React Aria の Link）はチップの見た目にする
  "[&_a:not([data-rac])]:text-accent-text [&_a:not([data-rac])]:underline [&_a:not([data-rac])]:decoration-1 [&_a:not([data-rac])]:underline-offset-2",
  "[&_del]:text-muted [&_hr]:border-0 [&_hr]:border-t [&_hr]:border-border",
  "[&_blockquote]:border-l-[3px] [&_blockquote]:border-border-strong [&_blockquote]:pl-3 [&_blockquote]:text-muted",
  "[&_:not(pre)>code]:rounded-sm [&_:not(pre)>code]:border [&_:not(pre)>code]:border-border [&_:not(pre)>code]:bg-sunken [&_:not(pre)>code]:px-1 [&_:not(pre)>code]:font-mono [&_:not(pre)>code]:text-[12.5px]",
  "[&_table]:block [&_table]:max-w-full [&_table]:border-collapse [&_table]:overflow-x-auto [&_table]:text-[13px]",
  "[&_td]:border [&_td]:border-border [&_td]:px-2.5 [&_td]:py-0.5 [&_th]:border [&_th]:border-border [&_th]:bg-sunken [&_th]:px-2.5 [&_th]:py-0.5 [&_th]:text-left [&_th]:font-semibold",
].join(" ");
