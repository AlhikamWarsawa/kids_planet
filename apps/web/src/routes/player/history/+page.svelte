<script lang="ts">
    import { onMount } from "svelte";

    import { ApiError } from "$lib/api/client";
    import { getPlayerHistory, type PlayerHistoryItem } from "$lib/api/history";
    import { isLoggedIn } from "$lib/auth/playerAuth";

    const DEFAULT_LIMIT = 10;
    const rtf = new Intl.RelativeTimeFormat("id-ID", { numeric: "auto" });

    let items: PlayerHistoryItem[] = [];
    let loading = true;
    let loadingPage = false;
    let errorMsg: string | null = null;
    let unauthorized = false;

    let page = 1;
    let limit = DEFAULT_LIMIT;
    let total = 0;

    $: latest = items.length > 0 ? items[0] : null;
    $: canPrev = page > 1 && !loadingPage && !loading;
    $: canNext = page * limit < total && !loadingPage && !loading;

    function formatRelative(iso: string): string {
        const t = Date.parse(iso);
        if (!Number.isFinite(t)) return "-";

        const diffSeconds = Math.round((t - Date.now()) / 1000);
        const abs = Math.abs(diffSeconds);

        if (abs < 60) return rtf.format(diffSeconds, "second");
        if (abs < 3600) return rtf.format(Math.round(diffSeconds / 60), "minute");
        if (abs < 86400) return rtf.format(Math.round(diffSeconds / 3600), "hour");
        if (abs < 2592000) return rtf.format(Math.round(diffSeconds / 86400), "day");

        return rtf.format(Math.round(diffSeconds / 2592000), "month");
    }

    function formatAbsolute(iso: string): string {
        const dt = new Date(iso);
        if (Number.isNaN(dt.getTime())) return "-";
        return dt.toLocaleString("id-ID", {
            dateStyle: "medium",
            timeStyle: "short",
        });
    }

    async function load(nextPage = page) {
        const initial = loading;
        if (initial) {
            loading = true;
        } else {
            loadingPage = true;
        }

        unauthorized = false;
        errorMsg = null;

        try {
            const res = await getPlayerHistory({ page: nextPage, limit: DEFAULT_LIMIT });
            items = res.data ?? [];
            page = res.pagination?.page ?? nextPage;
            limit = res.pagination?.limit ?? DEFAULT_LIMIT;
            total = res.pagination?.total ?? 0;
        } catch (e) {
            if (e instanceof ApiError) {
                unauthorized = e.status === 401;
                errorMsg =
                    e.status === 401
                        ? "Sesi player tidak valid. Silakan login ulang."
                        : e.message || "Gagal memuat history.";
            } else {
                errorMsg = "Gagal memuat history.";
            }

            if (initial) {
                items = [];
                total = 0;
                page = 1;
            }
        } finally {
            loading = false;
            loadingPage = false;
        }
    }

    function prevPage() {
        if (!canPrev) return;
        load(page - 1);
    }

    function nextPage() {
        if (!canNext) return;
        load(page + 1);
    }

    onMount(() => {
        if (!isLoggedIn()) {
            loading = false;
            unauthorized = true;
            errorMsg = "Silakan login dulu untuk melihat history.";
            return;
        }
        load(1);
    });
</script>

<svelte:head>
    <title>Player History</title>
</svelte:head>

