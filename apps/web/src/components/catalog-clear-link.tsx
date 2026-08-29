"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import {
  catalogDepthKey,
  catalogDepthResetEvent,
} from "@/components/catalog-depth";

export function CatalogClearLink({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  function requestCatalogReset() {
    window.sessionStorage.removeItem(catalogDepthKey);
    window.dispatchEvent(new Event(catalogDepthResetEvent));
  }

  return (
    <Link
      className={className}
      href="/"
      scroll={false}
      onClick={requestCatalogReset}
    >
      {children}
    </Link>
  );
}
