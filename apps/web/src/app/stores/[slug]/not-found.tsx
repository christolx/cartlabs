import Link from "next/link";

export default function StoreNotFound() {
  return (
    <main className="system-page shell">
      <section className="state-panel">
        <span className="state-index" aria-hidden="true">
          404
        </span>
        <p className="eyebrow">STORE / NOT FOUND</p>
        <h1>Store not found</h1>
        <p>Store may be unavailable or not verified.</p>
        <Link className="button button-primary" href="/">
          Browse catalog
        </Link>
      </section>
    </main>
  );
}
