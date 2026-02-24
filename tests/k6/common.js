import http from "k6/http";
import { check, fail } from "k6";

const DEFAULT_BASE_URL = "http://localhost";
const DEFAULT_ADMIN_EMAIL = "admin@kidsplanet.com";
const DEFAULT_ADMIN_PASSWORD = "12345678";

export const BASE_URL = (__ENV.BASE_URL || DEFAULT_BASE_URL).replace(/\/+$/, "");
export const API_BASE = `${BASE_URL}/api`;

export function jsonHeaders(extra = {}) {
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    ...extra,
  };
}

export function authHeaders(token, extra = {}) {
  return {
    Accept: "application/json",
    Authorization: `Bearer ${token}`,
    ...extra,
  };
}

export function getAdminToken() {
  const envToken = (__ENV.ADMIN_TOKEN || "").trim();
  if (envToken !== "") {
    return envToken;
  }

  const email = (__ENV.ADMIN_EMAIL || DEFAULT_ADMIN_EMAIL).trim();
  const password = (__ENV.ADMIN_PASSWORD || DEFAULT_ADMIN_PASSWORD).trim();

  const loginRes = http.post(
    `${API_BASE}/auth/admin/login`,
    JSON.stringify({ email, password }),
    { headers: jsonHeaders() },
  );

  const ok = check(loginRes, {
    "admin login status is 200": (r) => r.status === 200,
    "admin login returns access_token": (r) => {
      const token = r.json("data.access_token");
      return typeof token === "string" && token.length > 0;
    },
  });

  if (!ok) {
    fail(
      `admin login failed with status ${loginRes.status}; set ADMIN_TOKEN or valid ADMIN_EMAIL/ADMIN_PASSWORD`,
    );
  }

  return loginRes.json("data.access_token");
}

export function startSession(gameId, playerToken = "") {
  const headers = jsonHeaders();
  const token = (playerToken || "").trim();
  if (token !== "") {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = http.post(
    `${API_BASE}/sessions/start`,
    JSON.stringify({ game_id: gameId }),
    { headers, tags: { endpoint: "sessions_start" } },
  );

  const ok = check(res, {
    "start session status is 200": (r) => r.status === 200,
    "start session returns play_token": (r) => {
      const playToken = r.json("data.play_token");
      return typeof playToken === "string" && playToken.length > 0;
    },
  });

  if (!ok) {
    fail(`start session failed with status ${res.status}`);
  }

  return res.json("data.play_token");
}
