<script lang="ts">
    import { onMount } from 'svelte';
    import { listGames } from '$lib/api/games';
    import type { GameListItem } from '$lib/types/game';
    import { ApiError } from '$lib/api/client';

    import GameCard from '$lib/components/GameCard.svelte';
    import PickerModal from '$lib/components/PickerModal.svelte';
    import type { Option } from '$lib/types/picker';

    const ageOptions: Option[] = [
        { id: 1, label: '3 Tahun +' },
        { id: 2, label: '5 Tahun +' },
        { id: 3, label: '7 Tahun +' },
        { id: 4, label: '10 Tahun +' }
    ];

    const educationOptions: Option[] = [
        { id: 1, label: 'Math' },
        { id: 2, label: 'Reading' },
        { id: 3, label: 'Logic' },
        { id: 4, label: 'Memory' },
        { id: 5, label: 'Creativity' }
    ];

    let games: GameListItem[] = [];
    let page = 1;
    let limit = 24;
    let total = 0;

    let loading = false;
    let loadingMore = false;
    let error: string | null = null;

    let selectedAge: number | null = null;
    let selectedCategory: number | null = null;

    let sort: 'newest' | 'popular' = 'newest';

    let showAgeModal = false;
    let showCatModal = false;

    let reqSeq = 0;

    const hasMore = () => games.length < total;

    function getAgeLabel(id: number | null) {
        if (id == null) return 'Pilih Usia';
        return ageOptions.find((o) => o.id === id)?.label ?? 'Pilih Usia';
    }

    function getCatLabel(id: number | null) {
        if (id == null) return 'Pilih Kategori';
        return educationOptions.find((o) => o.id === id)?.label ?? 'Pilih Kategori';
    }

    $: ageLabel = getAgeLabel(selectedAge);
    $: catLabel = getCatLabel(selectedCategory);
    $: sortLabel = sort === 'newest' ? 'Terbaru' : 'Popular';

    async function loadInitial(opts: { keepList?: boolean } = {}) {
        const seq = ++reqSeq;

        if (!opts.keepList) loading = true;
        error = null;

        try {
            const res = await listGames({
                age_category_id: selectedAge ?? undefined,
                education_category_id: selectedCategory ?? undefined,
                sort,
                page: 1,
                limit
            });

            if (seq !== reqSeq) return;

            games = res.items;
            page = res.page;
            limit = res.limit;
            total = res.total;
        } catch (e) {
            if (seq !== reqSeq) return;

            error = e instanceof ApiError ? e.message : 'Failed to load games';
            games = [];
            total = 0;
            page = 1;
        } finally {
            if (seq === reqSeq) loading = false;
        }
    }

    async function loadMore() {
        if (loading || loadingMore || !hasMore()) return;

        const seq = ++reqSeq;
        loadingMore = true;
        error = null;

        try {
            const res = await listGames({
                age_category_id: selectedAge ?? undefined,
                education_category_id: selectedCategory ?? undefined,
                sort,
                page: page + 1,
                limit
            });

            if (seq !== reqSeq) return;

            games = [...games, ...res.items];
            page = res.page;
            limit = res.limit;
            total = res.total;
        } catch (e) {
            if (seq !== reqSeq) return;
            error = e instanceof ApiError ? e.message : 'Failed to load more games';
        } finally {
            if (seq === reqSeq) loadingMore = false;
        }
    }

    function applyFilters() {
        page = 1;
        loadInitial({ keepList: true });
    }

    function toggleSort() {
        sort = sort === 'newest' ? 'popular' : 'newest';
        applyFilters();
    }

    onMount(() => {
        loadInitial();
    });
</script>

<svelte:head>
    <title>Kids Planet Player</title>
    <meta
            name="description"
            content="Kids Planet player area. Browse educational games by age, category, and sorting, then start playing."
    />
</svelte:head>

