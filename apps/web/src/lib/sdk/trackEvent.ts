import { api, ApiError } from "$lib/api/client";
import { session } from "$lib/stores/session";

export async function trackEvent(
    name: string,
    data?: Record<string, any>
): Promise<void> {
    const eventName = (name ?? "").trim();
    if (!eventName) return;

    const { playToken } = session.getSnapshot();
    const token = (playToken ?? "").trim();
    if (!token) return;

    try {
        await api.post("/analytics/event", {
            play_token: token,
            name: eventName,
            data,
        });
    } catch (err) {
        if (import.meta.env?.DEV) {
            const msg =
                err instanceof ApiError
                    ? `${err.code}: ${err.message}`
                    : "Unknown error";
            console.warn("trackEvent failed:", msg);
        }
    }
}
