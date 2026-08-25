import Link from "next/link";

export default function StoreNotFound() {
  return (
    <main className="center-state shell">
      <h1>Store not found</h1>
      <p>Store may be unavailable or not verified.</p>
      <Link className="button button-primary" href="/">
        Browse catalog
      </Link>
    </main>
  );
}
