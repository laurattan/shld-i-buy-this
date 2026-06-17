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
  product: string;
  product_name: string;
  verdict: Verdict;
  reasoning: string;
  alternatives?: Alternative[];
}

export interface Session {
  product: string;
  verdict: Verdict;
  reasoning: string;
  created_at: string;
}
