// Development forwards API requests to the gateway; production serves pages.
export function gatewayProxy(env = process.env) {
  const target = new URL(env.GATEWAY_URL || "http://127.0.0.1:8082");
  if (
    !["http:", "https:"].includes(target.protocol) ||
    target.username ||
    target.password ||
    target.pathname !== "/" ||
    target.search ||
    target.hash
  )
    throw new Error("Invalid GATEWAY_URL");
  const host = env.GATEWAY_HOST || "admin.sitehub.local";
  if (!/^[a-zA-Z0-9.-]+(?::[0-9]+)?$/.test(host))
    throw new Error("Invalid GATEWAY_HOST");
  return {
    "/v1": {
      target: target.origin,
      changeOrigin: true,
      headers: { host },
      timeout: 600000,
      proxyTimeout: 600000,
    },
  };
}
