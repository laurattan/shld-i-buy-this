import { useEffect, useState } from "react";
import Chat from "./components/Chat";
import Header from "./components/Header";
import InputBox from "./components/InputBox";
import Sidebar from "./components/Sidebar";
import { evaluate, getSessions } from "./api";
import type { Evaluation, Session } from "./types";
import "./App.css";

function App() {
  const [evaluations, setEvaluations] = useState<Evaluation[]>([]);
  const [loading, setLoading] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sessions, setSessions] = useState<Session[]>([]);

  useEffect(() => {
    getSessions().then(setSessions);
  }, []);

  const handleSubmit = async (product: string) => {
    setLoading(true);
    try {
      const { verdict, reasoning } = await evaluate(product);
      setEvaluations((prev) => [...prev, { product, verdict, reasoning }]);
      getSessions().then(setSessions);
    } finally {
      setLoading(false);
    }
  };

  const handleSessionClick = (session: Session) => {
    setEvaluations([{ product: session.product, verdict: session.verdict, reasoning: session.reasoning }]);
  };

  return (
    <div className="app">
      <Header onToggleSidebar={() => setSidebarOpen((prev) => !prev)} />
      {sidebarOpen && <Sidebar sessions={sessions} onSessionClick={handleSessionClick} />}
      <Chat evaluations={evaluations} loading={loading} />
      <InputBox onSubmit={handleSubmit} loading={loading} />
    </div>
  );
}

export default App;
