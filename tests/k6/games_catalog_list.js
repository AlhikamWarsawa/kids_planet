import http from "k6/http";
import { check, sleep } from "k6";

import { API_BASE } from "./common.js";

export const options = {
  scenarios: {
    catalog: {
      executor: "constant-vus",
      vus: Number(__ENV.VUS || 25),
      duration: __ENV.DURATION || "2m",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:games_catalog}": ["p(95)<300", "avg<180"],
    checks: ["rate>0.99"],
  },
  summaryTrendStats: ["avg", "p(90)", "p(95)", "p(99)", "min", "max"],
};

export default function () {
  const res = http.get(`${API_BASE}/games?page=1&limit=24&sort=newest`, {
    headers: { Accept: "application/json" },
    tags: { endpoint: "games_catalog" },
  });

  check(res, {
    "games status is 200": (r) => r.status === 200,
    "games response has array": (r) => Array.isArray(r.json("data.items")),
    "games pagination has total": (r) => typeof r.json("data.total") === "number",
  });

  sleep(0.5);
}
