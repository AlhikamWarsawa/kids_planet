import http from "k6/http";
import { check, sleep } from "k6";

import { API_BASE } from "./common.js";

const GAME_ID = Number(__ENV.GAME_ID || 1);

export const options = {
  scenarios: {
    leaderboard_read: {
      executor: "constant-vus",
      vus: Number(__ENV.VUS || 40),
      duration: __ENV.DURATION || "2m",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:leaderboard_read}": ["p(95)<150", "avg<100"],
    checks: ["rate>0.99"],
  },
  summaryTrendStats: ["avg", "p(90)", "p(95)", "p(99)", "min", "max"],
};

export default function () {
  const res = http.get(
    `${API_BASE}/leaderboard/${GAME_ID}?period=daily&scope=game&limit=10`,
    {
      headers: { Accept: "application/json" },
      tags: { endpoint: "leaderboard_read" },
    },
  );

  check(res, {
    "leaderboard status is 200": (r) => r.status === 200,
    "leaderboard items is array": (r) => Array.isArray(r.json("data.items")),
  });

  sleep(0.3);
}