<main class="screen">
    <header class="topbar">
        <a class="pill" href="/player">← Back</a>
        <div class="titleWrap">
            <h1>Recent Plays</h1>
            <p>Lihat progres bermain terbaru kamu.</p>
        </div>
    </header>

    {#if loading}
        <section class="panel">
            <div class="headline skeleton"></div>
            <div class="subline skeleton"></div>
            <div class="list">
                {#each Array(5) as _, i (i)}
                    <article class="row skeletonRow" aria-hidden="true">
                        <div class="line w40"></div>
                        <div class="line w25"></div>
                        <div class="line w15"></div>
                    </article>
                {/each}
            </div>
        </section>
    {:else if errorMsg && items.length === 0}
        <section class="state error" role="alert">
            <div class="stateTitle">Gagal memuat history</div>
            <div class="stateMsg">{errorMsg}</div>
            <div class="actions">
                {#if unauthorized}
                    <a class="pill" href="/login?next=/player/history">Login</a>
                {:else}
                    <button class="pill" type="button" on:click={() => load(page)}>Retry</button>
                {/if}
            </div>
        </section>
    {:else if items.length === 0}
        <section class="state">
            <div class="stateTitle">Belum ada history</div>
            <div class="stateMsg">Mainkan game dulu, nanti recent plays muncul di sini.</div>
        </section>
    {:else}
        {#if latest}
            <section class="progress">
                <div class="progressLabel">Terakhir dimainkan</div>
                <div class="progressTitle">{latest.title}</div>
                <div class="progressMeta">{formatRelative(latest.played_at)}</div>
            </section>
        {/if}

        {#if errorMsg}
            <div class="banner" role="alert">
                <span>{errorMsg}</span>
                <button class="pill sm" type="button" on:click={() => load(page)}>Retry</button>
            </div>
        {/if}

        <section class="panel">
            <div class="list">
                {#each items as item (item.game_id + item.played_at)}
                    <article class="row">
                        <div class="main">
                            <h2>{item.title}</h2>
                            <p>{formatRelative(item.played_at)} · {formatAbsolute(item.played_at)}</p>
                        </div>
                        {#if typeof item.score === "number"}
                            <div class="score">Score: {item.score}</div>
                        {:else}
                            <div class="score muted">No score</div>
                        {/if}
                    </article>
                {/each}
            </div>

            <footer class="pager">
                <button class="pill" type="button" disabled={!canPrev} on:click={prevPage}>Prev</button>
                <div class="meta">
                    Page <b>{page}</b> · Showing <b>{items.length}</b> of <b>{total}</b>
                </div>
                <button class="pill" type="button" disabled={!canNext} on:click={nextPage}>Next</button>
            </footer>
        </section>
    {/if}
</main>

<style>
    .screen {
        min-height: 100vh;
        padding: 16px;
        background: linear-gradient(180deg, #f8fafc 0%, #ffffff 40%);
        font-family: system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif;
        box-sizing: border-box;
    }

    .topbar {
        display: flex;
        gap: 12px;
        align-items: center;
        margin-bottom: 16px;
        flex-wrap: wrap;
    }

    .titleWrap h1 {
        margin: 0;
        font-size: clamp(22px, 4.8vw, 30px);
        color: #0f172a;
    }

    .titleWrap p {
        margin: 2px 0 0;
        color: #475569;
        font-size: 13px;
        font-weight: 600;
    }

    .pill {
        padding: 10px 16px;
        border: 1.5px solid #cbd5e1;
        background: #ffffff;
        border-radius: 999px;
        color: #0f172a;
        font-weight: 800;
        text-decoration: none;
        cursor: pointer;
    }

    .pill:hover {
        background: #f8fafc;
    }

    .pill:disabled {
        opacity: 0.55;
        cursor: not-allowed;
    }

    .pill.sm {
        padding: 8px 12px;
        font-size: 12px;
    }

    .progress {
        border: 1.5px solid #cbd5e1;
        border-radius: 16px;
        background: linear-gradient(120deg, #fff7ed 0%, #ffedd5 100%);
        padding: 14px 16px;
        margin-bottom: 12px;
    }

    .progressLabel {
        color: #9a3412;
        text-transform: uppercase;
        letter-spacing: 0.06em;
        font-size: 11px;
        font-weight: 900;
    }

    .progressTitle {
        font-size: 18px;
        font-weight: 900;
        color: #7c2d12;
        margin-top: 4px;
    }

    .progressMeta {
        font-size: 13px;
        color: #9a3412;
        margin-top: 2px;
        font-weight: 700;
    }

    .banner {
        margin: 0 0 12px;
        padding: 10px 12px;
        border: 1.5px solid #fda4af;
        background: #fff1f2;
        border-radius: 12px;
        display: flex;
        justify-content: space-between;
        gap: 10px;
        align-items: center;
        color: #9f1239;
        font-size: 13px;
        font-weight: 700;
    }

    .panel {
        border: 1.5px solid #e2e8f0;
        border-radius: 16px;
        padding: 12px;
        background: #ffffff;
    }

    .list {
        display: grid;
        gap: 10px;
    }

    .row {
        border: 1.5px solid #e2e8f0;
        border-radius: 14px;
        padding: 12px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
    }

    .main {
        min-width: 0;
    }

    .main h2 {
        margin: 0;
        font-size: 16px;
        color: #0f172a;
    }

    .main p {
        margin: 4px 0 0;
        font-size: 12px;
        color: #64748b;
        font-weight: 600;
    }

    .score {
        white-space: nowrap;
        border-radius: 999px;
        padding: 8px 12px;
        background: #f1f5f9;
        color: #0f172a;
        font-size: 12px;
        font-weight: 800;
    }

    .score.muted {
        color: #64748b;
    }

    .pager {
        margin-top: 14px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
    }

    .meta {
        color: #475569;
        font-size: 13px;
        font-weight: 600;
    }

    .state {
        border: 1.5px solid #e2e8f0;
        border-radius: 16px;
        background: #ffffff;
        padding: 16px;
    }

    .state.error {
        border-color: #fca5a5;
        background: #fff5f5;
    }

    .stateTitle {
        font-size: 17px;
        font-weight: 900;
        color: #0f172a;
    }

    .stateMsg {
        margin-top: 4px;
        color: #475569;
        font-size: 13px;
        font-weight: 600;
    }

    .actions {
        margin-top: 12px;
        display: flex;
        gap: 10px;
        flex-wrap: wrap;
    }

    .skeleton {
        border-radius: 12px;
        background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 38%, #f1f5f9 60%);
        background-size: 200% 100%;
        animation: pulse 1.2s infinite linear;
    }

    .headline {
        height: 28px;
        width: min(280px, 70%);
    }

    .subline {
        margin-top: 10px;
        height: 16px;
        width: min(380px, 90%);
    }

    .skeletonRow {
        display: block;
        padding: 12px;
    }

    .line {
        height: 12px;
        border-radius: 999px;
        background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 38%, #f1f5f9 60%);
        background-size: 200% 100%;
        animation: pulse 1.2s infinite linear;
    }

    .line + .line {
        margin-top: 8px;
    }

    .w40 {
        width: 40%;
    }

    .w25 {
        width: 25%;
    }

    .w15 {
        width: 15%;
    }

    @keyframes pulse {
        from {
            background-position: 200% 0;
        }
        to {
            background-position: -200% 0;
        }
    }

    @media (max-width: 640px) {
        .row {
            flex-direction: column;
            align-items: flex-start;
        }

        .score {
            align-self: flex-start;
        }

        .w40,
        .w25,
        .w15 {
            width: 90%;
        }
    }
</style>
