// navItemClassName が参照する --nav-* の割り当て。同じ一覧を画面ごとに色だけ変えて使う
export const surfaceNavTone =
  "[--dot-ring:var(--c-surface)] [--nav-active-fg:var(--c-accent-text)] [--nav-active:var(--c-accent-soft)] [--nav-fg:var(--c-text)] [--nav-hover:var(--c-hover)] [--nav-muted:var(--c-muted)] [--nav-strong:var(--c-text)] [--tree-line:var(--c-border-strong)]";

// モバイルの画面は本文の配色で、指で押しやすいよう行を高くする。一覧は表示中の画面の下に隠れるので現在地は目立たせない
export const mobileNavTone = `${surfaceNavTone} [--nav-active-fg:var(--nav-fg)] [--nav-active-weight:inherit] [--nav-active:transparent] [--nav-row:44px] [--nav-size:15px]`;

// 一覧の行。リンクにもボタンにも使う。現在地は data-status="active"
export const navItemClassName =
  "relative flex h-(--nav-row,30px) w-full min-w-0 cursor-pointer items-center gap-2 rounded-md pr-1.5 pl-2 text-left text-[length:var(--nav-size,14px)] text-(--nav-fg) no-underline data-hovered:bg-(--nav-hover) data-[status=active]:bg-(--nav-active) data-[status=active]:[font-weight:var(--nav-active-weight,600)] data-[status=active]:text-(--nav-active-fg) [&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:text-(--nav-muted) data-[status=active]:[&_svg]:text-(--nav-active-fg)";

// 一覧の見出しに添える小さなアイコンボタン
export const navIconButtonClassName =
  "size-6 text-(--nav-muted) data-hovered:bg-(--nav-hover) data-hovered:text-(--nav-strong) [&_svg]:size-3.5";
