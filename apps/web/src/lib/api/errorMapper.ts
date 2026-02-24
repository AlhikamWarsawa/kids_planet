import { ApiError } from "$lib/api/client";

export type MappedApiError = {
    status: number;
    code: string;
    message: string;
    requestId: string | null;
    isApiError: boolean;
};

function normalizeText(value: unknown): string {
    if (typeof value !== "string") return "";
    return value.trim();
}

export function mapApiError(err: unknown, fallbackMessage = "Request failed"): MappedApiError {
    const fallback = normalizeText(fallbackMessage) || "Request failed";

    if (err instanceof ApiError) {
        const message = normalizeText(err.message) || fallback;
        return {
            status: Number.isFinite(err.status) ? err.status : 0,
            code: normalizeText(err.code) || "HTTP_ERROR",
            message,
            requestId: normalizeText(err.requestId) || null,
            isApiError: true,
        };
    }

    if (err instanceof Error) {
        const message = normalizeText(err.message) || fallback;
        return {
            status: 0,
            code: "UNKNOWN_ERROR",
            message,
            requestId: null,
            isApiError: false,
        };
    }

    return {
        status: 0,
        code: "UNKNOWN_ERROR",
        message: fallback,
        requestId: null,
        isApiError: false,
    };
}

export function formatMappedError(
    mapped: MappedApiError,
    opts?: {
        includeCode?: boolean;
        includeRequestId?: boolean;
    }
): string {
    const includeCode = opts?.includeCode ?? false;
    const includeRequestId = opts?.includeRequestId ?? true;

    const message = includeCode ? `${mapped.code}: ${mapped.message}` : mapped.message;
    if (!includeRequestId || !mapped.requestId) return message;
    return `${message}\nRequest ID: ${mapped.requestId}`;
}
