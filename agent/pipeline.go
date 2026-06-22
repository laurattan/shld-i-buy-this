package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laurattan/shld-i-buy-this/models"
)

type Analysis struct {
	ProductName  string
	Verdict      models.Verdict
	Reasoning    string
	Alternatives []models.Alternative
}

// research investigates the product and forms a verdict + reasoning in a
// single combined, profile-aware call. If the verdict is find_alternative,
// it brainstorms and concurrently verifies 5 candidate alternatives
// before returning.
func research(ctx context.Context, product string, profile models.UserProfile) (*Analysis, error) {
	systemPrompt, err := buildResearchPrompt(profile)
	if err != nil {
		return nil, err
	}

	messages := []map[string]any{
		{"role": "user", "content": product},
	}

	text, err := callClaude(ctx, systemPrompt, messages, true)
	if err != nil {
		return nil, fmt.Errorf("research call: %w", err)
	}

	jsonText, err := extractJSON(text)
	if err != nil {
		return nil, fmt.Errorf("extracting research JSON: %w", err)
	}

	var result struct {
		ProductName       string         `json:"product_name"`
		Verdict           models.Verdict `json:"verdict"`
		Reasoning         string         `json:"reasoning"`
		NeedsAlternatives bool           `json:"needs_alternatives"`
	}
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("parsing research JSON: %w (raw: %s)", err, text)
	}

	switch result.Verdict {
	case models.VerdictBuy, models.VerdictWait, models.VerdictFindAlternative, models.VerdictDoNotBuy:
	default:
		return nil, fmt.Errorf("invalid verdict: %q", result.Verdict)
	}

	analysis := &Analysis{
		ProductName: result.ProductName,
		Verdict:     result.Verdict,
		Reasoning:   result.Reasoning,
	}

	if result.NeedsAlternatives {
		candidates, err := brainstormCandidates(ctx, *analysis, profile)
		if err != nil {
			log.Printf("[research] brainstorm failed: %v", err)
		} else {
			analysis.Alternatives = researchAlternatives(ctx, candidates)
		}
	}

	return analysis, nil
}

// brainstormCandidates suggests 5 specific alternative product names,
// informed by the user's profile.
func brainstormCandidates(ctx context.Context, analysis Analysis, profile models.UserProfile) ([]string, error) {
	systemPrompt, err := buildCandidatesPrompt(analysis, profile)
	if err != nil {
		return nil, err
	}

	messages := []map[string]any{
		{"role": "user", "content": "Suggest 5 specific alternative product names."},
	}

	text, err := callClaude(ctx, systemPrompt, messages, false)
	if err != nil {
		return nil, fmt.Errorf("brainstorm call: %w", err)
	}

	jsonText, err := extractJSON(text)
	if err != nil {
		return nil, fmt.Errorf("extracting brainstorm JSON: %w", err)
	}

	var result struct {
		Candidates []string `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("parsing brainstorm JSON: %w (raw: %s)", err, text)
	}

	return result.Candidates, nil
}

// researchAlternatives verifies each candidate's real price/quality in
// parallel, one goroutine per candidate, each bounded by a 45s timeout.
// Returns only the candidates that were successfully verified.
func researchAlternatives(ctx context.Context, candidates []string) []models.Alternative {
	results := make([]models.Alternative, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate string) {
			defer wg.Done()
			candidateCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			defer cancel()
			alt, err := researchOneCandidate(candidateCtx, candidate)
			if err != nil {
				log.Printf("[research] candidate %q failed or timed out: %v", candidate, err)
				return
			}
			results[i] = *alt
		}(i, candidate)
	}
	wg.Wait()

	parsePrice := func(s string) float64 {
		s = strings.NewReplacer("$", "", ",", "").Replace(s)
		v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
		return v
	}

	var out []models.Alternative
	for _, a := range results {
		if a.Name == "" {
			continue
		}
		if parsePrice(a.Price) == math.MaxFloat64 {
			log.Printf("[research] dropping candidate %q: unparseable price %q", a.Name, a.Price)
			continue
		}
		out = append(out, a)
	}

	sort.Slice(out, func(i, j int) bool {
		return parsePrice(out[i].Price) < parsePrice(out[j].Price)
	})
	if len(out) > 3 {
		out = out[:3]
	}
	log.Printf("[research] verified %d/%d candidates", len(out), len(candidates))
	return out
}

// researchOneCandidate makes a single Claude call with web search to verify
// one candidate's standard-size price.
func researchOneCandidate(ctx context.Context, candidate string) (*models.Alternative, error) {
	systemPrompt, err := buildCandidateResearchPrompt()
	if err != nil {
		return nil, err
	}

	messages := []map[string]any{
		{"role": "user", "content": candidate},
	}

	text, err := callClaude(ctx, systemPrompt, messages, true)
	if err != nil {
		return nil, fmt.Errorf("candidate research call: %w", err)
	}

	jsonText, err := extractJSON(text)
	if err != nil {
		return nil, fmt.Errorf("extracting candidate JSON: %w", err)
	}

	var result struct {
		Name  string `json:"name"`
		Price string `json:"price"`
	}
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("parsing candidate JSON: %w (raw: %s)", err, text)
	}

	if result.Name == "" {
		return nil, fmt.Errorf("candidate response missing name")
	}

	return &models.Alternative{Name: result.Name, Price: result.Price}, nil
}
