import { cn } from "#/components/ui/styles/styles";

type SkeletonProps = {
  // 大きさと形は利用側で指定する（例: "h-4 w-32"、"size-8 rounded-full"）
  className: string;
};

export const Skeleton = ({ className }: SkeletonProps) => (
  <div
    aria-hidden
    className={cn("animate-pulse rounded-md bg-sunken motion-reduce:animate-none", className)}
  />
);
