import http from "k6/http";
import { check, sleep } from "k6";

import { API_BASE, jsonHeaders, startSession } from "./common.js";

const GAME_ID = Number(__ENV.GAME_ID || 1);

export const options = {
  scenarios: {
    analytics_ingest: {
      executor: "constant-vus",
      vus: Number(__ENV.VUS || 10),
      duration: __ENV.DURATION || "2m",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.02"],
    "http_req_duration{endpoint:analytics_event}": ["p(95)<300", "avg<200"],
    checks: ["rate>0.98"],
  },
  summaryTrendStats: ["avg", "p(90)", "p(95)", "p(99)", "min", "max"],
};

export default function () {
  const playToken = startSession(GAME_ID);

  const payload = {
    play_token: playToken,
    name: "k6_event",
    data: {
      vu: __VU,
      iteration: __ITER,
      source: "k6",
    },
  };

  const res = http.post(`${API_BASE}/analytics/event`, JSON.stringify(payload), {
    headers: jsonHeaders(),
    tags: { endpoint: "analytics_event" },
  });

  check(res, {
    "analytics status is 200": (r) => r.status === 200,
    "analytics response ok": (r) => r.json("data.ok") === true,
  });

  sleep(1);
}
