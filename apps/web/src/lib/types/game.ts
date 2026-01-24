export type GameListItem = {
    id: number;
    title: string;
    slug: string;
    thumbnail: string | null;
    game_url: string | null;
    age_category_id: number;
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
    age_category_id: number;
    free: boolean;
    created_at: string;
};
