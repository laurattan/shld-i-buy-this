export type Verdict = "buy" | "wait" | "find_alternative" | "do_not_buy";

export interface Alternative {
  name: string;
  price: string;
}

export interface Profile {
  name: string;
  purchase_patterns: string;
}

export interface Evaluation {
  id?: number;
  product: string;
  product_name: string;
  verdict: Verdict;
  reasoning: string;
  alternatives?: Alternative[];
  created_at?: number;
  isNew?: boolean;
}

export interface Session {
  id: number;
  product: string;
  product_name: string;
  verdict: Verdict;
  reasoning: string;
  alternatives?: Alternative[];
  created_at: string;
}

export const VERDICT_META: Record<Verdict, { label: string; color: string }> = {
  buy: { label: "BUY", color: "var(--stamp-green)" },
  wait: { label: "WAIT", color: "var(--stamp-amber)" },
  find_alternative: { label: "FIND ALT", color: "var(--stamp-blue)" },
  do_not_buy: { label: "DON'T BUY", color: "var(--stamp-red)" },
};

export const VERDICT_ROTATIONS: Record<Verdict, number> = {
  buy: -8,
  wait: -12,
  find_alternative: -6,
  do_not_buy: -10,
};
