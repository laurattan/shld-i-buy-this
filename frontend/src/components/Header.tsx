interface HeaderProps {
  onToggleSidebar: () => void;
}

export default function Header({ onToggleSidebar }: HeaderProps) {
  return (
    <header className="header">
      <span className="header-title">shld i buy this</span>
      <button type="button" onClick={onToggleSidebar}>
        History
      </button>
    </header>
  );
}
