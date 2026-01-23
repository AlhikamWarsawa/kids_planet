import { writable, derived, get } from "svelte/store";
import { browser } from "$app/environment";
import { api, ApiError } from "$lib/api/client";

const STORAGE_KEY = "kidsplanet_play_token";

export type StartSessionResponse = {
    play_token: string;
    expires_at?: string;
};

type SessionState = {
    playToken: string | null;
    expiresAt: number | null;
    loading: boolean;
};

function createSessionStore() {
    const { subscribe, set, update } = writable<SessionState>({
        playToken: null,
        expiresAt: null,
        loading: false,
    });

    const playToken = derived({ subscribe }, ($s) => $s.playToken);
    const expiresAt = derived({ subscribe }, ($s) => $s.expiresAt);
    const loading = derived({ subscribe }, ($s) => $s.loading);
    const isReady = derived(
        { subscribe },
        ($s) => Boolean($s.playToken) && Boolean($s.expiresAt) && Date.now() < ($s.expiresAt ?? 0)
    );

    function persist(token: string | null, expMs: number | null) {
        if (!browser) return;
        try {
            if (!token || !expMs) {
                localStorage.removeItem(STORAGE_KEY);
                return;
            }
            localStorage.setItem(
                STORAGE_KEY,
                JSON.stringify({ playToken: token, expiresAt: expMs })
            );
        } catch {
        }
    }

    function clearSession() {
        set({ playToken: null, expiresAt: null, loading: false });
        persist(null, null);
    }

    function loadFromStorage() {
        if (!browser) return;

        try {
            const raw = localStorage.getItem(STORAGE_KEY);
            if (!raw) return;

            const parsed = JSON.parse(raw) as { playToken?: string; expiresAt?: number };
            const token = (parsed.playToken ?? "").trim() || null;
            const expMs = typeof parsed.expiresAt === "number" ? parsed.expiresAt : null;

            if (!token || !expMs) {
                clearSession();
                return;
            }

            if (Date.now() >= expMs) {
                clearSession();
                return;
            }

            set({ playToken: token, expiresAt: expMs, loading: false });
        } catch {
            clearSession();
        }
    }

    async function startSession(gameId: number) {
        const gid = Number(gameId);
        if (!Number.isFinite(gid) || gid <= 0) {
            throw new ApiError(400, "BAD_REQUEST", "game_id must be a positive number");
        }

        update((s) => ({ ...s, loading: true }));
        try {
            const data = await api.post<StartSessionResponse>("/sessions/start", {
                game_id: gid,
            });

            const token = (data.play_token ?? "").trim();
            if (!token) {
                throw new ApiError(500, "INTERNAL_SERVER_ERROR", "missing play_token");
            }

            let expMs: number;
            if (data.expires_at) {
                const t = Date.parse(data.expires_at);
                expMs = Number.isFinite(t) ? t : Date.now() + 2 * 60 * 60 * 1000;
            } else {
                expMs = Date.now() + 2 * 60 * 60 * 1000;
            }

            set({ playToken: token, expiresAt: expMs, loading: false });
            persist(token, expMs);

            return { playToken: token, expiresAt: expMs };
        } catch (e) {
            update((s) => ({ ...s, loading: false }));
            if (e instanceof ApiError && e.status === 401) {
                clearSession();
            }
            throw e;
        }
    }

    function getSnapshot() {
        return get({ subscribe });
    }

    return {
        subscribe,
        playToken,
        expiresAt,
        loading,
        isReady,
        startSession,
        clearSession,
        loadFromStorage,
        getSnapshot,
    };
}

export const session = createSessionStore();
