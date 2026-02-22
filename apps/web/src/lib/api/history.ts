import { api } from "$lib/api/client";
import { getToken } from "$lib/auth/playerAuth";

export type PlayerHistoryItem = {
    game_id: number;
    title: string;
    played_at: string;
    score?: number;
    status?: string;
};

export type PlayerHistoryPagination = {
    page: number;
    limit: number;
    total: number;
};

export type PlayerHistoryResponse = {
    data: PlayerHistoryItem[];
    pagination: PlayerHistoryPagination;
};

export type PlayerHistoryParams = {
    page?: number;
    limit?: number;
};

function buildQuery(params: PlayerHistoryParams = {}): string {
    const q = new URLSearchParams();

    if (typeof params.page === "number") {
        q.set("page", String(params.page));
    }
    if (typeof params.limit === "number") {
        q.set("limit", String(params.limit));
    }

    const qs = q.toString();
    return qs ? `?${qs}` : "";
}

export function getPlayerHistory(params: PlayerHistoryParams = {}): Promise<PlayerHistoryResponse> {
    const qs = buildQuery(params);
    const token = getToken();

    return api.get<PlayerHistoryResponse>(`/player/history${qs}`, {
        token: token ?? undefined,
    });
}
