import { api } from '$lib/api/client';
import type { GameDetail, GameListResponse } from '$lib/types/game';

export type GamesSort = 'newest' | 'popular';

export type ListGamesParams = {
    age_category_id?: number | number[];
    education_category_id?: number | number[];
    sort?: GamesSort;
    page?: number;
    limit?: number;
};

function toInt(value: unknown): number | null {
    if (typeof value === 'number' && Number.isFinite(value)) return Math.trunc(value);
    if (typeof value === 'string' && value.trim() !== '') {
        const n = Number(value.trim());
        if (Number.isFinite(n)) return Math.trunc(n);
    }
    return null;
}

function toStringOrNull(value: unknown): string | null {
    if (typeof value !== 'string') return null;
    const v = value.trim();
    return v ? v : null;
}

function toNumberArray(value: unknown): number[] {
    if (!Array.isArray(value)) return [];
    return value
        .map((v) => toInt(v))
        .filter((v): v is number => v != null && v >= 1);
}

function normalizeEducationCategories(value: unknown): Array<{
    id: number;
    name: string;
}> {
    if (!Array.isArray(value)) return [];

    return value
        .map((raw) => {
            if (!raw || typeof raw !== 'object') return null;
            const row = raw as Record<string, unknown>;
            const id = toInt(row.id ?? row.ID ?? row.Id);
            const name = toStringOrNull(row.name ?? row.Name) ?? '';
            if (!id || !name) return null;
            return {
                id,
                name
            };
        })
        .filter((row): row is NonNullable<typeof row> => row !== null);
}

function normalizePublicGame(raw: unknown): any {
    const row = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>;

    return {
        id: toInt(row.id ?? row.ID) ?? 0,
        title: String(row.title ?? row.Title ?? ''),
        slug: String(row.slug ?? row.Slug ?? ''),
        thumbnail: toStringOrNull(row.thumbnail ?? row.Thumbnail),
        game_url: toStringOrNull(row.game_url ?? row.gameUrl ?? row.GameURL),
        icon: toStringOrNull(row.icon ?? row.Icon),
        age_category_id: toInt(row.age_category_id ?? row.ageCategoryId ?? row.AgeCategoryID) ?? 0,
        age_rating: toStringOrNull(row.age_rating ?? row.ageRating ?? row.AgeRating),
        age_range: row.age_range ?? row.ageRange ?? row.AgeRange ?? null,
        min_age: toInt(row.min_age ?? row.minAge ?? row.MinAge),
        max_age: toInt(row.max_age ?? row.maxAge ?? row.MaxAge),
        age_label: toStringOrNull(row.age_label ?? row.ageLabel ?? row.AgeLabel),
        age_category_label: toStringOrNull(
            row.age_category_label ?? row.ageCategoryLabel ?? row.AgeCategoryLabel
        ),
        education_category_ids: toNumberArray(
            row.education_category_ids ?? row.educationCategoryIds ?? row.EducationCategoryIDs
        ),
        education_categories: normalizeEducationCategories(
            row.education_categories ?? row.educationCategories ?? row.EducationCategories
        ),
        play_count: toInt(row.play_count ?? row.playCount ?? row.Plays),
        free: Boolean(row.free ?? row.Free),
        created_at: String(row.created_at ?? row.createdAt ?? row.CreatedAt ?? '')
    };
}

function normalizeListResponse(raw: unknown): GameListResponse {
    const row = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>;
    const itemsRaw = Array.isArray(row.items) ? row.items : [];

    return {
        items: itemsRaw.map((item) => normalizePublicGame(item)),
        page: toInt(row.page) ?? 1,
        limit: toInt(row.limit) ?? 24,
        total: toInt(row.total) ?? 0
    };
}

function normalizeSingleResponse(raw: unknown): GameDetail {
    return normalizePublicGame(raw) as GameDetail;
}

function appendCategoryParams(
    q: URLSearchParams,
    idKey: 'age_category_id' | 'education_category_id',
    aliasKey: 'age' | 'education',
    value: number | number[] | undefined
) {
    if (typeof value === 'number') {
        const normalized = Math.trunc(value);
        if (normalized >= 1) {
            q.set(idKey, String(normalized));
            q.set(aliasKey, String(normalized));
        }
        return;
    }

    if (Array.isArray(value) && value.length > 0) {
        const normalized = value
            .map((v) => Math.trunc(v))
            .filter((v) => Number.isFinite(v) && v >= 1);
        if (normalized.length === 0) return;

        if (normalized.length === 1) {
            q.set(idKey, String(normalized[0]));
        }
        q.set(aliasKey, normalized.join(','));
    }
}

function buildQuery(params: ListGamesParams = {}): string {
    const q = new URLSearchParams();

    appendCategoryParams(q, 'age_category_id', 'age', params.age_category_id);
    appendCategoryParams(q, 'education_category_id', 'education', params.education_category_id);

    if (params.sort === 'newest' || params.sort === 'popular') {
        q.set('sort', params.sort);
        if (params.sort === 'popular') {
            q.set('sort_by', 'play_count');
            q.set('order', 'desc');
        }
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
    return api.get<any>(`/games${qs}`).then((raw) => normalizeListResponse(raw));
}

export function getGame(id: number): Promise<GameDetail> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error('id must be a number >= 1'));
    }
    return api.get<any>(`/games/${id}`).then((raw) => normalizeSingleResponse(raw));
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
    icon?: string | null;
    age_category_id: number;
    age_rating?: string | null;
    age_range?: unknown;
    min_age?: number | null;
    max_age?: number | null;
    education_category_ids?: number[];
    education_categories?: Array<{
        id: number;
        name: string;
    }>;
    play_count?: number | null;
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
    education_category_ids?: number[];
    thumbnail?: string | null;
    game_url?: string | null;
    free?: boolean;
};

export type AdminUpdateGameRequest = {
    title?: string;
    slug?: string;
    age_category_id?: number;
    education_category_ids?: number[];
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

export type AdminUploadZipResult = {
    object_key: string;
    etag: string;
    size: number;
    game_url: string;
};

export function adminUploadGameZip(id: number, file: File): Promise<AdminUploadZipResult> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error('id must be a number >= 1'));
    }
    if (!(file instanceof File)) {
        return Promise.reject(new Error('file is required'));
    }

    const fd = new FormData();
    fd.append('file', file, file.name);

    return api.post<AdminUploadZipResult>(`/admin/games/${id}/upload`, fd as any);
}
