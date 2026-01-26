import { api } from "$lib/api/client";

export type LeaderboardItem = {
    member: string;
    score: number;
};

export type LeaderboardViewResponse = {
    game_id: number;
    period: string;
    scope: string;
    limit: number;
    items: LeaderboardItem[];
};

export type GetLeaderboardOpts = {
    period?: "daily" | "weekly";
    scope?: "game" | "global";
    limit?: number;
};

export async function getLeaderboard(gameId: number, opts: GetLeaderboardOpts = {}) {
    const qs = new URLSearchParams();

    if (opts.period) qs.set("period", opts.period);
    if (opts.scope) qs.set("scope", opts.scope);
    if (typeof opts.limit === "number") qs.set("limit", String(opts.limit));

    const q = qs.toString();
    const path = q ? `/leaderboard/${gameId}?${q}` : `/leaderboard/${gameId}`;

    return api.get<LeaderboardViewResponse>(path);
}