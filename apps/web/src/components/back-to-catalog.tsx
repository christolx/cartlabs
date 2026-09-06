"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  catalogNavigationKey,
  type CatalogNavigation,
} from "@/components/catalog-navigation";

export function BackToCatalog({
  href,
  productPath,
}: {
  href: string;
  productPath: string;
}) {
  const router = useRouter();

  return (
    <Link
      className="back-link"
      href={href}
      onNavigate={(event) => {
        try {
          const value = window.sessionStorage.getItem(catalogNavigationKey);
          if (!value) return;
          const navigation = JSON.parse(value) as CatalogNavigation;
          if (
            navigation.productPath !== productPath ||
            navigation.returnTo !== href
          )
            return;
          event.preventDefault();
          window.sessionStorage.removeItem(catalogNavigationKey);
          router.back();
        } catch {
          window.sessionStorage.removeItem(catalogNavigationKey);
        }
      }}
    >
      Back to catalog
    </Link>
  );
}
