<script lang="ts">
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";

    import { ApiError } from "$lib/api/client";
    import { getToken as getPlayerToken } from "$lib/auth/playerAuth";
    import { session } from "$lib/stores/session";
    import {
        getLeaderboard,
        getLeaderboardSelf,
        type LeaderboardSelfResponse,
        type LeaderboardViewResponse
    } from "$lib/api/leaderboard";

    let gameId: number | null = null;

    let period: "daily" | "weekly" = "daily";
    let scope: "game" | "global" = "game";
    let limit = 10;

    let stage: "idle" | "loading" | "ready" | "error" = "idle";
    let errorMsg: string | null = null;
    let data: LeaderboardViewResponse | null = null;
    let selfData: LeaderboardSelfResponse | null = null;
    let selfError: string | null = null;
    let selfLoading = false;

    $: hasSelfRank = selfData?.rank != null && selfData?.score != null;

    function selfToken(): string | null {
        const playerToken = getPlayerToken();
        if (playerToken) return playerToken;

        const snap = session.getSnapshot();
        return snap.playToken?.trim() || null;
    }

    function emptySelf(gid: number): LeaderboardSelfResponse {
        return {
            game_id: gid,
            rank: null,
            score: null,
            period,
            scope,
        };
    }

    function parseGameIdFromQuery(): number | null {
        const sp = get(page).url.searchParams;
        const raw = (sp.get("game_id") ?? "").trim();
        const n = Number(raw);
        if (!Number.isFinite(n) || n <= 0) return null;
        return n;
    }

    function shortenMember(member: string): string {
        const m = (member ?? "").trim();
        if (!m) return "-";

        if (m.startsWith("g:")) {
            const raw = m.slice(2);
            const head = raw.slice(0, 8);
            return head ? `guest:${head}` : "guest";
        }
        if (m.startsWith("s:")) {
            const raw = m.slice(2);
            const head = raw.slice(0, 8);
            return head ? `session:${head}` : "session";
        }
        if (m.startsWith("p:")) {
            const raw = m.slice(2);
            const head = raw.slice(0, 8);
            return head ? `player:${head}` : "player";
        }

        if (m.length <= 16) return m;
        return `${m.slice(0, 10)}…${m.slice(-4)}`;
    }

    async function load() {
        errorMsg = null;
        data = null;
        selfError = null;
        selfData = null;

        if (!gameId) {
            stage = "error";
            errorMsg = "Missing or invalid game_id.";
            return;
        }

        stage = "loading";
        selfLoading = true;
        try {
            data = await getLeaderboard(gameId, { period, scope, limit });
            stage = "ready";

            const token = selfToken();
            if (!token) {
                selfData = emptySelf(gameId);
                return;
            }

            try {
                selfData = await getLeaderboardSelf(gameId, { period, scope }, token);
            } catch (e) {
                if (e instanceof ApiError && e.status === 401) {
                    selfData = emptySelf(gameId);
                } else {
                    selfData = emptySelf(gameId);
                    selfError = e instanceof ApiError ? `${e.code}: ${e.message}` : "Failed to load your rank.";
                }
            }
        } catch (e) {
            stage = "error";
            errorMsg = e instanceof ApiError ? `${e.code}: ${e.message}` : "Failed to load leaderboard.";
        } finally {
            selfLoading = false;
        }
    }

    function applyFromQueryDefaults() {
        const sp = get(page).url.searchParams;

        const qp = (sp.get("period") ?? "").trim().toLowerCase();
        if (qp === "daily" || qp === "weekly") period = qp as any;

        const qs = (sp.get("scope") ?? "").trim().toLowerCase();
        if (qs === "game" || qs === "global") scope = qs as any;

        const ql = (sp.get("limit") ?? "").trim();
        if (ql) {
            const n = Number(ql);
            if (Number.isFinite(n) && n >= 1 && n <= 100) limit = Math.floor(n);
        }
    }

    onMount(async () => {
        session.loadFromStorage();
        gameId = parseGameIdFromQuery();
        applyFromQueryDefaults();
        await load();
    });
</script>

<svelte:head>
    <title>Leaderboard</title>
</svelte:head>

