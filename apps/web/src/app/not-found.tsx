import Link from "next/link";

export default function NotFound() {
  return <main className="center-state shell"><p className="eyebrow">404</p><h1>Page not found</h1><p>Route does not exist or resource is unavailable.</p><Link className="button button-primary" href="/">Browse catalog</Link></main>;
}
