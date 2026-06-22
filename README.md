# shld i buy this

A full-stack pre-purchase research tool designed to help users make more deliberate purchasing decisions. Submit a product URL or description, and a Claude-powered agent researches the product and returns a personalized verdict: **buy**, **wait**, **find an alternative**, or **don't buy**.

Results are displayed as digital shopping receipts.

**[Live Demo](https://shld-i-buy-this.vercel.app)**

## Tech Stack

- **Frontend:** React, TypeScript, CSS
- **Backend:** Go, Claude API (web search + fetch)
- **Data:** PostgreSQL, Redis
- **Deployment:** Vercel, Railway

## How It Works

1. Users submit a product URL or description and provide their purchasing habits for personalized recommendations.
2. The Go backend checks Redis for rate limits, purchase cooldowns, and cached results.
3. Claude researches the product using web search and fetch, returning a structured verdict and reasoning.
4. For `find_alternative` verdicts, the agent generates five candidate products and verifies them concurrently using Go goroutines, returning up to three alternatives.
5. Results are cached in Redis (24-hour TTL), and new evaluations are persisted to PostgreSQL.
6. The frontend displays decisions as receipt-style cards, with previous evaluations accessible through session history.

## Architecture

```text
React Frontend (Vercel)
          |
          v
Go HTTP Server (Railway)
    |
    +-- Redis
    |     - Rate limiting
    |     - Purchase cooldowns
    |     - Response caching
    |
    +-- Claude Research Pipeline
    |     - Product research + verdict
    |     - Alternative discovery (if needed)
    |     - Concurrent alternative verification
    |
    +-- PostgreSQL
          - User profile
          - Session history
```

## Local Setup

**Prerequisites:** Go 1.26.4, Node.js compatible with Vite 8, Docker, Anthropic API key.

Clone the repository:

```bash
git clone https://github.com/laurattan/shld-i-buy-this.git
cd shld-i-buy-this
```

Start PostgreSQL and Redis:

```bash
docker run -d --name shld-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=shld_i_buy_this \
  -p 5432:5432 postgres

docker run -d --name shld-redis \
  -p 6379:6379 redis
```

Create a `.env` file in the repository root:

```dotenv
ANTHROPIC_API_KEY=your_api_key
DATABASE_URL=postgres://postgres:postgres@localhost:5432/shld_i_buy_this?sslmode=disable
REDIS_URL=redis://localhost:6379
PORT=8080
ALLOWED_ORIGIN=http://localhost:5173
```

Start the backend:

```bash
go run main.go --server
```

Create `frontend/.env`:

```dotenv
VITE_API_URL=http://localhost:8080
```

Start the frontend:

```bash
cd frontend
npm install
npm run dev
```