<main class="screen">
    <header class="topbar">
        <a class="pill" href="/player">← Back</a>

        {#if gameId}
            <a class="pill" href={`/player/games/${gameId}`}>Play</a>
        {/if}

        <div class="title">
            <div class="h1">Leaderboard</div>
            <div class="sub">
                {#if gameId}
                    game_id: {gameId}
                {:else}
                    game_id: -
                {/if}
            </div>
        </div>
    </header>

    <section class="panel">
        <div class="controls">
            <label class="ctl">
                <span>Period</span>
                <select bind:value={period} on:change={load}>
                    <option value="daily">daily</option>
                    <option value="weekly">weekly</option>
                </select>
            </label>

            <label class="ctl">
                <span>Scope</span>
                <select bind:value={scope} on:change={load}>
                    <option value="game">game</option>
                    <option value="global">global</option>
                </select>
            </label>

            <label class="ctl">
                <span>Limit</span>
                <select bind:value={limit} on:change={load}>
                    <option value={10}>10</option>
                    <option value={25}>25</option>
                    <option value={50}>50</option>
                    <option value={100}>100</option>
                </select>
            </label>

            <button class="pill btn" type="button" on:click={load} disabled={stage === "loading"}>
                {stage === "loading" ? "Loading…" : "Refresh"}
            </button>
        </div>

        <div class="selfBox" aria-live="polite">
            {#if selfLoading}
                <div class="selfText">Loading your rank…</div>
            {:else if hasSelfRank}
                <div class="selfText">
                    Your rank: <b>#{selfData?.rank}</b> · Score: <b>{selfData?.score}</b>
                </div>
            {:else}
                <div class="selfText muted">You are not ranked yet</div>
            {/if}
            {#if selfError}
                <div class="selfErr">{selfError}</div>
            {/if}
        </div>

        {#if stage === "loading"}
            <div class="state">Loading leaderboard…</div>
        {:else if stage === "error"}
            <div class="state error" role="alert">
                <div class="errTitle">Error</div>
                <div class="errMsg">{errorMsg ?? "Failed."}</div>
                <button class="pill mt" type="button" on:click={load}>Retry</button>
            </div>
        {:else if stage === "ready"}
            {#if !data || data.items.length === 0}
                <div class="empty">Belum ada skor</div>
            {:else}
                <ol class="list">
                    {#each data.items as it, idx (it.member)}
                        <li class="row">
                            <div class="rank">#{idx + 1}</div>
                            <div class="member" title={it.member}>{shortenMember(it.member)}</div>
                            <div class="score">{it.score}</div>
                        </li>
                    {/each}
                </ol>
            {/if}
        {:else}
            <div class="empty">Belum ada skor</div>
        {/if}
    </section>
</main>

<style>
    .screen {
        min-height: 100vh;
        padding: 16px;
        font-family: system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif;
        background: #fff;
        box-sizing: border-box;
    }

    .topbar {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 12px;
        flex-wrap: wrap;
    }

    .title {
        min-width: 0;
        flex: 1 1 auto;
    }

    .h1 {
        font-weight: 900;
        font-size: 18px;
        color: #222;
    }

    .sub {
        font-size: 12px;
        opacity: 0.7;
        color: #222;
        margin-top: 2px;
    }

    .pill {
        padding: 12px 18px;
        border-radius: 999px;
        border: 4px solid #666;
        background: #fff;
        font-weight: 800;
        cursor: pointer;
        text-decoration: none;
        color: #222;
        box-sizing: border-box;
        white-space: nowrap;
    }

    .pill:hover {
        background: #f5f5f5;
    }

    .pill:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .panel {
        border: 4px solid #666;
        border-radius: 14px;
        padding: 12px;
        background: #fff;
        box-sizing: border-box;
        max-width: 820px;
    }

    .controls {
        display: flex;
        gap: 10px;
        align-items: end;
        flex-wrap: wrap;
        margin-bottom: 10px;
    }

    .selfBox {
        border: 3px solid #666;
        border-radius: 12px;
        padding: 10px 12px;
        margin-bottom: 10px;
        background: #fff;
    }

    .selfText {
        font-size: 13px;
        font-weight: 900;
        color: #222;
    }

    .selfText.muted {
        opacity: 0.8;
    }

    .selfErr {
        margin-top: 6px;
        font-size: 12px;
        font-weight: 800;
        color: #991b1b;
    }

    .ctl {
        display: grid;
        gap: 6px;
        font-size: 12px;
        font-weight: 900;
        color: #222;
    }

    select {
        padding: 10px 12px;
        border: 3px solid #666;
        border-radius: 12px;
        font-weight: 800;
        background: #fff;
        outline: none;
        min-width: 120px;
    }

    .btn {
        padding: 10px 14px;
        border-width: 3px;
    }

    .state {
        padding: 10px 2px;
        font-size: 13px;
        font-weight: 800;
        color: #222;
        opacity: 0.85;
    }

    .error {
        border: 3px solid #ef4444;
        border-radius: 12px;
        padding: 10px 12px;
        opacity: 1;
        color: #991b1b;
        background: #fff;
    }

    .errTitle {
        font-weight: 900;
        margin-bottom: 6px;
    }

    .errMsg {
        font-weight: 800;
        font-size: 12px;
    }

    .mt {
        margin-top: 12px;
    }

    .empty {
        padding: 10px 2px;
        font-size: 13px;
        font-weight: 900;
        color: #222;
        opacity: 0.8;
    }

    .list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: grid;
        gap: 8px;
    }

    .row {
        display: grid;
        grid-template-columns: 72px 1fr 120px;
        gap: 10px;
        align-items: center;
        border: 3px solid #666;
        border-radius: 12px;
        padding: 10px 12px;
        box-sizing: border-box;
    }

    .rank {
        font-weight: 900;
        color: #222;
        font-size: 12px;
        opacity: 0.9;
    }

    .member {
        font-weight: 900;
        color: #222;
        font-size: 12px;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .score {
        text-align: right;
        font-weight: 900;
        color: #222;
        font-size: 12px;
    }

    @media (min-width: 720px) {
        .screen {
            padding: 20px;
        }
        .h1 {
            font-size: 20px;
        }
        .row {
            grid-template-columns: 80px 1fr 140px;
        }
    }
</style>
