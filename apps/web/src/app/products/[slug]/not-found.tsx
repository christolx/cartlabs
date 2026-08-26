import Link from "next/link";
export default function ProductNotFound() {
  return (
    <main className="system-page shell">
      <section className="state-panel">
        <span className="state-index" aria-hidden="true">
          404
        </span>
        <p className="eyebrow">PRODUCT / NOT FOUND</p>
        <h1>Product not found</h1>
        <p>Product may be unpublished or removed.</p>
        <Link className="button button-primary" href="/#catalog">
          Browse catalog
        </Link>
      </section>
    </main>
  );
}
