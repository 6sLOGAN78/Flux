import { generateOpenApi } from "@ts-rest/open-api";
import { apiContract } from "./contracts";
import * as fs from "node:fs";
import * as path from "node:path";

const openApiDocument = generateOpenApi(apiContract, {
  info: {
    title: "Flux Platform REST API",
    version: "1.0.0",
    description: "Production-grade URL shortening, dynamic link routing, and real-time analytics API.",
  },
});

const targetPath = path.resolve(__dirname, "../../../apps/backend/static/openapi.json");
const targetDir = path.dirname(targetPath);

if (!fs.existsSync(targetDir)) {
  fs.mkdirSync(targetDir, { recursive: true });
}

fs.writeFileSync(targetPath, JSON.stringify(openApiDocument, null, 2), "utf-8");
console.log(`Successfully generated OpenAPI spec at: ${targetPath}`);
