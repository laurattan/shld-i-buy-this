import type { Evaluation } from "../types";

const verdictColors: Record<Evaluation["verdict"], string> = {
  buy: "green",
  wait: "#b8860b",
  find_alternative: "blue",
  do_not_buy: "red",
};

export default function DecisionDisplay({ product_name, verdict, reasoning }: Evaluation) {
  return (
    <div className="decision-card">
      <p className="decision-product">{product_name}</p>
      <p className="decision-verdict" style={{ color: verdictColors[verdict] }}>
        {verdict}
      </p>
      <p className="decision-reasoning">{reasoning}</p>
    </div>
  );
}
