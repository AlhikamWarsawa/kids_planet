import { api } from '$lib/api/client';
import type { GameDetail, GameListResponse } from '$lib/types/game';

export type GamesSort = 'newest' | 'popular';

export type ListGamesParams = {
    age_category_id?: number;
    education_category_id?: number;
    sort?: GamesSort;
    page?: number;
    limit?: number;
};

function buildQuery(params: ListGamesParams = {}): string {
    const q = new URLSearchParams();

    if (typeof params.age_category_id === 'number') {
        q.set('age_category_id', String(params.age_category_id));
    }
    if (typeof params.education_category_id === 'number') {
        q.set('education_category_id', String(params.education_category_id));
    }
    if (params.sort === 'newest' || params.sort === 'popular') {
        q.set('sort', params.sort);
    }
    if (typeof params.page === 'number') {
        q.set('page', String(params.page));
    }
    if (typeof params.limit === 'number') {
        q.set('limit', String(params.limit));
    }

    const qs = q.toString();
    return qs ? `?${qs}` : '';
}

export function listGames(params: ListGamesParams = {}): Promise<GameListResponse> {
    const qs = buildQuery(params);
    return api.get<GameListResponse>(`/games${qs}`);
}

export function getGame(id: number): Promise<GameDetail> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error('id must be a number >= 1'));
    }
    return api.get<GameDetail>(`/games/${id}`);
}