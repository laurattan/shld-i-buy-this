import type { Evaluation } from "../types";
import VerdictStamp from "./VerdictStamp";

interface DecisionDisplayProps extends Evaluation {
  receiptNum?: number;
  userName: string;
  fresh?: boolean;
}

function formatStamp(ts: number) {
  return new Date(ts)
    .toLocaleString("en-US", {
      month: "2-digit",
      day: "2-digit",
      year: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    })
    .replace(",", "");
}

function seededRng(seed: number) {
  let s = seed;
  return () => {
    s = (Math.imul(1664525, s) + 1013904223) | 0;
    return Math.abs(s);
  };
}

function hashStr(str: string) {
  let h = 2166136261;
  for (let i = 0; i < str.length; i++) {
    h ^= str.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

function SimpleBarcode({ code }: { code: string }) {
  const rng = seededRng(hashStr(code));
  // guard | data bars | guard
  const segments: { width: number; bar: boolean }[] = [];
  // left guard: thin bar, thin space, thin bar
  segments.push({ width: 3, bar: true }, { width: 3, bar: false }, { width: 3, bar: true });
  // 34 data elements alternating bar/space
  for (let i = 0; i < 34; i++) {
    const w = (rng() % 3) + 2;
    segments.push({ width: w, bar: i % 2 === 0 });
  }
  // right guard
  segments.push({ width: 3, bar: true }, { width: 3, bar: false }, { width: 3, bar: true });

  return (
    <div className="receipt-barcode" aria-hidden="true">
      <div className="receipt-barcode-inner">
        {segments.map((s, i) => (
          <div
            key={i}
            style={{
              width: s.width,
              height: 44,
              background: s.bar ? "var(--ink)" : "transparent",
              flexShrink: 0,
            }}
          />
        ))}
      </div>
    </div>
  );
}

export default function DecisionDisplay({
  product_name,
  verdict,
  reasoning,
  alternatives,
  receiptNum,
  userName,
  fresh = false,
  created_at,
}: DecisionDisplayProps) {
  return (
    <article className="receipt paper-grain shadow-paper">
      <hr className="perf" />
      <div className="receipt-body">
        <div className="receipt-no">
          RECEIPT NO. {receiptNum !== undefined ? String(receiptNum).padStart(4, "0") : "----"}
        </div>

        <div className="receipt-header-row">
          <span>shld i buy this</span>
          {created_at && <span>{formatStamp(created_at)}</span>}
        </div>

        <h2 className="receipt-product">{product_name}</h2>

        <div className="receipt-stamp-row">
          <div className="receipt-stamp-offset">
            <VerdictStamp verdict={verdict} animate={fresh} />
          </div>
        </div>

        <hr className="perf" />

        <p className="receipt-reasoning">{reasoning}</p>

        {verdict === "find_alternative" && alternatives && alternatives.length > 0 && (
          <ul className="receipt-alternatives">
            {alternatives.map((alt) => (
              <li key={alt.name} className="receipt-alt-item">
                <span className="receipt-alt-name">{alt.name}</span>
                <span className="receipt-alt-leader" aria-hidden="true" />
                <span className="receipt-alt-price">{alt.price}</span>
              </li>
            ))}
          </ul>
        )}

        <hr className="perf" style={{ marginTop: 20 }} />

        <div className="receipt-footer">
          <SimpleBarcode code={product_name} />
          <div className="receipt-sign-off font-accent">
            <p>purchased by: {userName || "—"}</p>
            <p>thank you for shopping with shld i buy this.</p>
          </div>
        </div>
      </div>

      <svg height="12" viewBox="0 0 800 12" preserveAspectRatio="none" style={{ display: "block", width: "100%" }}>
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
