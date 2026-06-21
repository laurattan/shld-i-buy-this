import { useEffect, useState } from "react";

const PHASES = [
  "Researching.",
  "Researching..",
  "Researching...",
  "Analyzing...",
  "Deciding...",
];

interface LoadingReceiptProps {
  product?: string;
}

export default function LoadingReceipt({ product }: LoadingReceiptProps) {
  const [i, setI] = useState(0);

  useEffect(() => {
    const t = setInterval(() => {
      setI((prev) => (prev + 1) % PHASES.length);
    }, 550);
    return () => clearInterval(t);
  }, []);

  return (
    <article className="loading-receipt paper-grain shadow-paper">
      <hr className="perf" />
      <div className="loading-receipt-body">
        <div className="receipt-no">RECEIPT NO. ----</div>
        <div className="receipt-header-row">
          <span>shld i buy this</span>
          <span>printing…</span>
        </div>
        {product && (
          <p className="receipt-product loading-product">{product}</p>
        )}
        <hr className="perf" style={{ marginTop: 16 }} />
        <p className="loading-status">
          {PHASES[i]}
          <span className="animate-blink" aria-hidden="true">_</span>
        </p>
        <p className="loading-sub">
          consulting the ledger, checking your habits, weighing the price…
        </p>
      </div>
      <svg className="receipt-tear" height="12" viewBox="0 0 800 12" preserveAspectRatio="none" style={{ display: "block", width: "100%" }}>
        <path
          d="M0,4 Q10,0 20,4 Q30,8 40,4 Q50,0 60,4 Q70,8 80,4 Q90,0 100,4 Q110,8 120,4 Q130,0 140,4 Q150,8 160,4 Q170,0 180,4 Q190,8 200,4 Q210,0 220,4 Q230,8 240,4 Q250,0 260,4 Q270,8 280,4 Q290,0 300,4 Q310,8 320,4 Q330,0 340,4 Q350,8 360,4 Q370,0 380,4 Q390,8 400,4 Q410,0 420,4 Q430,8 440,4 Q450,0 460,4 Q470,8 480,4 Q490,0 500,4 Q510,8 520,4 Q530,0 540,4 Q550,8 560,4 Q570,0 580,4 Q590,8 600,4 Q610,0 620,4 Q630,8 640,4 Q650,0 660,4 Q670,8 680,4 Q690,0 700,4 Q710,8 720,4 Q730,0 740,4 Q750,8 760,4 Q770,0 780,4 Q790,8 800,4"
          stroke="var(--ink-faded)"
          strokeWidth="1.5"
          fill="none"
        />
      </svg>
    </article>
  );
}
