import { browser } from "$app/environment";

export type ApiErrorBody = {
    error?: {
        code?: string;
        message?: string;
    };
};

export class ApiError extends Error {
    status: number;
    code: string;

    constructor(status: number, code: string, message: string) {
        super(message);
        this.name = "ApiError";
        this.status = status;
        this.code = code;
    }
}

type FetchLike = typeof fetch;

type ClientOptions = {
    baseUrl?: string;
    fetchFn?: FetchLike;
    getToken?: () => string | null;
};

function isFormData(body: unknown): body is FormData {
    return typeof FormData !== "undefined" && body instanceof FormData;
}

async function safeReadJson(res: Response): Promise<any | null> {
    const ct = res.headers.get("content-type") || "";
    if (!ct.includes("application/json")) return null;
    try {
        return await res.json();
    } catch {
        return null;
    }
}

export function createApiClient(opts: ClientOptions = {}) {
    const baseUrl = opts.baseUrl ?? "/api";
    const fetchFn: FetchLike = opts.fetchFn ?? fetch;
    const getToken = opts.getToken;

    async function request<T>(
        path: string,
        init: RequestInit & { token?: string } = {}
    ): Promise<T> {
        const url = path.startsWith("http")
            ? path
            : `${baseUrl}${path.startsWith("/") ? "" : "/"}${path}`;

        const headers = new Headers(init.headers ?? {});
        headers.set("Accept", "application/json");

        const shouldAttachToken = path.startsWith("/admin");
        const token =
            init.token ?? (shouldAttachToken && getToken ? getToken() : null);

        if (token) headers.set("Authorization", `Bearer ${token}`);

        let body = init.body as any;
        const isPlainObject =
            body &&
            typeof body === "object" &&
            !isFormData(body) &&
            !(body instanceof Blob) &&
            !(body instanceof ArrayBuffer);

        if (isPlainObject && !(body instanceof URLSearchParams)) {
            headers.set("Content-Type", "application/json");
            body = JSON.stringify(body);
        } else {
            if (isFormData(body)) {
                headers.delete("Content-Type");
            }
        }

        const res = await fetchFn(url, {
            ...init,
            headers,
            body,
        });

        if (res.status === 204) return undefined as unknown as T;

        const json = await safeReadJson(res);

        if (!res.ok) {
            const code =
                (json as ApiErrorBody | null)?.error?.code ||
                (res.status === 401 ? "UNAUTHORIZED" : "HTTP_ERROR");
            const message =
                (json as ApiErrorBody | null)?.error?.message ||
                res.statusText ||
                "Request failed";

            throw new ApiError(res.status, code, message);
        }

        if (json && typeof json === "object" && "data" in json) {
            if ("pagination" in json) {
                return json as T;
            }
            return (json as any).data as T;
        }

        return json as T;
    }

    return {
        request,

        get: <T>(path: string, init?: RequestInit & { token?: string }) =>
            request<T>(path, { ...init, method: "GET" }),

        post: <T>(path: string, body?: any, init?: RequestInit & { token?: string }) =>
            request<T>(path, { ...init, method: "POST", body }),

        put: <T>(path: string, body?: any, init?: RequestInit & { token?: string }) =>
            request<T>(path, { ...init, method: "PUT", body }),

        del: <T>(path: string, init?: RequestInit & { token?: string }) =>
            request<T>(path, { ...init, method: "DELETE" }),
    };
}

const ADMIN_STORAGE_KEY = "kidsplanet_admin_token";

function getAdminTokenFromStorage(): string | null {
    if (!browser) return null;
    try {
        const t = localStorage.getItem(ADMIN_STORAGE_KEY);
        return t?.trim() || null;
    } catch {
        return null;
    }
}

export const api = createApiClient({
    getToken: getAdminTokenFromStorage,
});
