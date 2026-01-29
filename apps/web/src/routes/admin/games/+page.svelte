<script lang="ts">
    import { onMount } from "svelte";
    import { auth } from "$lib/stores/auth";
    import { ApiError, createApiClient } from "$lib/api/client";
    import type {
        AdminGameDTO,
        AdminGameListResponse,
        AdminGameStatus,
        AdminCreateGameRequest,
        AdminUpdateGameRequest,
    } from "$lib/api/games";

    const adminApi = createApiClient({
        getToken: () => {
            return (auth as any).getSnapshot?.().token ?? null;
        },
    });

    let loading = true;
    let errorMsg: string | null = null;

    let items: AdminGameDTO[] = [];
    let page = 1;
    let limit = 24;
    let total = 0;

    let status: AdminGameStatus | "" = "";
    let q = "";

    let busyRowId: number | null = null;

    let toast: { kind: "ok" | "err"; message: string } | null = null;
    let toastTimer: any = null;

    function showToast(kind: "ok" | "err", message: string) {
        toast = { kind, message };
        if (toastTimer) clearTimeout(toastTimer);
        toastTimer = setTimeout(() => (toast = null), 2200);
    }

    function formatDate(s: string) {
        try {
            const d = new Date(s);
            if (Number.isNaN(d.getTime())) return s;
            return d.toLocaleString();
        } catch {
            return s;
        }
    }

    function clampInt(n: number, min: number, max: number) {
        if (!Number.isFinite(n)) return min;
        return Math.max(min, Math.min(max, Math.trunc(n)));
    }

    type AgeCategoryDTO = {
        id: number;
        label: string;
        min_age: number;
        max_age: number;
    };

    let ageCatsLoading = true;
    let ageCatsError: string | null = null;
    let ageCats: AgeCategoryDTO[] = [];

    function pickId(obj: any): number | null {
        const v = obj?.id ?? obj?.ID ?? obj?.Id ?? obj?.iD;
        const n = Number(v);
        return Number.isFinite(n) && n >= 1 ? n : null;
    }

    function normalizeAgeCat(raw: any): AgeCategoryDTO | null {
        const id = pickId(raw);
        if (!id) return null;

        const label = String(raw?.label ?? raw?.Label ?? "");
        const min_age = clampInt(Number(raw?.min_age ?? raw?.MinAge ?? raw?.minAge ?? 0), 0, 999);
        const max_age = clampInt(Number(raw?.max_age ?? raw?.MaxAge ?? raw?.maxAge ?? 0), 0, 999);

        return { id, label, min_age, max_age };
    }

    function ageCatText(a: AgeCategoryDTO) {
        const max = a.max_age ?? 0;
        const min = a.min_age ?? 0;

        if (a.label?.trim()) return `${a.label} (min ${min}, max ${max})`;
        return `#${a.id} (min ${min}, max ${max})`;
    }

    async function loadAgeCats() {
        ageCatsLoading = true;
        ageCatsError = null;

        const qs = new URLSearchParams();
        qs.set("page", "1");
        qs.set("limit", "100");

        try {
            const data = await adminApi.get<{ items: any[] }>(`/admin/age-categories?${qs.toString()}`);
            const raw = Array.isArray(data?.items) ? data.items : [];
            const normalized = raw.map(normalizeAgeCat).filter(Boolean) as AgeCategoryDTO[];

            normalized.sort((a, b) => {
                if (a.min_age !== b.min_age) return a.min_age - b.min_age;
                return a.id - b.id;
            });

            ageCats = normalized;

            if ((!Number.isFinite(formAgeCategoryId) || formAgeCategoryId < 1) && ageCats.length > 0) {
                formAgeCategoryId = ageCats[0].id;
            }
        } catch (e) {
            if (e instanceof ApiError) ageCatsError = `${e.code}: ${e.message}`;
            else ageCatsError = String(e);
        } finally {
            ageCatsLoading = false;
        }
    }

    async function loadList(opts?: { keepPage?: boolean }) {
        loading = true;
        errorMsg = null;

        if (!opts?.keepPage) page = 1;

        const qs = new URLSearchParams();
        if (status === "draft" || status === "active" || status === "archived") qs.set("status", status);
        if (q.trim()) qs.set("q", q.trim());
        qs.set("page", String(page));
        qs.set("limit", String(limit));

        try {
            const data = await adminApi.get<AdminGameListResponse>(`/admin/games?${qs.toString()}`);
            items = data.items ?? [];
            page = data.page ?? page;
            limit = data.limit ?? limit;
            total = data.total ?? 0;
        } catch (e) {
            if (e instanceof ApiError) errorMsg = `${e.code}: ${e.message}`;
            else errorMsg = String(e);
        } finally {
            loading = false;
        }
    }

    function totalPages() {
        const denom = limit > 0 ? limit : 24;
        return Math.max(1, Math.ceil((total || 0) / denom));
    }

    async function goPrev() {
        if (loading) return;
        if (page <= 1) return;
        page -= 1;
        await loadList({ keepPage: true });
    }

    async function goNext() {
        if (loading) return;
        const tp = totalPages();
        if (page >= tp) return;
        page += 1;
        await loadList({ keepPage: true });
    }

    type Mode = "create" | "edit";
    let mode: Mode = "create";
    let editId: number | null = null;

    let formTitle = "";
    let formSlug = "";
    let formAgeCategoryId = 1;
    let formFree = true;

    let submitting = false;

    function resetForm() {
        mode = "create";
        editId = null;
        formTitle = "";
        formSlug = "";
        if (ageCats.length > 0) formAgeCategoryId = ageCats[0].id;
        else formAgeCategoryId = 1;
        formFree = true;
    }

    function startEdit(g: AdminGameDTO) {
        mode = "edit";
        editId = g.id;
        formTitle = g.title ?? "";
        formSlug = g.slug ?? "";
        formAgeCategoryId = g.age_category_id ?? (ageCats[0]?.id ?? 1);
        formFree = Boolean(g.free);
        showToast("ok", "Edit mode");
    }

    function validateForm() {
        const t = formTitle.trim();
        const s = formSlug.trim();
        const age = clampInt(Number(formAgeCategoryId), 1, 1_000_000_000);

        if (!t) return "title is required";
        if (!s) return "slug is required";
        if (age < 1) return "age_category_id must be >= 1";

        return null;
    }

    async function submitCreate() {
        const err = validateForm();
        if (err) {
            showToast("err", err);
            return;
        }

        submitting = true;
        try {
            const payload: AdminCreateGameRequest = {
                title: formTitle.trim(),
                slug: formSlug.trim(),
                age_category_id: clampInt(Number(formAgeCategoryId), 1, 1_000_000_000),
                free: Boolean(formFree),
            };

            await adminApi.post<AdminGameDTO>("/admin/games", payload);
            showToast("ok", "Created");
            resetForm();
            await loadList();
        } catch (e) {
            if (e instanceof ApiError) showToast("err", `${e.code}: ${e.message}`);
            else showToast("err", String(e));
        } finally {
            submitting = false;
        }
    }

    async function submitUpdate() {
        if (editId == null) {
            showToast("err", "no game selected");
            return;
        }

        const err = validateForm();
        if (err) {
            showToast("err", err);
            return;
        }

        submitting = true;
        try {
            const payload: AdminUpdateGameRequest = {
                title: formTitle.trim(),
                slug: formSlug.trim(),
                age_category_id: clampInt(Number(formAgeCategoryId), 1, 1_000_000_000),
                free: Boolean(formFree),
            };

            await adminApi.put<AdminGameDTO>(`/admin/games/${editId}`, payload);
            showToast("ok", "Saved");
            resetForm();
            await loadList({ keepPage: true });
        } catch (e) {
            if (e instanceof ApiError) showToast("err", `${e.code}: ${e.message}`);
            else showToast("err", String(e));
        } finally {
            submitting = false;
        }
    }

    async function doPublish(g: AdminGameDTO) {
        if (busyRowId) return;
        busyRowId = g.id;
        try {
            await adminApi.post<AdminGameDTO>(`/admin/games/${g.id}/publish`);
            showToast("ok", "Published");
            await loadList({ keepPage: true });
        } catch (e) {
            if (e instanceof ApiError) showToast("err", `${e.code}: ${e.message}`);
            else showToast("err", String(e));
        } finally {
            busyRowId = null;
        }
    }

    async function doUnpublish(g: AdminGameDTO) {
        if (busyRowId) return;
        busyRowId = g.id;
        try {
            await adminApi.post<AdminGameDTO>(`/admin/games/${g.id}/unpublish`);
            showToast("ok", "Unpublished");
            await loadList({ keepPage: true });
        } catch (e) {
            if (e instanceof ApiError) showToast("err", `${e.code}: ${e.message}`);
            else showToast("err", String(e));
        } finally {
            busyRowId = null;
        }
    }

    onMount(() => {
        void Promise.all([loadAgeCats(), loadList()]);
    });
