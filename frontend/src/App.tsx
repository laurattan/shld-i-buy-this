import { useCallback, useEffect, useRef, useState } from "react";
import Chat from "./components/Chat";
import CustomerCard from "./components/CustomerCard";
import Header from "./components/Header";
import InputBox from "./components/InputBox";
import ProfileModal from "./components/ProfileModal";
import Sidebar from "./components/Sidebar";
import { evaluate, getProfile, getSessions } from "./api";
import type { Evaluation, Profile, Session } from "./types";
import "./App.css";

const PAGE_SIZE = 10;

function sessionToEval(s: Session): Evaluation {
  return {
    id: s.id,
    product: s.product,
    product_name: s.product_name || s.product,
    verdict: s.verdict,
    reasoning: s.reasoning,
    alternatives: s.alternatives,
    created_at: new Date(s.created_at).getTime(),
  };
}

function App() {
  const [evaluations, setEvaluations] = useState<Evaluation[]>([]);
  const [loading, setLoading] = useState(false);
  const [pendingProduct, setPendingProduct] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [errorCode, setErrorCode] = useState<string | null>(null);
  const [retryProduct, setRetryProduct] = useState<string | null>(null);
  const lastProduct = useRef<string>("");
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [profileModalOpen, setProfileModalOpen] = useState(false);
  const [activeSessionId, setActiveSessionId] = useState<number | null>(null);

  // feed pagination — DESC from DB, reversed for display (oldest-first)
  const [feedOffset, setFeedOffset] = useState(PAGE_SIZE);
  const [hasMore, setHasMore] = useState(false);
  const [feedLoading, setFeedLoading] = useState(false);

  const activeSession = activeSessionId !== null
    ? (sessions.find((s) => s.id === activeSessionId) ?? null)
    : null;

  useEffect(() => {
    getSessions().then(setSessions);

    getSessions({ limit: PAGE_SIZE, offset: 0 }).then((first) => {
      // DESC from DB → reverse so oldest of the batch is at top, newest at bottom
      setEvaluations([...first].reverse().map(sessionToEval));
      setHasMore(first.length === PAGE_SIZE);
      setFeedOffset(PAGE_SIZE);
    });

    getProfile().then((p) => {
      setProfile(p);
      if (!p?.name) setProfileModalOpen(true);
    });
  }, []);

  const loadMore = useCallback(async () => {
    if (feedLoading || !hasMore) return;
    setFeedLoading(true);
    try {
      // next DESC page → reverse → prepend (they're older)
      const page = await getSessions({ limit: PAGE_SIZE, offset: feedOffset });
      const older = [...page].reverse().map(sessionToEval);
      setEvaluations((prev) => [...older, ...prev]);
      setHasMore(page.length === PAGE_SIZE);
      setFeedOffset((prev) => prev + PAGE_SIZE);
    } finally {
      setFeedLoading(false);
    }
  }, [feedLoading, hasMore, feedOffset]);

  const handleProfileSaved = (p: Profile) => {
    setProfile(p);
    setProfileModalOpen(false);
  };

  const handleSubmit = async (product: string) => {
    lastProduct.current = product;
    setActiveSessionId(null);
    setLoading(true);
    setPendingProduct(product);
    setError(null);
    setErrorCode(null);
    setRetryProduct(null);
    try {
      const { product_name, verdict, reasoning, alternatives } = await evaluate(product);
      const newEval: Evaluation = {
        product,
        product_name,
        verdict,
        reasoning,
        alternatives,
        created_at: Date.now(),
        isNew: true,
      };
      setEvaluations((prev) => [...prev, newEval]);
      // bump offset so next DESC page skips the newly written session
      setFeedOffset((prev) => prev + 1);
      getSessions().then((fetched) => {
        setSessions(fetched);
        // backfill the id onto the optimistic eval so receipt number shows immediately
        const match = fetched.find((s) => s.product === product && s.verdict === verdict);
        if (match) {
          setEvaluations((prev) =>
            prev.map((e) => (e.isNew && e.product === product ? { ...e, id: match.id } : e))
          );
        }
      });
    } catch (e) {
      if (e instanceof Error) {
        const typed = e as Error & { verdict?: string; code?: string; status?: number };
        if (typed.verdict === "do_not_buy") {
          setError("you already decided not to buy this — don't.");
          setErrorCode("cooldown");
        } else if (typed.verdict === "wait") {
          setError("you decided to wait on this — hold off a bit longer.");
          setErrorCode("cooldown");
        } else {
          setError(e.message);
          const code = typed.status === 429 ? "cooldown" : (typed.code ?? null);
          setErrorCode(code);
        }
      } else {
        setError("Something went wrong — try again.");
        setErrorCode("unknown");
      }
      setRetryProduct(product);
    } finally {
      setLoading(false);
      setPendingProduct(null);
    }
  };

  const handleSessionClick = (session: Session) => {
    setActiveSessionId(session.id);
    setSidebarOpen(false);
  };

  const hasProfile = !!profile?.name;

  return (
    <div className="app">
      <Header
        onToggleSidebar={() => setSidebarOpen((v) => !v)}
        historyOpen={sidebarOpen}
      />

      <Sidebar
        open={sidebarOpen}
        sessions={sessions}
        activeId={activeSessionId}
        onSelect={handleSessionClick}
        onClose={() => setSidebarOpen(false)}
      />

      <div className="app-body">
        <CustomerCard
          profile={profile}
          onEditProfile={() => setProfileModalOpen(true)}
        />

        <div className="right-col">
          <Chat
            evaluations={evaluations}
            loading={loading}
            pendingProduct={pendingProduct}
            error={error}
            errorCode={errorCode}
            retryProduct={retryProduct}
            errorProduct={lastProduct.current}
            onRetry={handleSubmit}
            activeSession={activeSession}
            onBackToFeed={() => setActiveSessionId(null)}
            profile={profile}
            hasMore={hasMore}
            feedLoading={feedLoading}
            onLoadMore={loadMore}
          />
        </div>
      </div>

      <hr className="perf" />
      <InputBox
        onSubmit={handleSubmit}
        loading={loading}
        needsProfile={!hasProfile}
        onRequestProfile={() => setProfileModalOpen(true)}
      />

      {(!profile || profileModalOpen) && (
        <ProfileModal
          existing={profile}
          onSave={handleProfileSaved}
          dismissable={hasProfile}
          onClose={() => setProfileModalOpen(false)}
        />
      )}
    </div>
  );
}

export default App;
