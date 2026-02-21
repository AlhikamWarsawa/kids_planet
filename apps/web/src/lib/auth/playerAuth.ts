import { browser } from "$app/environment";
import { createApiClient, ApiError } from "$lib/api/client";

const STORAGE_KEY = "player_token";
const api = createApiClient();

export type PlayerIdentity = {
    id: string;
    email: string;
};

export type PlayerAuthResult = {
    token: string;
    player: PlayerIdentity;
};

function normalizeEmail(email: string): string {
    return email.trim().toLowerCase();
}

function persistToken(token: string | null) {
    if (!browser) return;
    try {
        if (token) localStorage.setItem(STORAGE_KEY, token);
        else localStorage.removeItem(STORAGE_KEY);
    } catch {
    }
}

export function getToken(): string | null {
    if (!browser) return null;
    try {
        const token = localStorage.getItem(STORAGE_KEY);
        return token?.trim() || null;
    } catch {
        return null;
    }
}

export function isLoggedIn(): boolean {
    return Boolean(getToken());
}

export async function register(email: string, pin: string): Promise<PlayerAuthResult> {
    const data = await api.post<PlayerAuthResult>("/auth/player/register", {
        email: normalizeEmail(email),
        pin,
    });

    if (!data?.token) {
        throw new ApiError(500, "INVALID_RESPONSE", "missing player token");
    }

    persistToken(data.token);
    return data;
}

export async function login(email: string, pin: string): Promise<PlayerAuthResult> {
    const data = await api.post<PlayerAuthResult>("/auth/player/login", {
        email: normalizeEmail(email),
        pin,
    });

    if (!data?.token) {
        throw new ApiError(500, "INVALID_RESPONSE", "missing player token");
    }

    persistToken(data.token);
    return data;
}

export async function logout(): Promise<void> {
    try {
        await api.post<void>("/auth/player/logout");
    } finally {
        persistToken(null);
    }
}
