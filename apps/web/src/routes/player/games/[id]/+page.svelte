<script lang="ts">
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";
    import { browser } from "$app/environment";

    import { session } from "$lib/stores/session";
    import { api, ApiError } from "$lib/api/client";
    import { getGame } from "$lib/api/games";
    import type { GameDetail } from "$lib/types/game";

    let gameId: number | null = null;
    let game: GameDetail | null = null;

    let stage: "idle" | "loading_game" | "starting_session" | "ready" | "error" = "idle";
    let errorMsg: string | null = null;

    let expiresAt: number | null = null;

    const GUEST_KEY = "kidsplanet_guest_id";
    let guestId: string = "";
    let devScore = "";
    let submitLoading = false;
    let submitError: string | null = null;
    let submitResult: { accepted: boolean; best_score: number } | null = null;

    let toastMsg: string | null = null;
    let toastTimer: any = null;

    function showToast(msg: string) {
        toastMsg = msg;
        if (toastTimer) clearTimeout(toastTimer);
        toastTimer = setTimeout(() => {
            toastMsg = null;
            toastTimer = null;
        }, 2200);
    }

    function getOrCreateGuestId(): string {
        if (!browser) return "";
        try {
            const existing = (localStorage.getItem(GUEST_KEY) ?? "").trim();
            if (existing) return existing;

            let id = "";
            const c: any = globalThis as any;
            if (c?.crypto?.randomUUID) {
                id = c.crypto.randomUUID();
            } else {
                id = `anon_${Date.now()}_${Math.random().toString(16).slice(2)}`;
            }
            localStorage.setItem(GUEST_KEY, id);
            return id;
        } catch {
            return "";
        }
    }

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

    function randomScore(): number {
        return Math.floor(Math.random() * 1000);
    }

    function parseScoreInput(raw: string): { ok: true; value: number } | { ok: false; message: string } {
        const s = raw.trim();
        if (!s) return { ok: true, value: randomScore() };

        const n = Number(s);
        if (!Number.isFinite(n)) return { ok: false, message: "Score must be a number (integer)." };

        const i = Math.floor(n);
        if (i !== n) return { ok: false, message: "Score must be an integer." };
        if (i < 0) return { ok: false, message: "Score must be >= 0." };

        return { ok: true, value: i };
    }

    async function submitScoreDev() {
        submitError = null;
        submitResult = null;

        if (stage !== "ready" || !gameId) {
            submitError = "Play page not ready yet.";
            return;
        }

        const snap = session.getSnapshot();
        const token = (snap.playToken ?? "").trim();
        if (!token) {
            submitError = "Missing play token. Try Retry.";
            return;
        }

        if (!guestId) {
            submitError = "Missing guest id. Refresh page.";
            return;
        }

        const parsed = parseScoreInput(devScore);
        if (!parsed.ok) {
            submitError = parsed.message;
            return;
        }

        submitLoading = true;
        try {
            const res = await api.post<{ accepted: boolean; best_score: number }>(
                "/leaderboard/submit",
                { game_id: gameId, score: parsed.value },
                {
                    token,
                    headers: {
                        "X-Guest-Id": guestId,
                    },
                }
            );

            submitResult = res;
            showToast(`Submitted best_score=${res.best_score}`);
        } catch (e) {
            const msg = e instanceof ApiError ? `${e.code}: ${e.message}` : "Submit failed.";
            submitError = msg;
            showToast("Submit failed");
        } finally {
            submitLoading = false;
        }
    }

    async function run() {
        errorMsg = null;
        game = null;
        gameId = null;
        expiresAt = null;

        submitError = null;
        submitResult = null;

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
        guestId = getOrCreateGuestId();
        await run();
    });
</script>

<svelte:head>
    <title>{game?.title ? `Play - ${game.title}` : "Play"}</title>
</svelte:head>

<main class="screen">
    {#if toastMsg}
        <div class="toast" role="status" aria-live="polite">{toastMsg}</div>
    {/if}

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
        <section class="devPanel" aria-label="Dev tools">
            <div class="devTitle">Dev tools</div>
            <div class="devRow">
                <label class="devLabel" for="score">Score</label>
                <input
                        id="score"
                        class="devInput"
                        type="number"
                        inputmode="numeric"
                        placeholder="(empty = random)"
                        bind:value={devScore}
                        min="0"
                />
                <button class="pill devBtn" type="button" on:click={submitScoreDev} disabled={submitLoading}>
                    {submitLoading ? "Submitting…" : "Submit Score (dev)"}
                </button>
            </div>

            <div class="devMeta">
                <div><b>guest_id</b>: {guestId || "-"}</div>
                <div><b>token</b>: {session.getSnapshot().playToken ? "ready" : "missing"}</div>
            </div>

            {#if submitError}
                <div class="devError" role="alert">{submitError}</div>
            {/if}
            {#if submitResult}
                <div class="devOk">
                    accepted: <b>{submitResult.accepted ? "true" : "false"}</b> · best_score: <b>{submitResult.best_score}</b>
                </div>
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
        overflow: hidden;
        box-sizing: border-box;
    }

    .toast {
        position: fixed;
        top: 14px;
        right: 14px;
        z-index: 50;
        padding: 10px 12px;
        border-radius: 12px;
        border: 3px solid #666;
        background: #fff;
        font-weight: 800;
        font-size: 12px;
        color: #222;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
        max-width: min(420px, calc(100vw - 28px));
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
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

    .pill:disabled {
        opacity: 0.6;
        cursor: not-allowed;
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

    .devPanel {
        border: 4px dashed #666;
        border-radius: 14px;
        padding: 12px;
        background: #fff;
        box-sizing: border-box;
        margin-bottom: 12px;
    }

    .devTitle {
        font-weight: 900;
        margin-bottom: 10px;
        color: #222;
        font-size: 12px;
        letter-spacing: 0.3px;
        text-transform: uppercase;
        opacity: 0.85;
    }

    .devRow {
        display: flex;
        gap: 10px;
        align-items: center;
        flex-wrap: wrap;
    }

    .devLabel {
        font-weight: 800;
        font-size: 12px;
        color: #222;
        opacity: 0.85;
    }

    .devInput {
        width: 160px;
        padding: 10px 12px;
        border: 3px solid #666;
        border-radius: 12px;
        font-weight: 800;
        outline: none;
    }

    .devBtn {
        padding: 10px 14px;
        border-width: 3px;
    }

    .devMeta {
        margin-top: 8px;
        font-size: 12px;
        color: #222;
        opacity: 0.85;
        display: flex;
        gap: 14px;
        flex-wrap: wrap;
    }

    .devError {
        margin-top: 10px;
        padding: 10px 12px;
        border-radius: 12px;
        border: 3px solid #ef4444;
        color: #991b1b;
        font-weight: 800;
        font-size: 12px;
    }

    .devOk {
        margin-top: 10px;
        padding: 10px 12px;
        border-radius: 12px;
        border: 3px solid #22c55e;
        color: #14532d;
        font-weight: 800;
        font-size: 12px;
    }

    .frameWrap {
        width: 100%;
        height: calc(100vh - 190px);
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
            height: calc(100vh - 200px);
        }
    }
</style>