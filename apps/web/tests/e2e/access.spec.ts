import { expect, test, type Page, type Route } from "@playwright/test";

const users = {
  buyer: {
    id: "11111111-1111-4111-8111-111111111111",
    email: "buyer@demo.cartlabs.local",
    displayName: "Demo Buyer",
    role: "buyer",
  },
  seller: {
    id: "22222222-2222-4222-8222-222222222222",
    email: "seller@demo.cartlabs.local",
    displayName: "Demo Seller",
    role: "seller",
  },
  admin: {
    id: "33333333-3333-4333-8333-333333333333",
    email: "admin@demo.cartlabs.local",
    displayName: "Demo Admin",
    role: "admin",
  },
} as const;

function session(role: keyof typeof users) {
  return {
    accessToken: `token-${role}`,
    tokenType: "Bearer",
    expiresIn: 900,
    user: users[role],
  };
}

async function json(route: Route, body: unknown, status = 200) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

async function mockAnonymous(page: Page) {
  await page.route("**/api/backend/**", async (route) =>
    json(route, { title: "Unauthorized", status: 401 }, 401),
  );
}

test("protected direct URL returns to login with intended destination", async ({
  page,
}) => {
  await mockAnonymous(page);
  await page.goto("/admin/users");
  await expect(page).toHaveURL(/\/login\?next=%2Fadmin%2Fusers/);
  await expect(
    page.getByRole("heading", { name: "Sign in to continue." }),
  ).toBeVisible();
});

test("wrong role direct URL renders forbidden without protected request", async ({
  page,
}) => {
  let adminDataRequested = false;
  await page.route("**/api/backend/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/auth/refresh"))
      return json(route, session("buyer"));
    if (url.pathname.endsWith("/me")) return json(route, users.buyer);
    if (url.pathname.includes("/admin/")) adminDataRequested = true;
    return json(route, { title: "Forbidden", status: 403 }, 403);
  });
  await page.goto("/admin");
  await expect(
    page.getByRole("heading", { name: "Access restricted" }),
  ).toBeVisible();
  expect(adminDataRequested).toBe(false);
});

test("demo quick login enters normal seller page", async ({ page }) => {
  await page.route("**/api/backend/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/auth/refresh"))
      return json(route, { title: "Unauthorized", status: 401 }, 401);
    if (url.pathname.endsWith("/auth/demo-login"))
      return json(route, session("seller"));
    if (url.pathname.endsWith("/me")) return json(route, users.seller);
    if (url.pathname.endsWith("/seller/store"))
      return json(route, { title: "Not found", status: 404 }, 404);
    if (
      url.pathname.endsWith("/seller/products") ||
      url.pathname.endsWith("/seller/orders")
    )
      return json(route, { items: [] });
    return json(route, { title: "Not found", status: 404 }, 404);
  });
  await page.goto("/demo");
  await page.getByRole("button", { name: "Enter seller" }).click();
  await expect(page).toHaveURL(/\/seller$/);
  await expect(
    page.getByRole("heading", { name: "Seller home" }),
  ).toBeVisible();
  await expect(page.getByText("No store exists.")).toBeVisible();
});
