<script lang="ts">
    import type { GameListItem } from '$lib/types/game';

    export let game: GameListItem;

    const ageLabel = (ageCategoryId: number) => `${ageCategoryId}+`;
</script>

<a class="card" href={`/player/games/${game.id}`} aria-label={`Play ${game.title}`}>
    <div class="thumb">
        {#if game.thumbnail}
            <img
                    src={game.thumbnail}
                    alt={game.title}
                    loading="lazy"
                    decoding="async"
            />
        {:else}
            <div class="thumb-fallback" aria-hidden="true">
                🎮
            </div>
        {/if}

        <div class="badge">{ageLabel(game.age_category_id)}</div>
    </div>

    <div class="body">
        <div class="title" title={game.title}>{game.title}</div>
    </div>
</a>

<style>
    .card {
        display: block;
        text-decoration: none;
        color: inherit;
        border-radius: 20px;
        overflow: hidden;
        background: #fff;
        border: 1px solid #e5e5e5;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
        transition: all 0.2s ease;
        -webkit-tap-highlight-color: transparent;
    }

    .card:focus-visible {
        outline: 2px solid #111;
        outline-offset: 2px;
    }

    @media (hover: hover) and (pointer: fine) {
        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12);
            border-color: #d0d0d0;
        }
    }

    .card:active {
        transform: translateY(-1px);
    }

    .thumb {
        position: relative;
        aspect-ratio: 16 / 9;
        background: #fafafa;
    }

    .thumb img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
    }

    .thumb-fallback {
        width: 100%;
        height: 100%;
        display: grid;
        place-items: center;
        font-size: 32px;
        opacity: 0.6;
    }

    .badge {
        position: absolute;
        top: 12px;
        left: 12px;
        font-size: 12px;
        font-weight: 600;
        padding: 6px 12px;
        border-radius: 999px;
        background: rgba(17, 17, 17, 0.85);
        color: #fff;
        border: 1px solid rgba(255, 255, 255, 0.1);
        backdrop-filter: blur(8px);
        letter-spacing: 0.2px;
    }

    .body {
        padding: 14px 16px 16px;
    }

    .title {
        font-size: 14px;
        font-weight: 700;
        line-height: 1.3;
        color: #1a1a1a;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        letter-spacing: -0.2px;
    }
</style>