</script>

<svelte:head>
    <title>Admin · Games</title>
</svelte:head>

<div style="display:flex; align-items:flex-start; justify-content:space-between; gap: 16px; flex-wrap: wrap;">
    <div>
        <h1 style="margin: 0 0 6px;">Games</h1>
    </div>

    {#if toast}
        <div
                style="
                padding: 10px 12px;
                border-radius: 12px;
                border: 1px solid {toast.kind === 'ok' ? '#bfe7c6' : '#ffd1d1'};
                background: {toast.kind === 'ok' ? '#f1fff3' : '#fff3f3'};
                font-size: 13px;
                max-width: 420px;
            "
        >
            <b style="text-transform:uppercase; font-size: 11px; letter-spacing:.4px;">
                {toast.kind === "ok" ? "OK" : "ERROR"}
            </b>
            <div style="margin-top: 4px; white-space: pre-wrap;">{toast.message}</div>
        </div>
    {/if}
</div>

<div style="margin-top: 16px; display:flex; gap: 10px; flex-wrap: wrap; align-items: end;">
    <div style="display:grid; gap: 6px;">
        <div style="font-size: 12px; opacity: .7;">Status</div>
        <select bind:value={status} style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background: #fff;">
            <option value="">All</option>
            <option value="draft">draft</option>
            <option value="active">active</option>
            <option value="archived">archived</option>
        </select>
    </div>

    <div style="flex:1; min-width: 220px; display:grid; gap: 6px;">
        <div style="font-size: 12px; opacity: .7;">Search (title/slug)</div>
        <input
                bind:value={q}
                placeholder="e.g. color-match"
                style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff; width: 100%;"
        />
    </div>

    <div style="display:grid; gap: 6px;">
        <div style="font-size: 12px; opacity: .7;">Limit</div>
        <input
                type="number"
                min="1"
                max="100"
                bind:value={limit}
                style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff; width: 110px;"
        />
    </div>

    <div style="display:flex; gap: 8px; align-items: center;">
        <button
                on:click={() => loadList()}
                disabled={loading}
                style="padding: 9px 12px; border-radius: 10px; border: 1px solid #ddd; background: #111; color: #fff;"
        >
            Apply
        </button>
        <button
                on:click={() => {
                status = "";
                q = "";
                limit = 24;
                page = 1;
                void loadList();
            }}
                disabled={loading}
                style="padding: 9px 12px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
        >
            Reset
        </button>
    </div>
</div>

<div style="margin-top: 18px; padding: 14px; border: 1px solid #eee; border-radius: 14px; background: #fff;">
    <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
        <div>
            <div style="font-weight: 800;">
                {mode === "create" ? "Create Game" : `Edit Game #${editId}`}
            </div>
        </div>

        <div style="display:flex; gap: 8px;">
            {#if mode === "edit"}
                <button
                        on:click={resetForm}
                        disabled={submitting}
                        style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
                >
                    Cancel edit
                </button>
            {/if}

            <button
                    on:click={mode === "create" ? submitCreate : submitUpdate}
                    disabled={submitting}
                    style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background: #111; color: #fff;"
            >
                {submitting ? "Saving…" : mode === "create" ? "Create" : "Save"}
            </button>
        </div>
    </div>

    <div style="margin-top: 12px; display:grid; grid-template-columns: repeat(12, 1fr); gap: 10px;">
        <div style="grid-column: span 6; display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7;">Title</div>
            <input
                    bind:value={formTitle}
                    placeholder="Color Match"
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
            />
        </div>

        <div style="grid-column: span 4; display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7;">Slug</div>
            <input
                    bind:value={formSlug}
                    placeholder="color-match"
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
            />
        </div>

        <div style="grid-column: span 2; display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7; display:flex; align-items:center; justify-content:space-between; gap: 8px;">
                <span>Age Category</span>
                {#if ageCatsLoading}
                    <span style="font-size: 11px; opacity:.6;">Loading…</span>
                {:else if ageCatsError}
                    <button
                            on:click={() => loadAgeCats()}
                            style="padding: 4px 8px; border-radius: 999px; border: 1px solid #ddd; background:#fff; font-size: 11px;"
                            type="button"
                    >
                        Retry
                    </button>
                {/if}
            </div>

            <select
                    bind:value={formAgeCategoryId}
                    disabled={ageCatsLoading || !!ageCatsError}
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
            >
                {#if !ageCatsLoading && !ageCatsError && ageCats.length === 0}
                    <option value={1}>No age categories (create first)</option>
                {:else}
                    {#each ageCats as a (a.id)}
                        <option value={a.id}>{ageCatText(a)}</option>
                    {/each}
                {/if}
            </select>

            {#if ageCatsError}
                <div style="margin-top: 6px; font-size: 11px; color:#b42318;">
                    {ageCatsError}
                </div>
            {/if}
        </div>

        <label style="grid-column: span 12; display:flex; gap: 10px; align-items:center; margin-top: 2px;">
            <input type="checkbox" bind:checked={formFree} />
            <span style="font-size: 13px;">Free</span>
        </label>
    </div>
</div>

<div style="margin-top: 18px;">
    {#if loading}
        <div style="opacity:.7; padding: 12px 0;">Loading games…</div>
    {:else if errorMsg}
        <div style="padding: 12px; border-radius: 12px; background: #fff3f3; border: 1px solid #ffd1d1;">
            <b>ERROR</b>
            <div style="margin-top: 6px; white-space: pre-wrap;">{errorMsg}</div>
            <div style="margin-top: 10px; display:flex; gap: 8px;">
                <button
                        on:click={() => loadList({ keepPage: true })}
                        style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                >
                    Retry
                </button>
            </div>
        </div>
    {:else}
        <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
            <div style="font-size: 13px; opacity:.75;">
                Total: <b>{total}</b> • Page <b>{page}</b> / <b>{totalPages()}</b>
            </div>
            <div style="display:flex; gap: 8px;">
                <button
                        on:click={goPrev}
                        disabled={page <= 1 || loading}
                        style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                >
                    Prev
                </button>
                <button
                        on:click={goNext}
                        disabled={page >= totalPages() || loading}
                        style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                >
                    Next
                </button>
            </div>
        </div>

        {#if items.length === 0}
            <div style="margin-top: 12px; padding: 12px; border-radius: 12px; border: 1px dashed #ddd; background:#fafafa;">
                No games found.
            </div>
        {:else}
            <div style="margin-top: 12px; overflow:auto; border: 1px solid #eee; border-radius: 14px; background:#fff;">
                <table style="width: 100%; border-collapse: collapse; min-width: 860px;">
                    <thead>
                    <tr style="text-align:left; background:#fafafa;">
                        <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Title</th>
                        <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Slug</th>
                        <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Status</th>
                        <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Age ID</th>
                        <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Updated</th>
                        <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Actions</th>
                    </tr>
                    </thead>
                    <tbody>
                    {#each items as g (g.id)}
                        <tr style="border-top: 1px solid #eee;">
                            <td style="padding: 10px 12px;">
                                <div style="font-weight: 700;">{g.title}</div>
                                <div style="font-size: 12px; opacity:.7;">
                                    #{g.id} • {g.free ? "free" : "paid"}
                                </div>
                            </td>

                            <td style="padding: 10px 12px; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13px;">
                                {g.slug}
                            </td>

                            <td style="padding: 10px 12px;">
                                    <span
                                            style="
                                            display:inline-block;
                                            padding: 4px 8px;
                                            border-radius: 999px;
                                            font-size: 12px;
                                            border: 1px solid #ddd;
                                            background: {g.status === 'active' ? '#f1fff3' : g.status === 'draft' ? '#fffbe6' : '#f3f3f3'};
                                        "
                                    >
                                        {g.status}
                                    </span>
                            </td>

                            <td style="padding: 10px 12px;">
                                {g.age_category_id}
                            </td>

                            <td style="padding: 10px 12px; font-size: 13px; opacity:.8;">
                                {formatDate(g.updated_at)}
                            </td>

                            <td style="padding: 10px 12px;">
                                <div style="display:flex; gap: 8px; flex-wrap: wrap;">
                                    <button
                                            on:click={() => startEdit(g)}
                                            style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                                    >
                                        Edit
                                    </button>

                                    {#if g.status === "active"}
                                        <button
                                                on:click={() => doUnpublish(g)}
                                                disabled={busyRowId === g.id}
                                                style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                                        >
                                            {busyRowId === g.id ? "Working…" : "Unpublish"}
                                        </button>
                                    {:else}
                                        <button
                                                on:click={() => doPublish(g)}
                                                disabled={busyRowId === g.id}
                                                style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#111; color:#fff;"
                                        >
                                            {busyRowId === g.id ? "Working…" : "Publish"}
                                        </button>
                                    {/if}
                                </div>
                            </td>
                        </tr>
                    {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    {/if}
</div>
