import { fileURLToPath } from "node:url";
import { createFrontendServer } from "./server.mjs";

const port = Number(process.env.PORT || 4311);
if (!Number.isInteger(port) || port < 1 || port > 65535)
  throw new Error("Invalid PORT");
const host = process.env.HOST || "127.0.0.1";
const server = createFrontendServer({
  dist: fileURLToPath(new URL("../dist/", import.meta.url)),
});
server.on("error", (error) => {
  console.error(error.message);
  process.exitCode = 1;
});
server.listen(port, host, () =>
  console.log(`admin listening on ${host}:${port}`),
);
for (const signal of ["SIGINT", "SIGTERM"])
  process.once(signal, () => {
    server.close(() => process.exit(0));
    setTimeout(() => {
      server.closeAllConnections();
      process.exit(0);
    }, 10000).unref();
  });
