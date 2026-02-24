import http from "k6/http";
import { check, sleep } from "k6";

import { API_BASE, authHeaders, getAdminToken } from "./common.js";

export const options = {
  scenarios: {
    dashboard: {
      executor: "constant-vus",
      vus: Number(__ENV.VUS || 10),
      duration: __ENV.DURATION || "2m",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:dashboard_overview}": ["p(95)<700", "avg<400"],
    checks: ["rate>0.99"],
  },
  summaryTrendStats: ["avg", "p(90)", "p(95)", "p(99)", "min", "max"],
};

export function setup() {
  return { adminToken: getAdminToken() };
}

export default function (data) {
  const res = http.get(`${API_BASE}/admin/dashboard/overview`, {
    headers: authHeaders(data.adminToken),
    tags: { endpoint: "dashboard_overview" },
  });

  check(res, {
    "dashboard status is 200": (r) => r.status === 200,
    "dashboard has total_active_games": (r) => typeof r.json("data.total_active_games") === "number",
    "dashboard has sessions_today": (r) => typeof r.json("data.sessions_today") === "number",
  });

  sleep(1);
}
