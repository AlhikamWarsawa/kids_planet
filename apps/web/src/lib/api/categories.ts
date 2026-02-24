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

export type PublicCategoriesResponse = {
    age_categories: AgeCategoryDTO[];
    education_categories: EducationCategoryDTO[];
};

function toInt(value: unknown): number | null {
    if (typeof value === "number" && Number.isFinite(value)) return Math.trunc(value);
    if (typeof value === "string" && value.trim() !== "") {
        const parsed = Number(value.trim());
        if (Number.isFinite(parsed)) return Math.trunc(parsed);
    }
    return null;
}

function toText(value: unknown): string {
    if (typeof value !== "string") return "";
    return value.trim();
}

function normalizeAgeCategories(value: unknown): AgeCategoryDTO[] {
    if (!Array.isArray(value)) return [];
    const out: AgeCategoryDTO[] = [];
    for (const raw of value) {
        const row = (raw && typeof raw === "object" ? raw : {}) as Record<string, unknown>;
        const id = toInt(row.id ?? row.ID);
        if (id == null || id < 1) continue;
        const createdAt = toText(row.created_at ?? row.createdAt ?? row.CreatedAt);
        out.push({
            id,
            label: toText(row.label ?? row.Label),
            min_age: toInt(row.min_age ?? row.minAge ?? row.MinAge) ?? 0,
            max_age: toInt(row.max_age ?? row.maxAge ?? row.MaxAge) ?? 0,
            ...(createdAt ? { created_at: createdAt } : {}),
        });
    }
    return out;
}

function normalizeEducationCategories(value: unknown): EducationCategoryDTO[] {
    if (!Array.isArray(value)) return [];
    const out: EducationCategoryDTO[] = [];
    for (const raw of value) {
        const row = (raw && typeof raw === "object" ? raw : {}) as Record<string, unknown>;
        const id = toInt(row.id ?? row.ID);
        if (id == null || id < 1) continue;
        const createdAt = toText(row.created_at ?? row.createdAt ?? row.CreatedAt);

        out.push({
            id,
            name: toText(row.name ?? row.Name),
            ...(createdAt ? { created_at: createdAt } : {}),
        });
    }
    return out;
}

function normalizePublicCategories(raw: unknown): PublicCategoriesResponse {
    const row = (raw && typeof raw === "object" ? raw : {}) as Record<string, unknown>;
    return {
        age_categories: normalizeAgeCategories(row.age_categories ?? row.ageCategories),
        education_categories: normalizeEducationCategories(
            row.education_categories ?? row.educationCategories
        ),
    };
}

type PublicCategoriesType = "age" | "education";

function buildPublicCategoriesQuery(type?: PublicCategoriesType): string {
    const q = new URLSearchParams();
    if (type === "age" || type === "education") {
        q.set("type", type);
    }
    const qs = q.toString();
    return qs ? `?${qs}` : "";
}

function buildAdminQuery(params: AdminListParams = {}): string {
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

export function getPublicCategories(
    opts: {
        type?: PublicCategoriesType;
    } = {}
): Promise<PublicCategoriesResponse> {
    const qs = buildPublicCategoriesQuery(opts.type);
    return api.get<any>(`/categories${qs}`).then((raw) => normalizePublicCategories(raw));
}

export function getPublicEducationCategories(): Promise<EducationCategoryDTO[]> {
    return getPublicCategories({ type: "education" }).then((raw) => raw.education_categories);
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
    const qs = buildAdminQuery(params);
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
};

export type AdminUpdateEducationCategoryRequest = {
    name?: string;
};

export function adminListEducationCategories(
    params: AdminListParams = {}
): Promise<AdminListResponse<EducationCategoryDTO>> {
    const qs = buildAdminQuery(params);
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
