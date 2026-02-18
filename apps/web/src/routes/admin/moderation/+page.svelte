<script lang="ts">
    import { onMount } from "svelte";
    import { auth } from "$lib/stores/auth";
    import { ApiError, createApiClient } from "$lib/api/client";

    type FlaggedSubmission = {
        id: number;
        game_id: number;
        player_name: string;
        score: number;
        flag_reason?: string | null;
        session_id?: string | null;
        created_at: string;
    };

    const adminApi = createApiClient({
        getToken: () => {
            return (auth as any).getSnapshot?.().token ?? null;
        },
    });

    let loading = true;
    let errorMsg: string | null = null;
    let items: FlaggedSubmission[] = [];
    let removingId: number | null = null;

    function formatDate(s: string) {
        try {
            const d = new Date(s);
            if (Number.isNaN(d.getTime())) return s;
            return d.toLocaleString();
        } catch {
            return s;
        }
    }

    function displayPlayer(name: string) {
        const v = (name ?? "").trim();
        return v ? v : "Guest";
    }

    function displayFlag(reason?: string | null) {
        const v = (reason ?? "").trim();
        return v ? v : "-";
    }

    async function load() {
        loading = true;
        errorMsg = null;

        try {
            const res = await adminApi.get<{ items: FlaggedSubmission[] }>(
                "/admin/moderation/flagged-submissions?limit=50"
            );
            items = res?.items ?? [];
        } catch (e) {
            if (e instanceof ApiError) errorMsg = `${e.code}: ${e.message}`;
            else errorMsg = "Failed to load flagged submissions.";
        } finally {
            loading = false;
        }
    }

    async function removeScore(item: FlaggedSubmission) {
        if (removingId) return;
        if (!item) return;

        const ok = typeof window === "undefined" ? true : window.confirm(
            `Remove score ${item.score} for ${displayPlayer(item.player_name)}?`
        );
        if (!ok) return;

        removingId = item.id;
        errorMsg = null;
        try {
            await adminApi.post<{ ok: boolean }>("/admin/moderation/remove-score", {
                submission_id: String(item.id),
            });
            items = items.filter((it) => it.id !== item.id);
        } catch (e) {
            if (e instanceof ApiError) errorMsg = `${e.code}: ${e.message}`;
            else errorMsg = "Failed to remove score.";
        } finally {
            removingId = null;
        }
    }

    onMount(() => {
        void load();
    });
</script>

<svelte:head>
    <title>Admin Moderation</title>
</svelte:head>

<div class="head">
    <div>
        <h1>Moderation</h1>
        <p class="muted">Flagged leaderboard submissions</p>
    </div>
    <button class="btn" type="button" on:click={load} disabled={loading}>Refresh</button>
</div>

{#if loading}
    <div class="state">Loading flagged submissions…</div>
{:else if errorMsg}
    <div class="state error" role="alert">
        <div><b>Error</b></div>
        <div>{errorMsg}</div>
        <button class="btn sm" type="button" on:click={load}>Retry</button>
    </div>
{:else if items.length === 0}
    <div class="state empty">No flagged submissions</div>
{:else}
    <div class="tableWrap">
        <table class="table">
            <thead>
                <tr>
                    <th>Player</th>
                    <th>Game</th>
                    <th>Score</th>
                    <th>Flag reason</th>
                    <th>Date</th>
                    <th>Action</th>
                </tr>
            </thead>
            <tbody>
                {#each items as it (it.id)}
                    <tr>
                        <td>{displayPlayer(it.player_name)}</td>
                        <td>#{it.game_id}</td>
                        <td>{it.score}</td>
                        <td>{displayFlag(it.flag_reason)}</td>
                        <td>{formatDate(it.created_at)}</td>
                        <td>
                            <button
                                class="btn danger sm"
                                type="button"
                                on:click={() => removeScore(it)}
                                disabled={removingId === it.id}
                            >
                                {removingId === it.id ? "Removing…" : "Remove"}
                            </button>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
{/if}

<style>
    h1 {
        margin: 0;
        font-size: 22px;
    }

    .muted {
        margin: 4px 0 0;
        opacity: 0.7;
    }

    .head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        margin-bottom: 16px;
        flex-wrap: wrap;
    }

    .state {
        padding: 14px;
        border-radius: 12px;
        background: #f7f7f7;
    }

    .state.error {
        background: #fff3f3;
        border: 1px solid #ffd1d1;
        display: grid;
        gap: 8px;
    }

    .state.empty {
        color: #444;
    }

    .tableWrap {
        overflow-x: auto;
        border: 1px solid #eee;
        border-radius: 12px;
        background: #fff;
    }

    .table {
        width: 100%;
        border-collapse: collapse;
        font-size: 14px;
    }

    th, td {
        padding: 12px 10px;
        text-align: left;
        border-bottom: 1px solid #f0f0f0;
        vertical-align: middle;
    }

    th {
        font-size: 12px;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        opacity: 0.7;
        background: #fafafa;
    }

    tr:last-child td {
        border-bottom: none;
    }

    .btn {
        padding: 8px 12px;
        border-radius: 10px;
        border: 1px solid #ddd;
        background: #fff;
        cursor: pointer;
    }

    .btn.sm {
        padding: 6px 10px;
        font-size: 12px;
    }

    .btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .btn.danger {
        border-color: #f4c2c2;
        color: #b00020;
        background: #fff5f5;
    }
</style>
