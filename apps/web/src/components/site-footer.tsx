import Link from "next/link";

export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="footer-main">
        <div className="footer-brand">
          <strong>Cartlabs</strong>
          <p>
            Independent stores.
            <br />
            Verified sellers.
            <br />
            Live inventory.
          </p>
        </div>
        <nav aria-label="Browse links">
          <Link href="/#catalog">Browse all</Link>
          <Link href="/#catalog">All categories</Link>
          <Link href="/#catalog">Stores</Link>
        </nav>
        <nav aria-label="Company links">
          <Link href="/demo">About</Link>
          <Link href="/demo">Help</Link>
          <Link href="/demo">Contact</Link>
        </nav>
        <button className="footer-region" type="button">
          INDONESIA (IDR)
        </button>
        <nav className="footer-social" aria-label="Social links">
          <a
            href="https://christofle.dev"
            target="_blank"
            rel="noopener noreferrer"
          >
            christofle.dev
          </a>
        </nav>
      </div>
      <div className="footer-legal">
        <span>© 2026 Cartlabs</span>
        <span>All rights reserved.</span>
        <i aria-hidden="true" />
      </div>
    </footer>
  );
}
