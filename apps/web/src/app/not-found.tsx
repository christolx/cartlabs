import Link from "next/link";

export default function NotFound() {
  return (
    <main className="system-page shell">
      <section className="state-panel">
        <span className="state-index" aria-hidden="true">
          404
        </span>
        <p className="eyebrow">SYSTEM / ROUTE</p>
        <h1>Page not found</h1>
        <p>Route does not exist or resource is unavailable.</p>
        <Link className="button button-primary" href="/">
          Browse catalog
        </Link>
      </section>
    </main>
  );
}
