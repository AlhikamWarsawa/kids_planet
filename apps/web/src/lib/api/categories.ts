import { api } from "$lib/api/client";

export type AgeCategoryDTO = {
    id: number;
    label: string;
    min_age: number;
    max_age: number;
    created_at?: string;
};

export type EducationCategoryDTO = {
    id: number;
    name: string;
    icon: string | null;
    color: string | null;
    created_at?: string;
};

export type AdminListParams = {
    q?: string;
    page?: number;
    limit?: number;
};

export type AdminListResponse<T> = {
    items: T[];
    page: number;
    limit: number;
};

function buildQuery(params: AdminListParams = {}): string {
    const q = new URLSearchParams();

    if (typeof params.q === "string" && params.q.trim() !== "") {
        q.set("q", params.q.trim());
    }
    if (typeof params.page === "number") {
        q.set("page", String(params.page));
    }
    if (typeof params.limit === "number") {
        q.set("limit", String(params.limit));
    }

    const qs = q.toString();
    return qs ? `?${qs}` : "";
}

// Age Categories

export type AdminCreateAgeCategoryRequest = {
    label: string;
    min_age: number;
    max_age: number;
};

export type AdminUpdateAgeCategoryRequest = {
    label?: string;
    min_age?: number;
    max_age?: number;
};

export function adminListAgeCategories(
    params: AdminListParams = {}
): Promise<AdminListResponse<AgeCategoryDTO>> {
    const qs = buildQuery(params);
    return api.get<AdminListResponse<AgeCategoryDTO>>(`/admin/age-categories${qs}`);
}

export function adminCreateAgeCategory(
    payload: AdminCreateAgeCategoryRequest
): Promise<AgeCategoryDTO> {
    if (!payload || typeof payload !== "object") {
        return Promise.reject(new Error("payload is required"));
    }
    return api.post<AgeCategoryDTO>(`/admin/age-categories`, payload);
}

export function adminUpdateAgeCategory(
    id: number,
    payload: AdminUpdateAgeCategoryRequest
): Promise<AgeCategoryDTO> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error("id must be a number >= 1"));
    }
    if (!payload || typeof payload !== "object") {
        return Promise.reject(new Error("payload is required"));
    }
    return api.put<AgeCategoryDTO>(`/admin/age-categories/${id}`, payload);
}

export function adminDeleteAgeCategory(
    id: number
): Promise<{ deleted: true }> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error("id must be a number >= 1"));
    }
    return api.del<{ deleted: true }>(`/admin/age-categories/${id}`);
}

// Education Categories
export type AdminCreateEducationCategoryRequest = {
    name: string;
    icon?: string | null;
    color?: string | null;
};

export type AdminUpdateEducationCategoryRequest = {
    name?: string;
    icon?: string | null;
    color?: string | null;
};

export function adminListEducationCategories(
    params: AdminListParams = {}
): Promise<AdminListResponse<EducationCategoryDTO>> {
    const qs = buildQuery(params);
    return api.get<AdminListResponse<EducationCategoryDTO>>(
        `/admin/education-categories${qs}`
    );
}

export function adminCreateEducationCategory(
    payload: AdminCreateEducationCategoryRequest
): Promise<EducationCategoryDTO> {
    if (!payload || typeof payload !== "object") {
        return Promise.reject(new Error("payload is required"));
    }
    return api.post<EducationCategoryDTO>(`/admin/education-categories`, payload);
}

export function adminUpdateEducationCategory(
    id: number,
    payload: AdminUpdateEducationCategoryRequest
): Promise<EducationCategoryDTO> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error("id must be a number >= 1"));
    }
    if (!payload || typeof payload !== "object") {
        return Promise.reject(new Error("payload is required"));
    }
    return api.put<EducationCategoryDTO>(`/admin/education-categories/${id}`, payload);
}

export function adminDeleteEducationCategory(
    id: number
): Promise<{ deleted: true }> {
    if (!Number.isFinite(id) || id < 1) {
        return Promise.reject(new Error("id must be a number >= 1"));
    }
    return api.del<{ deleted: true }>(`/admin/education-categories/${id}`);
}