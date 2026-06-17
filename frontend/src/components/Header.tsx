interface HeaderProps {
  onToggleSidebar: () => void;
  onOpenProfile: () => void;
}

export default function Header({ onToggleSidebar, onOpenProfile }: HeaderProps) {
  return (
    <header className="header">
      <span className="header-title">shld i buy this</span>
      <button type="button" onClick={onOpenProfile}>
        Profile
      </button>
      <button type="button" onClick={onToggleSidebar}>
        History
      </button>
    </header>
  );
}
