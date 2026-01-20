import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import { auth } from "$lib/stores/auth";
import { api, ApiError } from "$lib/api/client";
import type { AdminMe } from "$lib/stores/auth";

export async function fetchAdminMe(opts?: { redirectOnFail?: boolean }) {
    const redirectOnFail = opts?.redirectOnFail ?? false;

    const snap = auth.getSnapshot();
    const token = snap.token?.trim();
    if (!token) {
        if (redirectOnFail && browser) await goto("/admin/login");
        return null;
    }

    try {
        const me = await api.get<AdminMe>("/admin/me", { token });
        await auth.fetchMe(token);
        return me;
    } catch (e) {
        if (e instanceof ApiError && (e.status === 401 || e.status === 403)) {
            auth.clear();
            if (redirectOnFail && browser) await goto("/admin/login");
            return null;
        }
        throw e;
    }
}

export async function requireAdmin() {
    await auth.loadFromStorage({ fetchMe: false });

    const snap = auth.getSnapshot();
    const token = snap.token?.trim();
    if (!token) {
        if (browser) await goto("/admin/login");
        return false;
    }

    const me = await fetchAdminMe({ redirectOnFail: true });
    return Boolean(me);
}
