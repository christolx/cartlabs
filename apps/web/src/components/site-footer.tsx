import Link from "next/link";

export function SiteFooter() {
  return <footer className="site-footer"><div className="shell footer-inner"><p>Cartlabs marketplace demo. Inventory and orders use live API state.</p><Link href="/demo">Demo guide</Link></div></footer>;
}
