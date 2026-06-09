package models

import "time"

type Verdict string

const (
	VerdictBuy             Verdict = "buy"
	VerdictWait            Verdict = "wait"
	VerdictFindAlternative Verdict = "find_alternative"
	VerdictDoNotBuy        Verdict = "do_not_buy"
)

// Decision is what the agent returns in memory after evaluating a product.
type Decision struct {
	Label     Verdict `json:"label"`
	Reasoning string  `json:"reasoning"`
}

// UserProfile holds the user's purchase patterns, collected during onboarding.
// Used as context in every agent evaluation.
type UserProfile struct {
	PurchasePatterns string
}

// Session is the persisted record of one evaluation. Written to DB in M3.
type Session struct {
	ID        int
	Product   string
	Verdict   Verdict
	Reasoning string
	CreatedAt time.Time
}
