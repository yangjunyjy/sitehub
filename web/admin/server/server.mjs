import http from "node:http";
import { existsSync } from "node:fs";
import { resolve } from "node:path";
import sirv from "sirv";

function json(res, status, message) {
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Cache-Control": "no-store",
  });
  res.end(JSON.stringify({ code: status, message }));
}

export function createFrontendServer({
  dist,
  websiteURL = process.env.WEB_SITE_URL || "",
}) {
  if (!existsSync(resolve(dist, "index.html")))
    throw new Error(`Build admin before starting its server`);
  const configuredWebsiteURL = websiteURL.trim();
  if (configuredWebsiteURL) {
    let parsed;
    try {
      parsed = new URL(configuredWebsiteURL);
    } catch {
      throw new Error(
        "Invalid WEB_SITE_URL: expected an absolute HTTP or HTTPS URL",
      );
    }
    if (
      !/^https?:\/\//i.test(configuredWebsiteURL) ||
      parsed.username ||
      parsed.password
    )
      throw new Error(
        "Invalid WEB_SITE_URL: expected an absolute HTTP or HTTPS URL without credentials",
      );
  }
  const runtimeScript = `window.__SITEHUB_CONFIG__ = ${JSON.stringify({
    websiteURL: configuredWebsiteURL,
  })
    .replace(/</g, "\\u003c")
    .replace(/\u2028/g, "\\u2028")
    .replace(/\u2029/g, "\\u2029")};\n`;
  const assets = sirv(dist, {
    etag: true,
    single: true,
    setHeaders(res, pathname) {
      res.setHeader("X-Content-Type-Options", "nosniff");
      res.setHeader(
        "Cache-Control",
        pathname.includes("/assets/")
          ? "public, max-age=31536000, immutable"
          : "no-cache",
      );
    },
  });
  const server = http.createServer((req, res) => {
    const pathname = req.url.split("?")[0];
    if (
      pathname === "/sitehub-runtime.js" &&
      ["GET", "HEAD"].includes(req.method)
    ) {
      res.writeHead(200, {
        "Content-Type": "application/javascript; charset=utf-8",
        "Cache-Control": "no-store",
        "X-Content-Type-Options": "nosniff",
      });
      return res.end(req.method === "HEAD" ? undefined : runtimeScript);
    }
    if (req.url === "/healthz" && ["GET", "HEAD"].includes(req.method)) {
      res.writeHead(200, {
        "Content-Type": "application/json",
        "Cache-Control": "no-store",
      });
      return res.end(JSON.stringify({ status: "ok", app: "admin" }));
    }
    if (/^\/(?:v1|api|uploads)(?:\/|$)/.test(pathname))
      return json(res, 404, "API traffic must use the gateway");
    if (!["GET", "HEAD"].includes(req.method))
      return json(res, 405, "Method not allowed");
    assets(req, res, () => json(res, 404, "Not found"));
  });
  server.requestTimeout = 600000;
  return server;
}
