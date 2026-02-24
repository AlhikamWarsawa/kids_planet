export type GameListItem = {
    id: number;
    title: string;
    slug: string;
    thumbnail: string | null;
    game_url: string | null;
    icon?: string | null;
    age_category_id: number;
    age_rating?: string | null;
    age_range?: unknown;
    min_age?: number | null;
    max_age?: number | null;
    age_label?: string | null;
    age_category_label?: string | null;
    education_category_ids?: number[];
    education_categories?: Array<{
        id: number;
        name: string;
    }>;
    play_count?: number | null;
    free: boolean;
    created_at: string;
};

export type GameListResponse = {
    items: GameListItem[];
    page: number;
    limit: number;
    total: number;
};

export type GameDetail = {
    id: number;
    title: string;
    slug: string;
    thumbnail: string | null;
    game_url: string | null;
    icon?: string | null;
    age_category_id: number;
    age_rating?: string | null;
    age_range?: unknown;
    min_age?: number | null;
    max_age?: number | null;
    age_label?: string | null;
    age_category_label?: string | null;
    education_category_ids?: number[];
    education_categories?: Array<{
        id: number;
        name: string;
    }>;
    play_count?: number | null;
    free: boolean;
    created_at: string;
};
