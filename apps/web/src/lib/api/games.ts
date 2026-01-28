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

// ADMIN

export type AdminGameStatus = 'draft' | 'active' | 'archived';

export type AdminGameDTO = {
    id: number;
    title: string;
    slug: string;
    status: AdminGameStatus;
    thumbnail: string | null;
    game_url: string | null;
    age_category_id: number;
    free: boolean;
    created_at: string;
    updated_at: string;
};

export type AdminGameListResponse = {
    items: AdminGameDTO[];
    page: number;
    limit: number;
    total: number;
};

export type AdminListGamesParams = {
    status?: AdminGameStatus;
    page?: number;
    limit?: number;
    q?: string;
};

function buildAdminQuery(params: AdminListGamesParams = {}): string {
    const q = new URLSearchParams();

    if (params.status === 'draft' || params.status === 'active' || params.status === 'archived') {
        q.set('status', params.status);
    }
    if (typeof params.page === 'number') {
        q.set('page', String(params.page));
    }
    if (typeof params.limit === 'number') {
        q.set('limit', String(params.limit));
    }
    if (typeof params.q === 'string' && params.q.trim() !== '') {
        q.set('q', params.q.trim());
    }

    const qs = q.toString();
    return qs ? `?${qs}` : '';
}

export type AdminCreateGameRequest = {
    title: string;
    slug: string;
    age_category_id: number;
    thumbnail?: string | null;
    game_url?: string | null;
    free?: boolean;
};

export type AdminUpdateGameRequest = {
    title?: string;
    slug?: string;
    age_category_id?: number;
    thumbnail?: string | null;
    game_url?: string | null;
    free?: boolean;
};

export function adminListGames(params: AdminListGamesParams = {}): Promise<AdminGameListResponse> {
    const qs = buildAdminQuery(params);
    return api.get<AdminGameListResponse>(`/admin/games${qs}`);
}

export function adminCreateGame(payload: AdminCreateGameRequest): Promise<AdminGameDTO> {
    if (!payload || typeof payload !== 'object') {
        return Promise.reject(new Error('payload is required'));
    }
    return api.post<AdminGameDTO>(`/admin/games`, payload);
}

export function adminUpdateGame(id: number, payload: AdminUpdateGameRequest): Promise<AdminGameDTO> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error('id must be a number >= 1'));
    }
    if (!payload || typeof payload !== 'object') {
        return Promise.reject(new Error('payload is required'));
    }
    return api.put<AdminGameDTO>(`/admin/games/${id}`, payload);
}

export function adminPublishGame(id: number): Promise<AdminGameDTO> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error('id must be a number >= 1'));
    }
    return api.post<AdminGameDTO>(`/admin/games/${id}/publish`);
}

export function adminUnpublishGame(id: number): Promise<AdminGameDTO> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error('id must be a number >= 1'));
    }
    return api.post<AdminGameDTO>(`/admin/games/${id}/unpublish`);
}