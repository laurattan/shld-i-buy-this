import { useState } from "react";

interface HeaderProps {
  onToggleSidebar: () => void;
  historyOpen: boolean;
}

export default function Header({ onToggleSidebar, historyOpen }: HeaderProps) {
  const [tooltip, setTooltip] = useState(false);

  return (
    <header className="app-header">
      <h1>shld i buy this</h1>

      <div style={{ position: "relative" }}>
        <button
          type="button"
          className="header-icon-btn"
          onClick={onToggleSidebar}
          onMouseEnter={() => setTooltip(true)}
          onMouseLeave={() => setTooltip(false)}
          aria-label="Toggle past receipts"
          aria-pressed={historyOpen}
        >
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <circle cx="12" cy="12" r="10" />
            <polyline points="12 6 12 12 16 14" />
          </svg>
        </button>
        {tooltip && (
          <div className="header-tooltip">past receipts</div>
        )}
      </div>
    </header>
  );
}
