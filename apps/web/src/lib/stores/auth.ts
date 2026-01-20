import { writable, derived, get } from "svelte/store";
import { browser } from "$app/environment";
import { api, ApiError } from "$lib/api/client";

const STORAGE_KEY = "kidsplanet_admin_token";

export type AdminMe = {
    id: number;
    email: string;
    role: string;
};

type AuthState = {
    token: string | null;
    me: AdminMe | null;
    loading: boolean;
};

function createAuthStore() {
    const { subscribe, set, update } = writable<AuthState>({
        token: null,
        me: null,
        loading: false,
    });

    const token = derived({ subscribe }, ($s) => $s.token);
    const me = derived({ subscribe }, ($s) => $s.me);
    const loading = derived({ subscribe }, ($s) => $s.loading);
    const isAuthed = derived({ subscribe }, ($s) => Boolean($s.token));

    function persistToken(t: string | null) {
        if (!browser) return;
        try {
            if (t) localStorage.setItem(STORAGE_KEY, t);
            else localStorage.removeItem(STORAGE_KEY);
        } catch {
        }
    }

    function clear() {
        set({ token: null, me: null, loading: false });
        persistToken(null);
    }

    async function fetchMe(tokenStr?: string) {
        const t = (tokenStr ?? get({ subscribe }).token)?.trim();
        if (!t) return null;

        update((s) => ({ ...s, loading: true }));
        try {
            const data = await api.get<AdminMe>("/admin/me", { token: t });
            update((s) => ({ ...s, me: data, loading: false }));
            return data;
        } catch (e) {
            if (e instanceof ApiError && (e.status === 401 || e.status === 403)) {
                clear();
                return null;
            }
            update((s) => ({ ...s, loading: false }));
            throw e;
        }
    }

    async function setToken(tokenStr: string, opts?: { fetchMe?: boolean }) {
        const t = tokenStr?.trim() || null;
        if (!t) {
            clear();
            return;
        }

        set({ token: t, me: null, loading: false });
        persistToken(t);

        const doFetch = opts?.fetchMe ?? true;
        if (doFetch) await fetchMe(t);
    }

    async function loadFromStorage(opts?: { fetchMe?: boolean }) {
        if (!browser) return;

        let stored: string | null = null;
        try {
            stored = localStorage.getItem(STORAGE_KEY);
        } catch {
            stored = null;
        }

        stored = stored?.trim() || null;
        if (!stored) {
            clear();
            return;
        }

        set({ token: stored, me: null, loading: false });

        const doFetch = opts?.fetchMe ?? true;
        if (doFetch) await fetchMe(stored);
    }

    function getSnapshot() {
        return get({ subscribe });
    }

    return {
        subscribe,
        token,
        me,
        loading,
        isAuthed,
        setToken,
        clear,
        loadFromStorage,
        fetchMe,
        getSnapshot,
    };
}

export const auth = createAuthStore();
