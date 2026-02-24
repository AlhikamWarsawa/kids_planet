<script lang="ts">
    import { onMount } from "svelte";
    import Spinner from "$lib/components/Spinner.svelte";
    import { formatMappedError, mapApiError } from "$lib/api/errorMapper";
    import {
        adminDashboardOverview,
        type DashboardOverviewDTO,
    } from "$lib/api/dashboard";

    let loading = true;
    let errorMsg: string | null = null;
    let overview: DashboardOverviewDTO | null = null;
    let lastUpdated: string | null = null;

    function formatNumber(value: number) {
        if (!Number.isFinite(value)) return String(value);
        return value.toLocaleString();
    }

    async function loadOverview() {
        loading = true;
        errorMsg = null;

        try {
            overview = await adminDashboardOverview();
            lastUpdated = new Date().toLocaleString();
        } catch (err) {
            errorMsg = formatMappedError(mapApiError(err, "Dashboard failed to load"), {
                includeCode: false,
                includeRequestId: true,
            });
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        void loadOverview();
    });
</script>

<svelte:head>
    <title>Admin Dashboard</title>
</svelte:head>

<div class="page">
    <div class="header">
        <div>
            <h1>Dashboard</h1>
            <p class="sub">Live snapshot of platform activity.</p>
        </div>
        <div class="actions">
            <button on:click={loadOverview} disabled={loading}>
                {loading ? "Refreshing…" : "Refresh"}
            </button>
        </div>
    </div>

    {#if loading}
        <div class="card loading">
            <Spinner />
            <div>Loading overview…</div>
        </div>
    {:else if errorMsg}
        <div class="card error">
            <div class="error-title">Dashboard failed to load</div>
            <div class="error-msg">{errorMsg}</div>
            <button on:click={loadOverview}>Try again</button>
        </div>
    {:else if overview}
        <div class="grid">
            <div class="card metric">
                <div class="label">Sessions Today (UTC)</div>
                <div class="value">{formatNumber(overview.sessions_today)}</div>
            </div>
            <div class="card metric">
                <div class="label">Total Active Games</div>
                <div class="value">{formatNumber(overview.total_active_games)}</div>
            </div>
            <div class="card metric">
                <div class="label">Total Players</div>
                <div class="value">{formatNumber(overview.total_players)}</div>
            </div>
        </div>

        <div class="card top-games">
            <div class="label">Top Games</div>
            {#if overview.top_games?.length}
                <ol class="top-list">
                    {#each overview.top_games as game, index}
                        <li>
                            <span class="rank">#{index + 1}</span>
                            <span class="title">{game.title}</span>
                            <span class="plays">{formatNumber(game.plays)} plays</span>
                        </li>
                    {/each}
                </ol>
            {:else}
                <div class="empty">No sessions yet.</div>
            {/if}
        </div>

        {#if lastUpdated}
            <div class="updated">Last updated: {lastUpdated}</div>
        {/if}
    {/if}
</div>

<style>
    .page {
        display: flex;
        flex-direction: column;
        gap: 16px;
        max-width: 1000px;
    }

    .header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        flex-wrap: wrap;
    }

    h1 {
        margin: 0 0 4px;
    }

    .sub {
        margin: 0;
        opacity: 0.7;
    }

    .actions button {
        padding: 8px 12px;
        border-radius: 10px;
        border: 1px solid #d7d7d7;
        background: #fff;
        cursor: pointer;
    }

    .actions button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: 12px;
    }

    .card {
        background: #fff;
        border: 1px solid #eee;
        border-radius: 14px;
        padding: 16px;
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
    }

    .metric .label {
        font-size: 12px;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        opacity: 0.6;
    }

    .metric .value {
        font-size: 28px;
        font-weight: 700;
        margin-top: 6px;
    }

    .top-games .label {
        font-weight: 600;
        margin-bottom: 10px;
    }

    .top-list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: grid;
        gap: 10px;
    }

    .top-list li {
        display: grid;
        grid-template-columns: auto 1fr auto;
        gap: 12px;
        align-items: center;
        padding: 8px 10px;
        border-radius: 10px;
        background: #f8f8f8;
    }

    .rank {
        font-weight: 700;
    }

    .title {
        font-weight: 600;
    }

    .plays {
        opacity: 0.7;
        font-size: 13px;
    }

    .loading {
        display: flex;
        align-items: center;
        gap: 10px;
        justify-content: center;
        min-height: 120px;
    }

    .error {
        border-color: #ffd1d1;
        background: #fff5f5;
        display: grid;
        gap: 10px;
    }

    .error-title {
        font-weight: 700;
    }

    .error-msg {
        white-space: pre-wrap;
    }

    .error button {
        width: fit-content;
        padding: 6px 10px;
        border-radius: 10px;
        border: 1px solid #d7d7d7;
        background: #fff;
        cursor: pointer;
    }

    .empty {
        opacity: 0.6;
    }

    .updated {
        font-size: 12px;
        opacity: 0.6;
    }

    @media (max-width: 720px) {
        .top-list li {
            grid-template-columns: 1fr;
            align-items: flex-start;
        }

        .plays {
            margin-top: 4px;
        }
    }
</style>
