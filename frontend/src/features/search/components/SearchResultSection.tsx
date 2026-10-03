import type { ReactNode } from "react";

type SearchResultSectionProps = {
  title: string;
  children: ReactNode;
};

export const SearchResultSection = ({ title, children }: SearchResultSectionProps) => (
  <section className="flex flex-col">
    <h3 className="m-0 px-4.5 pt-3 pb-1 text-caption font-semibold text-muted">{title}</h3>
    {children}
  </section>
);
