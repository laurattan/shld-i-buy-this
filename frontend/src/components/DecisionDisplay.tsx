import type { Evaluation } from "../types";

const verdictColors: Record<Evaluation["verdict"], string> = {
  buy: "green",
  wait: "#b8860b",
  find_alternative: "blue",
  do_not_buy: "red",
};

export default function DecisionDisplay({ product_name, verdict, reasoning, alternatives }: Evaluation) {
  return (
    <div className="decision-card">
      <p className="decision-product">{product_name}</p>
      <p className="decision-verdict" style={{ color: verdictColors[verdict] }}>
        {verdict}
      </p>
      <div className="decision-reasoning">
        <p>{reasoning}</p>
      </div>
      {verdict === "find_alternative" && alternatives && alternatives.length > 0 && (
        <ul className="decision-alternatives">
          {alternatives.map((alt) => (
            <li key={alt.name} className="decision-alternative-item">
              <span className="alt-name">{alt.name}</span>
              <span className="alt-leader" aria-hidden="true" />
              <span className="alt-price">{alt.price}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
