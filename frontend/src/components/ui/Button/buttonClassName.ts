import { cn, focusRing, withBaseClassName } from "#/components/ui/styles/styles";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
export type ButtonSize = "md" | "sm";

const variants: Record<ButtonVariant, string> = {
  danger: "bg-danger text-danger-fg data-hovered:opacity-90",
  ghost: "text-muted data-hovered:bg-hover data-hovered:text-text",
  primary: "bg-accent text-accent-fg data-hovered:bg-accent-hover",
  secondary: "border-border-strong bg-surface text-text data-hovered:bg-hover",
};

const sizes: Record<ButtonSize, string> = {
  md: "h-8 px-3 text-body-sm [&_svg]:size-3.75 max-md:h-11 max-md:px-4",
  sm: "h-7 px-2.5 text-xs [&_svg]:size-3.5 max-md:h-11 max-md:px-3",
};

// Button と LinkButton で見た目を共有する
export const buttonClassName = <T>(
  variant: ButtonVariant,
  size: ButtonSize,
  className: string | ((values: T) => string) | undefined,
) =>
  withBaseClassName(
    className,
    cn(
      "inline-flex cursor-pointer items-center justify-center gap-1.5 whitespace-nowrap rounded-md border border-transparent font-sans font-semibold no-underline transition-colors [&_svg]:shrink-0",
      focusRing,
      variants[variant],
      sizes[size],
      "data-disabled:cursor-default data-disabled:border-border data-disabled:bg-sunken data-disabled:text-subtle data-disabled:opacity-100",
    ),
  );
