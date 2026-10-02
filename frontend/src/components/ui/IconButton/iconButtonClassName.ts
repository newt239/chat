import { focusRing } from "#/components/ui/styles/styles";

// IconToggleButton と共有する見た目。モバイルでは WCAG のタップ領域に合わせて 44px にする
export const iconButtonClassName = `relative inline-grid size-[30px] shrink-0 cursor-pointer place-items-center rounded-md text-muted transition-colors data-disabled:cursor-default data-disabled:text-subtle data-hovered:bg-hover data-hovered:text-text data-pressed:bg-hover [&_svg]:size-[18px] max-md:size-11 max-md:[&_svg]:size-[22px] ${focusRing}`;
