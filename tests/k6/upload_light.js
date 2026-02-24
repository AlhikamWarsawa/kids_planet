import http from "k6/http";
import { check, sleep } from "k6";

import { API_BASE, authHeaders, getAdminToken } from "./common.js";

const GAME_ID = Number(__ENV.GAME_ID || 1);
const ZIP_FILE = (__ENV.ZIP_FILE || "").trim();
const ZIP_NAME = (__ENV.ZIP_NAME || "k6-light-upload.zip").trim();

if (ZIP_FILE === "") {
  throw new Error("ZIP_FILE env var is required (path to a small valid game zip with root index.html)");
}

const ZIP_BINARY = open(ZIP_FILE, "b");

export const options = {
  scenarios: {
    upload_light: {
      executor: "constant-vus",
      vus: Number(__ENV.VUS || 1),
      duration: __ENV.DURATION || "1m",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.05"],
    "http_req_duration{endpoint:upload_light}": ["p(95)<5000", "avg<3000"],
    checks: ["rate>0.95"],
  },
  summaryTrendStats: ["avg", "p(90)", "p(95)", "p(99)", "min", "max"],
};

export function setup() {
  return { adminToken: getAdminToken() };
}

export default function (data) {
  const form = {
    file: http.file(ZIP_BINARY, ZIP_NAME, "application/zip"),
  };

  const res = http.post(`${API_BASE}/admin/games/${GAME_ID}/upload`, form, {
    headers: authHeaders(data.adminToken),
    tags: { endpoint: "upload_light" },
  });

  check(res, {
    "upload status is 200": (r) => r.status === 200,
    "upload returns object key": (r) => {
      const key = r.json("data.object_key");
      return typeof key === "string" && key.length > 0;
    },
    "upload returns playable game_url": (r) => {
      const gameUrl = r.json("data.game_url");
      return typeof gameUrl === "string" && gameUrl.includes(`/games/${GAME_ID}/current/index.html`);
    },
  });

  sleep(2);
}