<main class="screen">
    <div class="container">
        <div class="topbar">
            <div class="left">
                <a class="pill" href="/admin/login">Login</a>
            </div>

            <div class="right">
                <button class="pill" type="button" on:click={toggleSort}>
                    {sortLabel}
                </button>

                <button class="pill" type="button" on:click={() => (showAgeModal = true)}>
                    {ageLabel}
                </button>

                <button class="pill" type="button" on:click={() => (showCatModal = true)}>
                    {catLabel}
                </button>
            </div>
        </div>

        {#if error}
            <div class="state" role="alert">{error}</div>
        {/if}

        {#if loading}
            <div class="grid">
                {#each Array(8) as _, i (i)}
                    <div class="box skeleton" aria-hidden="true"></div>
                {/each}
            </div>
        {:else if !error && games.length === 0}
            <div class="empty">
                <div class="empty-title">Belum ada game</div>
                <div class="empty-sub">Coba ubah filter/sort atau cek lagi nanti.</div>
            </div>
        {:else}
            <div class="grid">
                {#each games as game (game.id)}
                    <GameCard {game} />
                {/each}
            </div>

            <div class="footer">
                <div class="meta">
                    Showing <b>{games.length}</b> of <b>{total}</b>
                </div>

                {#if hasMore()}
                    <button class="pill" on:click={loadMore} disabled={loadingMore}>
                        {loadingMore ? 'Loading...' : 'Load more'}
                    </button>
                {/if}
            </div>
        {/if}

        <PickerModal
                open={showAgeModal}
                title="Pilih Usia"
                options={ageOptions}
                selectedId={selectedAge}
                on:close={() => (showAgeModal = false)}
                on:clear={() => {
                selectedAge = null;
                showAgeModal = false;
                applyFilters();
            }}
                on:select={(e) => {
                selectedAge = e.detail.id;
                showAgeModal = false;
                applyFilters();
            }}
        />

        <PickerModal
                open={showCatModal}
                title="Pilih Kategori"
                options={educationOptions}
                selectedId={selectedCategory}
                on:close={() => (showCatModal = false)}
                on:clear={() => {
                selectedCategory = null;
                showCatModal = false;
                applyFilters();
            }}
                on:select={(e) => {
                selectedCategory = e.detail.id;
                showCatModal = false;
                applyFilters();
            }}
        />
    </div>
</main>

<style>
    .screen {
        min-height: calc(100vh - 80px);
        padding: 20px 16px 28px;
        font-family: system-ui, -apple-system, 'Segoe UI', Roboto, Arial, sans-serif;
        background: #fff;
        overflow-x: hidden;
    }

    .container {
        max-width: 1200px;
        margin: 0 auto;
    }

    .topbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        margin-bottom: 18px;
        flex-wrap: wrap;
    }

    .left {
        flex: 0 0 auto;
    }

    .right {
        display: flex;
        gap: 12px;
        align-items: center;
        flex-wrap: wrap;
        justify-content: flex-end;
        flex: 1 1 280px;
    }

    .pill {
        padding: 12px 18px;
        border-radius: 999px;
        border: 4px solid #666;
        background: #fff;
        font-weight: 600;
        cursor: pointer;
        text-decoration: none;
        color: #222;

        box-sizing: border-box;
        max-width: 100%;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .pill:hover {
        background: #f5f5f5;
    }

    .pill:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .grid {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 16px;
        min-width: 0;
    }

    @media (min-width: 720px) {
        .screen {
            padding: 24px 22px 34px;
        }
        .grid {
            grid-template-columns: repeat(3, minmax(0, 1fr));
            gap: 20px;
        }
    }

    @media (min-width: 1024px) {
        .screen {
            padding: 26px 26px 40px;
        }
        .grid {
            grid-template-columns: repeat(4, minmax(0, 1fr));
            gap: 24px;
        }
    }

    @media (min-width: 1280px) {
        .grid {
            grid-template-columns: repeat(5, minmax(0, 1fr));
        }
    }

    .box {
        height: 160px;
        border: 4px solid #666;
        border-radius: 12px;
        background: #fff;
        box-sizing: border-box;
    }

    .skeleton {
        position: relative;
        overflow: hidden;
    }

    .skeleton::after {
        content: '';
        position: absolute;
        inset: 0;
        transform: translateX(-100%);
        background: linear-gradient(
                90deg,
                rgba(255, 255, 255, 0) 0%,
                rgba(0, 0, 0, 0.06) 50%,
                rgba(255, 255, 255, 0) 100%
        );
        animation: shimmer 1.4s infinite;
    }

    @keyframes shimmer {
        100% {
            transform: translateX(100%);
        }
    }

    .footer {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 18px;
        gap: 12px;
        flex-wrap: wrap;
    }

    .meta {
        color: #222;
        opacity: 0.7;
        font-size: 13px;
    }

    .state {
        margin: 0 0 14px;
        padding: 14px 16px;
        border-radius: 12px;
        border: 2px solid #fecaca;
        background: #fef2f2;
        color: #991b1b;
        font-weight: 600;
        box-sizing: border-box;
    }

    .empty {
        border: 4px solid #666;
        border-radius: 12px;
        padding: 18px;
        background: #fff;
        box-sizing: border-box;
    }

    .empty-title {
        font-weight: 800;
        margin-bottom: 6px;
        color: #222;
    }

    .empty-sub {
        opacity: 0.7;
        color: #222;
        font-size: 13px;
    }
</style>
