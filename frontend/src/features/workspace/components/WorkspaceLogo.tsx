type WorkspaceLogoProps = {
  name: string;
};

export const WorkspaceLogo = ({ name }: WorkspaceLogoProps) => {
  const [initial] = new Intl.Segmenter().segment(name);
  return (
    <span
      aria-hidden
      className="grid size-6 shrink-0 place-items-center rounded-[7px] bg-accent text-xs leading-none font-bold text-accent-fg"
    >
      {initial?.segment.toUpperCase()}
    </span>
  );
};
