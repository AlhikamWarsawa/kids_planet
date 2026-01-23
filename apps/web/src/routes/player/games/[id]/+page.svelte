<script lang="ts">
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";

    import { session } from "$lib/stores/session";
    import { ApiError } from "$lib/api/client";

    let gameId: number | null = null;

    let starting = true;
    let errorMsg: string | null = null;

    let expiresAt: number | null = null;

    function parseGameId(): number | null {
        const idStr = (get(page).params as any)?.id;
        const n = Number(idStr);
        if (!Number.isFinite(n) || n <= 0) return null;
        return n;
    }

    async function start() {
        errorMsg = null;
        starting = true;

        const gid = parseGameId();
        if (!gid) {
            starting = false;
            errorMsg = "Invalid game id.";
            return;
        }

        gameId = gid;

        try {
            const res = await session.startSession(gid);
            expiresAt = res.expiresAt;
        } catch (e) {
            if (e instanceof ApiError) {
                errorMsg = e.message;
            } else {
                errorMsg = "Failed to start session.";
            }
        } finally {
            starting = false;
        }
    }

    onMount(async () => {
        session.loadFromStorage();

        await start();
    });

    function formatExp(ms: number | null) {
        if (!ms) return "-";
        try {
            return new Date(ms).toLocaleString();
        } catch {
            return "-";
        }
    }
</script>

<section class="p-4">
    <h1 class="text-xl font-semibold">Play</h1>

    {#if starting}
        <div class="mt-4">
            <p>Starting session…</p>
        </div>
    {:else if errorMsg}
        <div class="mt-4">
            <p class="text-red-600">Error: {errorMsg}</p>
            <button class="mt-3 px-3 py-2 border rounded" on:click={start}>
                Retry
            </button>
        </div>
    {:else}
        <div class="mt-4">
            <p>Session ready for game ID: {gameId}</p>
            <p class="text-sm opacity-80">Expires at: {formatExp(expiresAt)}</p>

            <div class="mt-4 p-3 border rounded">
                <p class="text-sm">
                    Day 10 note: iframe belum dirender. Day 11 akan load game iframe setelah play_token siap.
                </p>
            </div>
        </div>
    {/if}
</section>
