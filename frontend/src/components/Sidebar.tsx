import type { Session } from "../types";

const verdictColors: Record<Session["verdict"], string> = {
  buy: "green",
  wait: "#b8860b",
  find_alternative: "blue",
  do_not_buy: "red",
};

interface SidebarProps {
  sessions: Session[];
  onSessionClick: (session: Session) => void;
}

export default function Sidebar({ sessions, onSessionClick }: SidebarProps) {
  return (
    <aside className="sidebar">
      {sessions.map((session, i) => (
        <div className="sidebar-item" key={i} onClick={() => onSessionClick(session)}>
          <span className="sidebar-product">{session.product}</span>
          <span className="sidebar-verdict" style={{ color: verdictColors[session.verdict] }}>
            {session.verdict}
          </span>
        </div>
      ))}
    </aside>
  );
}
