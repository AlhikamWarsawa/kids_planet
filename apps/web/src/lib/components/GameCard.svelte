<script lang="ts">
    import type { GameListItem } from '$lib/types/game';
    import { formatGameAgeTag, resolveGameIconUrl } from '$lib/utils/gameDisplay';

    export let game: GameListItem;

    let iconError = false;
    let prevGameId = 0;

    $: if (game?.id !== prevGameId) {
        prevGameId = game?.id ?? 0;
        iconError = false;
    }

    $: ageLabel = formatGameAgeTag(game);
    $: iconUrl = resolveGameIconUrl(game);
    $: educationText = (() => {
        const names = Array.isArray(game?.education_categories)
            ? game.education_categories
                .map((category) => String(category?.name ?? '').trim())
                .filter((name) => name !== '')
            : [];

        if (names.length > 0) return names.join(', ');
        return '';
    })();
</script>

<a class="card" href={`/player/games/${game.id}`} aria-label={`Play ${game.title}`}>
    <div class="thumb">
        {#if iconUrl && !iconError}
            <img
                    src={iconUrl}
                    alt={game.title}
                    loading="lazy"
                    decoding="async"
                    on:error={() => (iconError = true)}
            />
        {:else}
            <div class="thumb-fallback" aria-hidden="true">
                🎮
            </div>
        {/if}

        <div class="badge">{ageLabel}</div>
    </div>

    <div class="body">
        <div class="title" title={game.title}>{game.title}</div>
        {#if educationText}
            <div class="categories" title={educationText}>{educationText}</div>
        {/if}
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
        line-clamp: 2;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        letter-spacing: -0.2px;
    }

    .categories {
        margin-top: 6px;
        font-size: 12px;
        line-height: 1.4;
        color: #475569;
        display: -webkit-box;
        line-clamp: 2;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>
