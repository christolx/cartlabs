import Link from "next/link";
export default function ProductNotFound() {
  return (
    <main className="center-state shell">
      <h1>Product not found</h1>
      <p>Product may be unpublished or removed.</p>
      <Link className="button button-primary" href="/#catalog">
        Browse catalog
      </Link>
    </main>
  );
}
