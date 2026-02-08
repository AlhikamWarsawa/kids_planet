import { api } from "$lib/api/client";

export type DashboardTopGameDTO = {
    game_id: number;
    title: string;
    plays: number;
};

export type DashboardOverviewDTO = {
    sessions_today: number;
    top_games: DashboardTopGameDTO[];
    total_active_games: number;
    total_players: number;
};

export function adminDashboardOverview(): Promise<DashboardOverviewDTO> {
    return api.get<DashboardOverviewDTO>("/admin/dashboard/overview");
}
