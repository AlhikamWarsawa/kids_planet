<script lang="ts">
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";

    import { session } from "$lib/stores/session";
    import { ApiError } from "$lib/api/client";
    import { getGame } from "$lib/api/games";
    import type { GameDetail } from "$lib/types/game";

    let gameId: number | null = null;
    let game: GameDetail | null = null;

    let stage: "idle" | "loading_game" | "starting_session" | "ready" | "error" = "idle";
    let errorMsg: string | null = null;

    let expiresAt: number | null = null;

    function parseGameId(): number | null {
        const idStr = (get(page).params as any)?.id;
        const n = Number(idStr);
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

    async function run() {
        errorMsg = null;
        game = null;
        gameId = null;
        expiresAt = null;

        const gid = parseGameId();
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
            errorMsg = e instanceof ApiError ? e.message : "Failed to fetch game detail.";
            return;
        }

        if (!game?.game_url) {
            stage = "error";
            errorMsg = "Game is not playable yet (missing game_url).";
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
    }

    onMount(async () => {
        session.loadFromStorage();
        await run();
    });
</script>

<svelte:head>
    <title>{game?.title ? `Play - ${game.title}` : "Play"}</title>
</svelte:head>

<main class="screen">
    <header class="topbar">
        <a class="pill" href="/player">← Back</a>
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
            <button class="pill mt" type="button" on:click={run}>Retry</button>
        </div>
    {:else if stage === "ready" && game?.game_url}
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
        overflow: hidden;
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

    .pill:hover {
        background: #f5f5f5;
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

    .error {
        border-color: #ef4444;
    }

    .mt {
        margin-top: 12px;
    }

    .frameWrap {
        width: 100%;
        height: calc(100vh - 110px);
        border: 4px solid #666;
        border-radius: 14px;
        overflow: hidden;
        box-sizing: border-box;
        background: #fff;
    }

    .frame {
        width: 100%;
        height: 100%;
        border: 0;
        display: block;
    }

    @media (min-width: 720px) {
        .screen {
            padding: 20px;
        }
        .h1 {
            font-size: 20px;
        }
        .frameWrap {
            height: calc(100vh - 120px);
        }
    }
</style>