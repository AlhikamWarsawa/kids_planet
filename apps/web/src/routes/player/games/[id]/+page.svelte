<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { browser } from "$app/environment";

    import { session } from "$lib/stores/session";
    import { ApiError } from "$lib/api/client";
    import { getGame } from "$lib/api/games";
    import type { GameDetail } from "$lib/types/game";

    import { getLeaderboard } from "$lib/api/leaderboard";
    import type { LeaderboardViewResponse, LeaderboardItem } from "$lib/api/leaderboard";

    let gameId: number | null = null;
    let game: GameDetail | null = null;

    let stage: "idle" | "loading_game" | "starting_session" | "ready" | "unavailable" | "error" = "idle";
    let errorMsg: string | null = null;

    let expiresAt: number | null = null;

    let lbStage: "idle" | "loading" | "ready" | "error" = "idle";
    let lbError: string | null = null;
    let lbItems: LeaderboardItem[] = [];
    let lbReqSeq = 0;

    let initialized = false;
    let lastRouteId: string | null = null;


    function parseGameIdFromParam(raw: any): number | null {
        const n = Number(raw);
        if (!Number.isFinite(n) || n <= 0) return null;
        return n;
    }

    function formatExp(ms: number | null) {
        if (!ms) return "-";
        try {
            return new Date(ms).toLocaleString();
        } catch {
            return "-";
        }
    }

    function base64UrlDecode(input: string): string {
        const pad = "=".repeat((4 - (input.length % 4)) % 4);
        const b64 = (input + pad).replace(/-/g, "+").replace(/_/g, "/");
        try {
            return atob(b64);
        } catch {
            return "";
        }
    }

    function getTokenGameId(token: string | null): number | null {
        if (!token) return null;
        const parts = token.split(".");
        if (parts.length < 2) return null;

        const payloadJson = base64UrlDecode(parts[1]);
        if (!payloadJson) return null;

        try {
            const payload = JSON.parse(payloadJson) as any;
            const gid = Number(payload?.game_id);
            if (!Number.isFinite(gid) || gid <= 0) return null;
            return gid;
        } catch {
            return null;
        }
    }

    function isExistingSessionUsableForGame(gid: number): boolean {
        const snap = session.getSnapshot();
        if (!snap.playToken || !snap.expiresAt) return false;
        if (Date.now() >= snap.expiresAt) return false;

        const tokenGid = getTokenGameId(snap.playToken);
        return tokenGid === gid;
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

    async function loadLeaderboardForPlayPage() {
        if (!gameId) return;

        const seq = ++lbReqSeq;
        lbStage = "loading";
        lbError = null;

        try {
            const res: LeaderboardViewResponse = await getLeaderboard(gameId, {
                period: "daily",
                scope: "game",
                limit: 10,
            });

            if (seq !== lbReqSeq) return;

            lbItems = res.items ?? [];
            lbStage = "ready";
        } catch (e) {
            if (seq !== lbReqSeq) return;

            lbStage = "error";
            lbError = e instanceof ApiError ? `${e.code}: ${e.message}` : "Failed to load leaderboard.";
        }
    }


    async function run() {
        errorMsg = null;
        game = null;
        gameId = null;
        expiresAt = null;

        lbStage = "idle";
        lbError = null;
        lbItems = [];

        const rid = ($page.params as any)?.id;
        const gid = parseGameIdFromParam(rid);

        if (!gid) {
            stage = "error";
            errorMsg = "Invalid game id.";
            return;
        }

        gameId = gid;

        stage = "loading_game";
        try {
            game = await getGame(gid);
        } catch (e) {
            stage = "error";
            if (e instanceof ApiError && e.status === 404) {
                errorMsg = "Game not found / not active.";
            } else {
                errorMsg = e instanceof ApiError ? e.message : "Failed to fetch game detail.";
            }
            return;
        }

        if (!game?.game_url) {
            stage = "unavailable";
            errorMsg = "Game not available yet.";
            return;
        }

        stage = "starting_session";
        try {
            if (!isExistingSessionUsableForGame(gid)) {
                const res = await session.startSession(gid);
                expiresAt = res.expiresAt;
            } else {
                const snap = session.getSnapshot();
                expiresAt = snap.expiresAt;
            }
        } catch (e) {
            stage = "error";
            errorMsg = e instanceof ApiError ? e.message : "Failed to start session.";
            return;
        }

        stage = "ready";

        await loadLeaderboardForPlayPage();
    }

    onMount(async () => {
        session.loadFromStorage();
        initialized = true;

        lastRouteId = String(($page.params as any)?.id ?? "");
        await run();
    });

    $: if (initialized && browser) {
        const rid = String(($page.params as any)?.id ?? "");
        if (rid !== lastRouteId) {
            lastRouteId = rid;
            run();
        }
    }
</script>

<svelte:head>
    <title>{game?.title ? `Play - ${game.title}` : "Play"}</title>
</svelte:head>

<main class="screen">
    <header class="topbar">
        <a class="pill" href="/player">← Back</a>

        {#if gameId}
            <a class="pill" href={`/player/leaderboard?game_id=${gameId}&period=daily&scope=game&limit=10`}>Leaderboard</a>
        {/if}

        <div class="title">
            <div class="h1">{game?.title ?? "Play"}</div>
            {#if expiresAt}
                <div class="sub">Session expires: {formatExp(expiresAt)}</div>
            {/if}
        </div>
    </header>

    {#if stage === "loading_game"}
        <div class="state">
            <div class="state-title">Loading game…</div>
            <div class="state-sub">Fetching game detail</div>
        </div>

    {:else if stage === "starting_session"}
        <div class="state">
            <div class="state-title">Starting session…</div>
            <div class="state-sub">Requesting play token</div>
        </div>

    {:else if stage === "error"}
        <div class="state error" role="alert">
            <div class="state-title">Error</div>
            <div class="state-sub">{errorMsg ?? "Something went wrong."}</div>
            <div class="rowBtns">
                <button class="pill mt" type="button" on:click={run}>Retry</button>
                <a class="pill mt" href="/player">Back to catalog</a>
            </div>
        </div>

    {:else if stage === "unavailable"}
        <div class="state" role="status" aria-live="polite">
            <div class="state-title">Game not available yet</div>
            <div class="state-sub">{errorMsg ?? "Game not available yet."}</div>
            <div class="rowBtns">
                <button class="pill mt" type="button" on:click={run}>Retry</button>
                <a class="pill mt" href="/player">Back to catalog</a>
            </div>
        </div>

    {:else if stage === "ready" && game?.game_url}
        <section class="lbPanel" aria-label="Leaderboard panel">
            <div class="lbHead">
                <div class="lbTitle">Leaderboard (daily · game)</div>
                <button class="pill sm" type="button" on:click={loadLeaderboardForPlayPage} disabled={lbStage === "loading"}>
                    {lbStage === "loading" ? "Loading…" : "Refresh"}
                </button>
            </div>

            {#if lbStage === "loading"}
                <div class="lbState">Loading leaderboard…</div>
            {:else if lbStage === "error"}
                <div class="lbError" role="alert">
                    <div><b>Error</b></div>
                    <div>{lbError ?? "Failed."}</div>
                    <button class="pill sm mt2" type="button" on:click={loadLeaderboardForPlayPage}>Retry</button>
                </div>
            {:else if lbStage === "ready"}
                {#if lbItems.length === 0}
                    <div class="lbEmpty">Belum ada skor</div>
                {:else}
                    <ol class="lbList">
                        {#each lbItems as it, idx (it.member)}
                            <li class="lbRow">
                                <div class="lbRank">#{idx + 1}</div>
                                <div class="lbMember" title={it.member}>{shortenMember(it.member)}</div>
                                <div class="lbScore">{it.score}</div>
                            </li>
                        {/each}
                    </ol>
                {/if}
            {:else}
                <div class="lbEmpty">Belum ada skor</div>
            {/if}
        </section>

        <section class="frameWrap">
            <iframe
                    class="frame"
                    title={game.title}
                    src={game.game_url}
                    loading="eager"
                    allowfullscreen
                    allow="fullscreen; gamepad; autoplay"
                    sandbox="allow-scripts allow-same-origin allow-forms allow-pointer-lock"
            />
        </section>
    {/if}
</main>

<style>
    .screen {
        min-height: 100vh;
        padding: 16px;
        font-family: system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif;
        background: #fff;
        box-sizing: border-box;
        overflow-x: hidden;
    }


    .topbar {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 12px;
        flex-wrap: wrap;
    }

    .title { min-width: 0; flex: 1 1 auto; }

    .h1 {
        font-weight: 800;
        font-size: 18px;
        color: #222;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
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
        font-weight: 700;
        cursor: pointer;
        text-decoration: none;
        color: #222;
        box-sizing: border-box;
        white-space: nowrap;
    }

    .pill:hover { background: #f5f5f5; }
    .pill:disabled { opacity: 0.6; cursor: not-allowed; }

    .pill.sm {
        padding: 8px 12px;
        border-width: 3px;
        font-weight: 900;
        font-size: 12px;
    }

    .state {
        border: 4px solid #666;
        border-radius: 14px;
        padding: 16px;
        background: #fff;
        box-sizing: border-box;
        max-width: 720px;
    }

    .state-title {
        font-weight: 800;
        margin-bottom: 6px;
        color: #222;
    }

    .state-sub {
        opacity: 0.75;
        color: #222;
        font-size: 13px;
    }

    .error { border-color: #ef4444; }
    .mt { margin-top: 12px; }

    .rowBtns { display: flex; gap: 10px; flex-wrap: wrap; }


    .lbPanel {
        border: 4px solid #666;
        border-radius: 14px;
        padding: 12px;
        background: #fff;
        box-sizing: border-box;
        margin-bottom: 12px;
        max-width: 920px;
    }

    .lbHead {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
        flex-wrap: wrap;
        margin-bottom: 10px;
    }

    .lbTitle {
        font-weight: 900;
        color: #222;
        font-size: 12px;
        letter-spacing: 0.3px;
        text-transform: uppercase;
        opacity: 0.9;
    }

    .lbState {
        font-weight: 900;
        font-size: 12px;
        color: #222;
        opacity: 0.8;
        padding: 6px 2px;
    }

    .lbEmpty {
        font-weight: 900;
        font-size: 12px;
        color: #222;
        opacity: 0.8;
        padding: 6px 2px;
    }

    .lbError {
        border: 3px solid #ef4444;
        border-radius: 12px;
        padding: 10px 12px;
        color: #991b1b;
        font-weight: 900;
        font-size: 12px;
        background: #fff;
    }

    .mt2 { margin-top: 8px; }

    .lbList {
        list-style: none;
        padding: 0;
        margin: 0;
        display: grid;
        gap: 8px;
    }

    .lbRow {
        display: grid;
        grid-template-columns: 70px 1fr 120px;
        gap: 10px;
        align-items: center;
        border: 3px solid #666;
        border-radius: 12px;
        padding: 10px 12px;
        box-sizing: border-box;
    }

    .lbRank { font-weight: 900; font-size: 12px; color: #222; opacity: 0.9; }
    .lbMember {
        font-weight: 900;
        font-size: 12px;
        color: #222;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    .lbScore { text-align: right; font-weight: 900; font-size: 12px; color: #222; }

    .frameWrap {
        width: 100%;
        border: 4px solid #666;
        border-radius: 14px;
        overflow: hidden;
        box-sizing: border-box;
        background: #fff;

        height: min(62vh, 720px);
        min-height: 420px;
    }

    .frame {
        width: 100%;
        height: 100%;
        border: 0;
        display: block;
    }

    @media (min-width: 720px) {
        .screen { padding: 20px; }
        .h1 { font-size: 20px; }
        .frameWrap { height: min(68vh, 760px); }
        .lbRow { grid-template-columns: 80px 1fr 140px; }
    }
</style>
