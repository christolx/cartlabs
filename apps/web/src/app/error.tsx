"use client";

export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="system-page shell" role="alert">
      <section className="state-panel state-panel-danger">
        <span className="state-index" aria-hidden="true">
          !
        </span>
        <p className="eyebrow">SYSTEM / ERROR</p>
        <h1>Page unavailable</h1>
        <p>Request could not complete. Check service connection, then retry.</p>
        <button className="button button-primary" type="button" onClick={reset}>
          Try again
        </button>
      </section>
    </main>
  );
}
