import { useEffect, useRef } from "react";
import type { Evaluation } from "../types";
import DecisionDisplay from "./DecisionDisplay";

interface ChatProps {
  evaluations: Evaluation[];
  loading: boolean;
}

export default function Chat({ evaluations, loading }: ChatProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [evaluations, loading]);

  return (
    <div className="chat">
      {evaluations.map((evaluation, i) => (
        <DecisionDisplay key={i} {...evaluation} />
      ))}
      {loading && <p className="analyzing">Analyzing...</p>}
      <div ref={bottomRef} />
    </div>
  );
}
