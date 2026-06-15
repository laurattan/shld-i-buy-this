export type Verdict = "buy" | "wait" | "find_alternative" | "do_not_buy";

export interface Evaluation {
  product: string;
  verdict: Verdict;
  reasoning: string;
}

export interface Session {
  product: string;
  verdict: Verdict;
  reasoning: string;
  created_at: string;
}
