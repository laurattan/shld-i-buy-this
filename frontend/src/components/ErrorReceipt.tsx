interface ErrorReceiptProps {
  product: string;
  message: string;
  isCooldown: boolean;
  isRetryable: boolean;
  onRetry?: () => void;
}

export default function ErrorReceipt({ product, message, isCooldown, isRetryable, onRetry }: ErrorReceiptProps) {
  const now = new Date().toLocaleString("en-US", {
    month: "2-digit", day: "2-digit", year: "2-digit",
    hour: "2-digit", minute: "2-digit", hour12: false,
  }).replace(",", "");

  return (
    <article className={`receipt paper-grain shadow-paper error-receipt${isCooldown ? " error-receipt--cooldown" : ""}`}>
      <hr className="perf" />
      <div className="receipt-body">
        <div className="receipt-header-row">
          <span>shld i buy this</span>
          <span>{now}</span>
        </div>

        <h2 className="receipt-product">{product}</h2>

        <div className="receipt-stamp-row">
          <div className="receipt-stamp-offset">
            {isRetryable && (
              <button type="button" className="error-receipt-retry" onClick={onRetry}>
                try again
              </button>
            )}
          </div>
        </div>

        <hr className="perf" style={{ margin: "0 0 20px" }} />

        <p className={`receipt-reasoning${isCooldown ? " error-receipt-message--info" : " error-receipt-message--error"}`}>
          {message}
        </p>
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
