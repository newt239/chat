// 本文の Markdown の見た目。コードブロックは CodeBlock 側で整える
export const markdownClassName = [
  "min-w-0 text-body leading-relaxed [overflow-wrap:anywhere]",
  "[&_p]:m-0 [&>*]:m-0 [&>*+*]:mt-1.5",
  "[&_h1]:text-base [&_h2]:text-title [&_h3]:text-title [&_:is(h1,h2,h3)]:leading-snug [&_:is(h1,h2,h3,h4,h5,h6)]:font-bold",
  "[&_:is(h4,h5,h6)]:text-body",
  "[&_ol]:list-decimal [&_ol]:pl-5.5 [&_ul]:list-disc [&_ul]:pl-5.5 [&_li+li]:mt-px [&_li>:is(ul,ol)]:mt-px",
  // タスクリストはチェックボックスを行頭の記号の位置に置く
  "[&_li:has(>input)]:-ml-5 [&_li:has(>input)]:list-none [&_li>input]:mr-1.5 [&_li>input]:ml-0 [&_li>input]:align-[-2px] [&_li>input]:accent-accent",
  // チャンネルリンク（React Aria の Link）はチップの見た目にする
  "[&_a:not([data-rac])]:text-accent-text [&_a:not([data-rac])]:underline [&_a:not([data-rac])]:decoration-1 [&_a:not([data-rac])]:underline-offset-2",
  "[&_del]:text-muted [&_hr]:border-0 [&_hr]:border-t [&_hr]:border-border",
  "[&_blockquote]:border-l-3 [&_blockquote]:border-border-strong [&_blockquote]:pl-3 [&_blockquote]:text-muted [&_blockquote>*+*]:mt-1.5",
  "[&_:not(pre)>code]:rounded-sm [&_:not(pre)>code]:border [&_:not(pre)>code]:border-border [&_:not(pre)>code]:bg-sunken [&_:not(pre)>code]:px-1 [&_:not(pre)>code]:font-mono [&_:not(pre)>code]:text-mono",
  "[&_table]:block [&_table]:max-w-full [&_table]:border-collapse [&_table]:overflow-x-auto [&_table]:text-body-sm",
  "[&_:is(td,th)]:border [&_:is(td,th)]:border-border [&_:is(td,th)]:px-2.5 [&_:is(td,th)]:py-0.5 [&_:is(td,th)]:text-left [&_:is(td,th)]:whitespace-nowrap [&_th]:bg-sunken [&_th]:font-semibold",
].join(" ");

// 絵文字だけの短い投稿
export const jumboClassName = "text-emoji leading-tight";
