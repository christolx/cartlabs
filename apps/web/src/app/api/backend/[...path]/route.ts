const backendURL = () =>
  (
    process.env.API_INTERNAL_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080/api/v1"
  ).replace(/\/$/, "");

async function proxy(
  request: Request,
  context: RouteContext<"/api/backend/[...path]">,
) {
  const { path } = await context.params;
  const source = new URL(request.url);
  const target = `${backendURL()}/${path.map(encodeURIComponent).join("/")}${source.search}`;
  const headers = new Headers();
  for (const name of [
    "accept",
    "authorization",
    "content-type",
    "cookie",
    "idempotency-key",
    "x-request-id",
  ]) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }
  const contentLength = Number(request.headers.get("content-length") ?? 0);
  if (contentLength > 1_048_576)
    return Response.json({ detail: "Request body too large" }, { status: 413 });
  const body =
    request.method === "GET" || request.method === "HEAD"
      ? undefined
      : await request.arrayBuffer();
  if (body && body.byteLength > 1_048_576)
    return Response.json({ detail: "Request body too large" }, { status: 413 });
  try {
    const upstream = await fetch(target, {
      method: request.method,
      headers,
      body,
      cache: "no-store",
      signal: AbortSignal.timeout(10_000),
    });
    const responseHeaders = new Headers();
    const contentType = upstream.headers.get("content-type");
    if (contentType) responseHeaders.set("content-type", contentType);
    const setCookie = upstream.headers.get("set-cookie");
    if (setCookie)
      responseHeaders.set(
        "set-cookie",
        setCookie.replace("Path=/api/v1/auth", "Path=/api/backend/auth"),
      );
    for (const name of ["x-request-id", "retry-after"]) {
      const value = upstream.headers.get(name);
      if (value) responseHeaders.set(name, value);
    }
    return new Response(upstream.body, {
      status: upstream.status,
      headers: responseHeaders,
    });
  } catch {
    return Response.json({ detail: "API unavailable" }, { status: 503 });
  }
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
