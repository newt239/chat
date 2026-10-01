type WorkspaceLogoProps = {
  name: string;
  // 未設定なら名前の頭文字を出す
  iconUrl: string | undefined;
};

export const WorkspaceLogo = ({ name, iconUrl }: WorkspaceLogoProps) => {
  if (iconUrl) {
    return <img src={iconUrl} alt="" className="size-6 shrink-0 rounded-[7px] object-cover" />;
  }
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
