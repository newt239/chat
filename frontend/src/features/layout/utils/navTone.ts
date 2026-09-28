// NavLink などが参照する --nav-* の割り当て。同じ一覧をサイドバーとモバイルの画面で色だけ変えて使う
export const sidebarNavTone =
  "[--dot-ring:var(--c-side)] [--nav-active-fg:var(--c-side-active-fg)] [--nav-active:var(--c-side-active)] [--nav-fg:var(--c-side-fg)] [--nav-hover:var(--c-side-hover)] [--nav-muted:var(--c-side-muted)] [--nav-strong:var(--c-side-strong)]";

export const mobileNavTone =
  "[--dot-ring:var(--c-surface)] [--nav-active-fg:var(--c-accent-text)] [--nav-active:var(--c-accent-soft)] [--nav-fg:var(--c-text)] [--nav-hover:var(--c-hover)] [--nav-muted:var(--c-muted)] [--nav-row:44px] [--nav-size:15px] [--nav-strong:var(--c-text)]";

// 一覧の行。リンクは NavLink、ボタンの行はこのクラスを直接使う。現在地は data-status="active"
export const navItemClassName =
  "relative flex h-(--nav-row,30px) w-full min-w-0 cursor-pointer items-center gap-2 rounded-[6px] pr-1.5 pl-2 text-left text-[length:var(--nav-size,14px)] text-(--nav-fg) no-underline data-hovered:bg-(--nav-hover) data-[status=active]:bg-(--nav-active) data-[status=active]:font-semibold data-[status=active]:text-(--nav-active-fg) [&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:text-(--nav-muted) data-[status=active]:[&_svg]:text-(--nav-active-fg)";
