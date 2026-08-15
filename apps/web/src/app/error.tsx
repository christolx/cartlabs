"use client";

export default function GlobalError({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return <main className="center-state shell" role="alert"><h1>Page unavailable</h1><p>Request could not complete. Check service connection, then retry.</p><button className="button button-primary" type="button" onClick={reset}>Try again</button></main>;
}
