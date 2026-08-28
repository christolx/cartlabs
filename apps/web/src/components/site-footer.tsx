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
          <Link href="/#catalog">Browse</Link>
          <Link href="/about">About</Link>
          <Link href="/docs">Docs</Link>
        </nav>
        <button className="footer-region" type="button">
          INDONESIA (IDR)
        </button>
        <nav className="footer-social" aria-label="Project links">
          <a
            href="https://github.com/christolx/cartlabs"
            target="_blank"
            rel="noopener noreferrer"
          >
            Repository
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
