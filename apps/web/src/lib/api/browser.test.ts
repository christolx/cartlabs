import { afterEach, describe, expect, it, vi } from "vitest";
import { BrowserAPIError, browserRequest, errorMessage } from "./browser";

describe("browserRequest", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("sends access token and parses successful JSON", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);
    await expect(
      browserRequest<{ ok: boolean }>("/me", "token-1"),
    ).resolves.toEqual({ ok: true });
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(new Headers(init.headers).get("Authorization")).toBe(
      "Bearer token-1",
    );
  });

  it("preserves problem detail and request id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            type: "about:blank",
            title: "Conflict",
            status: 409,
            detail: "stock changed",
            requestId: "req-7",
          }),
          {
            status: 409,
            headers: { "content-type": "application/problem+json" },
          },
        ),
      ),
    );
    const error = await browserRequest("/cart").catch((cause) => cause);
    expect(error).toBeInstanceOf(BrowserAPIError);
    expect(error).toMatchObject({
      status: 409,
      message: "stock changed",
      requestId: "req-7",
    });
  });

  it("maps network failure to service unavailable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    await expect(browserRequest("/me")).rejects.toMatchObject({ status: 503 });
    expect(errorMessage("bad", "fallback")).toBe("fallback");
  });
});
