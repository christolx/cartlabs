import Link from "next/link";

export function SiteHeader() {
  return (
    <header className="site-header">
      <div className="shell header-inner">
        <Link className="wordmark" href="/" aria-label="Cartlabs home">
          cartlabs<span>/market</span>
        </Link>
        <nav aria-label="Primary navigation">
          <Link href="/#catalog">Catalog</Link>
          <Link href="/demo">Demo roles</Link>
          <a href="https://github.com/christolx/cartlabs">Source</a>
        </nav>
      </div>
    </header>
  );
}
