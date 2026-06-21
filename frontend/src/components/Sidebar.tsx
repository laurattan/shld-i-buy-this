import { VERDICT_META, type Session } from "../types";

interface SidebarProps {
  open: boolean;
  sessions: Session[];
  activeId: number | null;
  onSelect: (session: Session) => void;
  onClose: () => void;
}

export default function Sidebar({ open, sessions, activeId, onSelect, onClose }: SidebarProps) {
  return (
    <>
      {open && (
        <button
          type="button"
          aria-label="Close sidebar"
          className="sidebar-backdrop"
          onClick={onClose}
        />
      )}

      <aside className={`sidebar${open ? " open" : ""}`} aria-hidden={!open}>
        <div className="sidebar-header">
          <h2 className="sidebar-title">Past Receipts</h2>
          <button type="button" className="sidebar-close" onClick={onClose} aria-label="Close">
            ✕
          </button>
        </div>

        <div className="sidebar-list">
          {sessions.length === 0 ? (
            <p className="sidebar-empty">no receipts yet. print something up.</p>
          ) : (
            sessions.map((s) => {
              const meta = VERDICT_META[s.verdict];
              return (
                <button
                  key={s.id}
                  type="button"
                  className={`sidebar-item-btn shadow-paper${s.id === activeId ? " active" : ""}`}
                  onClick={() => onSelect(s)}
                >
                  <hr className="perf sidebar-item-perf" />
                  <span className="sidebar-item-product">{s.product}</span>
                  <div className="sidebar-item-verdict">
                    <span
                      className="sidebar-verdict-dot"
                      style={{ background: meta.color }}
                      aria-hidden="true"
                    />
                    <span
                      className="sidebar-verdict-label"
                      style={{ color: meta.color }}
                    >
                      {meta.label}
                    </span>
                  </div>
                </button>
              );
            })
          )}
        </div>
      </aside>
    </>
  );
}
