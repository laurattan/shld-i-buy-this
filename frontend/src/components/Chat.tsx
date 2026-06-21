import { useCallback, useEffect, useLayoutEffect, useRef } from "react";
import type { Evaluation, Profile, Session } from "../types";
import DecisionDisplay from "./DecisionDisplay";
import ErrorReceipt from "./ErrorReceipt";
import LoadingReceipt from "./LoadingReceipt";

interface ChatProps {
  evaluations: Evaluation[];
  loading: boolean;
  pendingProduct: string | null;
  error: string | null;
  errorCode: string | null;
  retryProduct: string | null;
  errorProduct: string;
  onRetry: (product: string) => void;
  activeSession: Session | null;
  onBackToFeed: () => void;
  profile: Profile | null;
  hasMore: boolean;
  feedLoading: boolean;
  onLoadMore: () => void;
}

export default function Chat({
  evaluations,
  loading,
  pendingProduct,
  error,
  errorCode,
  retryProduct,
  errorProduct,
  onRetry,
  activeSession,
  onBackToFeed,
  profile,
  hasMore,
  feedLoading,
  onLoadMore,
}: ChatProps) {
  const feedRef = useRef<HTMLDivElement>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const initialScrollDone = useRef(false);
  const scrollHeightBeforeLoad = useRef(0);

  // scroll to bottom instantly on initial load so user lands at most recent receipt
  useEffect(() => {
    if (initialScrollDone.current || activeSession) return;
    if (evaluations.length > 0 && feedRef.current) {
      feedRef.current.scrollTop = feedRef.current.scrollHeight;
      initialScrollDone.current = true;
    }
  }, [evaluations, activeSession]);

  // scroll to bottom smoothly when a submission starts or new eval lands
  useEffect(() => {
    if (activeSession || !feedRef.current) return;
    const newest = evaluations[evaluations.length - 1];
    if (loading || newest?.isNew) {
      feedRef.current.scrollTop = feedRef.current.scrollHeight;
    }
  }, [evaluations, loading, activeSession]);

  // after prepending older evals, restore scroll position so view doesn't jump
  useLayoutEffect(() => {
    if (scrollHeightBeforeLoad.current > 0 && feedRef.current) {
      feedRef.current.scrollTop =
        feedRef.current.scrollHeight - scrollHeightBeforeLoad.current;
      scrollHeightBeforeLoad.current = 0;
    }
  }, [evaluations]);

  // sentinel at top: fires when user scrolls up near the beginning of loaded content
  const setupObserver = useCallback(() => {
    const sentinel = sentinelRef.current;
    const feed = feedRef.current;
    if (!sentinel || !feed) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !feedLoading && hasMore) {
          scrollHeightBeforeLoad.current = feed.scrollHeight;
          onLoadMore();
        }
      },
      { root: feed, rootMargin: "100px" }
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [feedLoading, hasMore, onLoadMore]);

  useEffect(setupObserver, [setupObserver]);

  return (
    <div className="feed" ref={feedRef}>
      <div className="feed-inner">
        {activeSession ? (
          <>
            <button type="button" className="back-to-feed" onClick={onBackToFeed}>
              ← back to feed
            </button>
            <DecisionDisplay
              product={activeSession.product}
              product_name={activeSession.product_name || activeSession.product}
              verdict={activeSession.verdict}
              reasoning={activeSession.reasoning}
              alternatives={activeSession.alternatives}
              created_at={new Date(activeSession.created_at).getTime()}
              receiptNum={activeSession.id}
              userName={profile?.name ?? ""}
              fresh={false}
            />
          </>
        ) : (
          <>
            {/* sentinel at top — triggers loading older sessions on scroll up */}
            {hasMore && (
              <div ref={sentinelRef} className="feed-load-more">
                {feedLoading && (
                  <p className="feed-load-more-label font-accent">loading more…</p>
                )}
              </div>
            )}

            {evaluations.length === 0 && !loading ? (
              <div className="feed-empty">
                <p className="feed-empty-title">no receipts yet</p>
                <p className="feed-empty-sub font-accent">
                  tell us what you&apos;re eyeing and we&apos;ll print a verdict.
                </p>
              </div>
            ) : (
              evaluations.map((ev, i) => (
                <DecisionDisplay
                  key={ev.id ?? ev.created_at ?? i}
                  {...ev}
                  receiptNum={ev.id}
                  userName={profile?.name ?? ""}
                  fresh={!!ev.isNew}
                />
              ))
            )}

            {loading && <LoadingReceipt product={pendingProduct ?? undefined} />}
            {error && (() => {
              const isCooldown = errorCode === "cooldown";
              const isRetryable = !!retryProduct && !isCooldown && errorCode !== "insufficient_credits" && errorCode !== "invalid_request";
              const product = errorProduct || retryProduct || "";
              return (
                <ErrorReceipt
                  product={product}
                  message={error}
                  isCooldown={isCooldown}
                  isRetryable={isRetryable}
                  onRetry={isRetryable ? () => onRetry(retryProduct!) : undefined}
                />
              );
            })()}
          </>
        )}
      </div>
    </div>
  );
}
