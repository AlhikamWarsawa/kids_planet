type AgeSource = {
    age_rating?: string | null;
    age_range?: unknown;
    min_age?: number | null;
    max_age?: number | null;
    age_category_id?: number | null;
    age_label?: string | null;
    age_category_label?: string | null;
};

type IconSource = {
    id?: number | null;
    slug?: string | null;
    icon?: string | null;
    game_url?: string | null;
};

type AgeRange = {
    min: number;
    max: number | null;
};

function isFiniteNumber(value: unknown): value is number {
    return typeof value === "number" && Number.isFinite(value);
}

function toIntOrNull(value: unknown): number | null {
    if (isFiniteNumber(value)) return Math.trunc(value);
    if (typeof value === "string" && value.trim() !== "") {
        const num = Number(value.trim());
        if (Number.isFinite(num)) return Math.trunc(num);
    }
    return null;
}

function parseRangeFromString(value: string): AgeRange | null {
    const cleaned = value.trim().toLowerCase();
    if (!cleaned) return null;

    const plusMatch = cleaned.match(/(\d{1,2})\s*\+/);
    if (plusMatch) {
        const min = Number(plusMatch[1]);
        if (Number.isFinite(min)) return { min, max: null };
    }

    const rangeMatch = cleaned.match(/(\d{1,2})\s*[-–]\s*(\d{1,2})/);
    if (rangeMatch) {
        const min = Number(rangeMatch[1]);
        const max = Number(rangeMatch[2]);
        if (Number.isFinite(min) && Number.isFinite(max)) {
            return { min, max: Math.max(min, max) };
        }
    }

    const singleMatch = cleaned.match(/(\d{1,2})/);
    if (singleMatch) {
        const min = Number(singleMatch[1]);
        if (Number.isFinite(min)) return { min, max: null };
    }

    return null;
}

function parseRangeFromUnknown(value: unknown): AgeRange | null {
    if (value == null) return null;

    if (typeof value === "string") {
        return parseRangeFromString(value);
    }

    if (typeof value === "object") {
        const obj = value as Record<string, unknown>;
        const min = toIntOrNull(obj.min ?? obj.from ?? obj.start ?? obj.minimum);
        const max = toIntOrNull(obj.max ?? obj.to ?? obj.end ?? obj.maximum);
        if (min == null && max == null) return null;
        if (min == null && max != null) return { min: max, max: null };
        if (min != null && max == null) return { min, max: null };
        return { min: min as number, max: Math.max(min as number, max as number) };
    }

    return null;
}

function normalizeRange(range: AgeRange | null): AgeRange | null {
    if (!range) return null;
    if (range.min < 0) return null;
    if (range.max != null && range.max < range.min) {
        return { min: range.min, max: range.min };
    }
    return range;
}

function rangeToLabel(range: AgeRange): string {
    if (range.max == null || range.max <= range.min) return `Age ${range.min}+`;
    return `Age ${range.min}\u2013${range.max}`;
}

export function formatGameAgeTag(source: AgeSource): string {
    const fromRating = normalizeRange(parseRangeFromString(source.age_rating ?? ""));
    if (fromRating) return rangeToLabel(fromRating);

    const fromRange = normalizeRange(parseRangeFromUnknown(source.age_range));
    if (fromRange) return rangeToLabel(fromRange);

    const minAge = toIntOrNull(source.min_age);
    const maxAge = toIntOrNull(source.max_age);
    if (minAge != null || maxAge != null) {
        const range = normalizeRange({
            min: minAge ?? (maxAge as number),
            max: maxAge,
        });
        if (range) return rangeToLabel(range);
    }

    const freeLabel = (source.age_label ?? source.age_category_label ?? "").trim();
    const parsedFreeLabel = parseRangeFromString(freeLabel);
    if (parsedFreeLabel) return rangeToLabel(parsedFreeLabel);

    return "Age N/A";
}

function isAbsoluteUrl(value: string): boolean {
    return /^(https?:)?\/\//i.test(value);
}

function getMinioBaseUrl(): string {
    const env = import.meta.env as Record<string, string | undefined>;
    const raw =
        env.PUBLIC_MINIO_BASE_URL ??
        env.VITE_MINIO_BASE_URL ??
        env.PUBLIC_MINIO_URL ??
        env.VITE_MINIO_URL ??
        "";
    return raw.trim().replace(/\/+$/, "");
}

function getDefaultGameIconName(): string {
    const env = import.meta.env as Record<string, string | undefined>;
    const raw =
        env.PUBLIC_GAME_ICON_NAME ??
        env.VITE_GAME_ICON_NAME ??
        env.PUBLIC_GAME_ICON_FILE ??
        env.VITE_GAME_ICON_FILE ??
        "icon.png";
    return raw.trim() || "icon.png";
}

function joinPath(base: string, tail: string): string {
    return `${base.replace(/\/+$/, "")}/${tail.replace(/^\/+/, "")}`;
}

function dirnamePath(pathname: string): string {
    const idx = pathname.lastIndexOf("/");
    if (idx <= 0) return "/";
    return pathname.slice(0, idx);
}

export function resolveGameIconUrl(source: IconSource): string | null {
    const explicitIcon = (source.icon ?? "").trim();
    if (explicitIcon) {
        if (isAbsoluteUrl(explicitIcon)) return explicitIcon;
        if (explicitIcon.startsWith("/")) return explicitIcon;
    }

    const iconName = explicitIcon || getDefaultGameIconName();
    if (!iconName) return null;

    const gameUrl = (source.game_url ?? "").trim();
    if (gameUrl) {
        if (isAbsoluteUrl(gameUrl)) {
            try {
                const url = new URL(gameUrl);
                const path = joinPath(dirnamePath(url.pathname), iconName);
                return `${url.origin}${path}`;
            } catch {
            }
        } else {
            const path = joinPath(dirnamePath(gameUrl), iconName);
            return path;
        }
    }

    if (source.id && Number.isFinite(source.id) && source.id > 0) {
        return `/games/${Math.trunc(source.id)}/current/${iconName}`;
    }

    const minioBase = getMinioBaseUrl();
    if (minioBase) return joinPath(minioBase, iconName);

    if (source.slug?.trim()) {
        return `/games/${source.slug.trim()}/current/${iconName}`;
    }

    return null;
}

export function normalizePlayCount(value: unknown): number {
    const count = toIntOrNull(value);
    if (count == null || count < 0) return 0;
    return count;
}
