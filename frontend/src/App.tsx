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
  const [error, setError] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sessions, setSessions] = useState<Session[]>([]);

  useEffect(() => {
    getSessions().then(setSessions);
  }, []);

  const handleSubmit = async (product: string) => {
    setLoading(true);
    setError(null);
    try {
      const { product_name, verdict, reasoning } = await evaluate(product);
      setEvaluations((prev) => [...prev, { product, product_name, verdict, reasoning }]);
      getSessions().then(setSessions);
    } catch (e) {
      if (e instanceof Error) {
        const verdict = (e as Error & { verdict?: string }).verdict;
        if (verdict === "do_not_buy") {
          setError("you already decided not to buy this — don't buy it");
        } else if (verdict === "wait") {
          setError("you decided to wait on this — hold off a bit longer");
        } else {
          setError(e.message);
        }
      } else {
        setError("Something went wrong");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSessionClick = (session: Session) => {
    setEvaluations([{ product: session.product, product_name: session.product, verdict: session.verdict, reasoning: session.reasoning }]);
  };

  return (
    <div className="app">
      <Header onToggleSidebar={() => setSidebarOpen((prev) => !prev)} />
      {sidebarOpen && <Sidebar sessions={sessions} onSessionClick={handleSessionClick} />}
      <Chat evaluations={evaluations} loading={loading} error={error} />
      <InputBox onSubmit={handleSubmit} loading={loading} />
    </div>
  );
}

export default App;
