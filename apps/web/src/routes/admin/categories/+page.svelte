<script lang="ts">
    import { onMount } from "svelte";
    import { auth } from "$lib/stores/auth";
    import { createApiClient } from "$lib/api/client";
    import { formatMappedError, mapApiError } from "$lib/api/errorMapper";

    type Tab = "age" | "education";
    let tab: Tab = "age";

    const adminApi = createApiClient({
        getToken: () => {
            return (auth as any).getSnapshot?.().token ?? null;
        },
    });

    type AgeCategoryDTO = {
        id: number;
        label: string;
        min_age: number;
        max_age: number;
        created_at?: string;
    };

    type EducationCategoryDTO = {
        id: number;
        name: string;
        created_at?: string;
    };

    type ListResp<T> = { items: T[]; page?: number; limit?: number };

    let toast: { kind: "ok" | "err"; message: string } | null = null;
    let toastTimer: any = null;
    function showToast(kind: "ok" | "err", message: string) {
        toast = { kind, message };
        if (toastTimer) clearTimeout(toastTimer);
        toastTimer = setTimeout(() => (toast = null), 2200);
    }

    function toErrorText(error: unknown, fallback = "Request failed") {
        return formatMappedError(mapApiError(error, fallback), {
            includeCode: false,
            includeRequestId: true,
        });
    }

    function clampInt(n: number, min: number, max: number) {
        if (!Number.isFinite(n)) return min;
        return Math.max(min, Math.min(max, Math.trunc(n)));
    }

    function pickId(obj: any): number | null {
        const v = obj?.id ?? obj?.ID ?? obj?.Id ?? obj?.iD;
        const n = Number(v);
        return Number.isFinite(n) && n >= 1 ? n : null;
    }

    function makeKey(prefix: string, id: number | null, fallback: string, idx: number) {
        if (id != null) return `${prefix}:${id}`;
        return `${prefix}:fallback:${fallback}:${idx}`;
    }

    let ageLoading = true;
    let ageError: string | null = null;
    let ageItems: AgeCategoryDTO[] = [];
    let ageQ = "";
    let agePage = 1;
    let ageLimit = 24;

    let ageMode: "create" | "edit" = "create";
    let ageEditId: number | null = null;

    let ageLabel = "";
    let ageMin = 0;
    let ageMax = 0;
    let ageSubmitting = false;
    let ageBusyRowId: number | null = null;

    function normalizeAgeItem(raw: any): AgeCategoryDTO {
        const id = pickId(raw) ?? 0;
        return {
            id,
            label: String(raw?.label ?? raw?.Label ?? ""),
            min_age: clampInt(Number(raw?.min_age ?? raw?.MinAge ?? raw?.minAge ?? 0), 0, 999),
            max_age: clampInt(Number(raw?.max_age ?? raw?.MaxAge ?? raw?.maxAge ?? 0), 0, 999),
            created_at: raw?.created_at ?? raw?.CreatedAt,
        };
    }

    async function loadAge(opts?: { keepPage?: boolean }) {
        ageLoading = true;
        ageError = null;
        if (!opts?.keepPage) agePage = 1;

        const qs = new URLSearchParams();
        if (ageQ.trim()) qs.set("q", ageQ.trim());
        qs.set("page", String(agePage));
        qs.set("limit", String(ageLimit));

        try {
            const data = await adminApi.get<ListResp<any>>(`/admin/age-categories?${qs.toString()}`);
            const rawItems = Array.isArray(data?.items) ? data.items : [];
            const normalized = rawItems.map(normalizeAgeItem);

            ageItems = normalized.filter((x) => x.id >= 1 || x.label.trim() !== "");

            if (typeof data?.page === "number") agePage = clampInt(data.page, 1, 1_000_000);
            if (typeof data?.limit === "number") ageLimit = clampInt(data.limit, 1, 100);
        } catch (e) {
            ageError = toErrorText(e, "Failed to load age categories");
        } finally {
            ageLoading = false;
        }
    }

    function resetAgeForm() {
        ageMode = "create";
        ageEditId = null;
        ageLabel = "";
        ageMin = 0;
        ageMax = 0;
    }

    function startEditAge(it: AgeCategoryDTO) {
        const id = Number(it?.id);
        if (!Number.isFinite(id) || id < 1) {
            showToast("err", "Cannot edit: invalid id from API (missing json tag?)");
            return;
        }
        ageMode = "edit";
        ageEditId = id;
        ageLabel = it.label ?? "";
        ageMin = it.min_age ?? 0;
        ageMax = it.max_age ?? 0;
        showToast("ok", "Edit age category");
    }

    function validateAgeForm() {
        const label = ageLabel.trim();
        const minAge = clampInt(Number(ageMin), 0, 999);
        const maxAge = clampInt(Number(ageMax), 0, 999);

        if (!label) return "label is required";
        if (label.length > 60) return "label max length is 60";
        if (minAge < 0) return "min_age must be >= 0";
        if (maxAge < 0) return "max_age must be >= 0";
        if (maxAge < minAge) return "max_age must be >= min_age";

        return null;
    }

    async function submitAge() {
        const err = validateAgeForm();
        if (err) return showToast("err", err);

        ageSubmitting = true;
        try {
            const payload = {
                label: ageLabel.trim(),
                min_age: clampInt(Number(ageMin), 0, 999),
                max_age: clampInt(Number(ageMax), 0, 999),
            };

            if (ageMode === "create") {
                await adminApi.post(`/admin/age-categories`, payload);
                showToast("ok", "Created");
                resetAgeForm();
                await loadAge();
            } else {
                if (!ageEditId || ageEditId < 1) throw new Error("no age category selected");
                await adminApi.put(`/admin/age-categories/${ageEditId}`, payload);
                showToast("ok", "Saved");
                resetAgeForm();
                await loadAge({ keepPage: true });
            }
        } catch (e) {
            showToast("err", toErrorText(e));
        } finally {
            ageSubmitting = false;
        }
    }

    async function deleteAge(it: AgeCategoryDTO) {
        const id = Number(it?.id);
        if (!Number.isFinite(id) || id < 1) {
            showToast("err", "Cannot delete: invalid id from API (missing json tag?)");
            return;
        }

        if (ageBusyRowId) return;
        const ok = confirm(`Delete age category #${id} (${it.label})?`);
        if (!ok) return;

        ageBusyRowId = id;
        try {
            await adminApi.del(`/admin/age-categories/${id}`);
            showToast("ok", "Deleted");
            if (ageMode === "edit" && ageEditId === id) resetAgeForm();

            if (ageItems.length <= 1 && agePage > 1) {
                agePage -= 1;
                await loadAge({ keepPage: true });
            } else {
                await loadAge({ keepPage: true });
            }
        } catch (e) {
            showToast("err", toErrorText(e));
        } finally {
            ageBusyRowId = null;
        }
    }

    async function agePrev() {
        if (ageLoading) return;
        if (agePage <= 1) return;
        agePage -= 1;
        await loadAge({ keepPage: true });
    }

    async function ageNext() {
        if (ageLoading) return;
        if (ageItems.length === 0) return;
        agePage += 1;
        await loadAge({ keepPage: true });
    }

    let eduLoading = true;
    let eduError: string | null = null;
    let eduItems: EducationCategoryDTO[] = [];
    let eduQ = "";
    let eduPage = 1;
    let eduLimit = 24;

    let eduMode: "create" | "edit" = "create";
    let eduEditId: number | null = null;

    let eduName = "";
    let eduSubmitting = false;
    let eduBusyRowId: number | null = null;

    function normalizeEduItem(raw: any): EducationCategoryDTO {
        const id = pickId(raw) ?? 0;
        return {
            id,
            name: String(raw?.name ?? raw?.Name ?? ""),
            created_at: raw?.created_at ?? raw?.CreatedAt,
        };
    }

    async function loadEdu(opts?: { keepPage?: boolean }) {
        eduLoading = true;
        eduError = null;
        if (!opts?.keepPage) eduPage = 1;

        const qs = new URLSearchParams();
        if (eduQ.trim()) qs.set("q", eduQ.trim());
        qs.set("page", String(eduPage));
        qs.set("limit", String(eduLimit));

        try {
            const data = await adminApi.get<ListResp<any>>(
                `/admin/education-categories?${qs.toString()}`
            );
            const rawItems = Array.isArray(data?.items) ? data.items : [];
            const normalized = rawItems.map(normalizeEduItem);

            eduItems = normalized.filter((x) => x.id >= 1 || x.name.trim() !== "");

            if (typeof data?.page === "number") eduPage = clampInt(data.page, 1, 1_000_000);
            if (typeof data?.limit === "number") eduLimit = clampInt(data.limit, 1, 100);
        } catch (e) {
            eduError = toErrorText(e, "Failed to load education categories");
        } finally {
            eduLoading = false;
        }
    }

    function resetEduForm() {
        eduMode = "create";
        eduEditId = null;
        eduName = "";
    }

    function startEditEdu(it: EducationCategoryDTO) {
        const id = Number(it?.id);
        if (!Number.isFinite(id) || id < 1) {
            showToast("err", "Cannot edit: invalid id from API (missing json tag?)");
            return;
        }
        eduMode = "edit";
        eduEditId = id;
        eduName = it.name ?? "";
        showToast("ok", "Edit education category");
    }

    function validateEduForm() {
        const name = eduName.trim();

        if (!name) return "name is required";
        if (name.length > 100) return "name max length is 100";

        return null;
    }

    async function submitEdu() {
        const err = validateEduForm();
        if (err) return showToast("err", err);

        eduSubmitting = true;
        try {
            const payload = {
                name: eduName.trim(),
            };

            if (eduMode === "create") {
                await adminApi.post(`/admin/education-categories`, payload);
                showToast("ok", "Created");
                resetEduForm();
                await loadEdu();
            } else {
                if (!eduEditId || eduEditId < 1) throw new Error("no education category selected");
                await adminApi.put(`/admin/education-categories/${eduEditId}`, payload);
                showToast("ok", "Saved");
                resetEduForm();
                await loadEdu({ keepPage: true });
            }
        } catch (e) {
            showToast("err", toErrorText(e));
        } finally {
            eduSubmitting = false;
        }
    }

    async function deleteEdu(it: EducationCategoryDTO) {
        const id = Number(it?.id);
        if (!Number.isFinite(id) || id < 1) {
            showToast("err", "Cannot delete: invalid id from API (missing json tag?)");
            return;
        }

        if (eduBusyRowId) return;
        const ok = confirm(`Delete education category #${id} (${it.name})?`);
        if (!ok) return;

        eduBusyRowId = id;
        try {
            await adminApi.del(`/admin/education-categories/${id}`);
            showToast("ok", "Deleted");
            if (eduMode === "edit" && eduEditId === id) resetEduForm();

            if (eduItems.length <= 1 && eduPage > 1) {
                eduPage -= 1;
                await loadEdu({ keepPage: true });
            } else {
                await loadEdu({ keepPage: true });
            }
        } catch (e) {
            showToast("err", toErrorText(e));
        } finally {
            eduBusyRowId = null;
        }
    }

    async function eduPrev() {
        if (eduLoading) return;
        if (eduPage <= 1) return;
        eduPage -= 1;
        await loadEdu({ keepPage: true });
    }

    async function eduNext() {
        if (eduLoading) return;
        if (eduItems.length === 0) return;
        eduPage += 1;
        await loadEdu({ keepPage: true });
    }

    onMount(async () => {
        await Promise.all([loadAge(), loadEdu()]);
    });
</script>

<svelte:head>
    <title>Admin · Categories</title>
</svelte:head>

<div style="display:flex; align-items:flex-start; justify-content:space-between; gap: 16px; flex-wrap: wrap;">
    <div>
        <h1 style="margin: 0 0 6px;">Categories</h1>
        <div style="font-size: 13px; opacity:.7;">Manage Age Categories & Education Categories</div>
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

<div style="margin-top: 14px; display:flex; gap: 8px; flex-wrap: wrap;">
    <button
            on:click={() => (tab = "age")}
            style="padding: 8px 12px; border-radius: 999px; border: 1px solid #ddd; background: {tab==='age' ? '#111' : '#fff'}; color: {tab==='age' ? '#fff' : '#111'};"
    >
        Age Categories
    </button>
    <button
            on:click={() => (tab = "education")}
            style="padding: 8px 12px; border-radius: 999px; border: 1px solid #ddd; background: {tab==='education' ? '#111' : '#fff'}; color: {tab==='education' ? '#fff' : '#111'};"
    >
        Education Categories
    </button>
</div>

{#if tab === "age"}
    <div style="margin-top: 16px; display:flex; gap: 10px; flex-wrap: wrap; align-items:end;">
        <div style="flex:1; min-width: 220px; display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7;">Search</div>
            <input
                    bind:value={ageQ}
                    placeholder="e.g. 7+"
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff; width: 100%;"
            />
        </div>

        <div style="display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7;">Limit</div>
            <input
                    type="number"
                    min="1"
                    max="100"
                    bind:value={ageLimit}
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff; width: 110px;"
            />
        </div>

        <div style="display:flex; gap: 8px; align-items: center;">
            <button
                    on:click={() => loadAge()}
                    disabled={ageLoading}
                    style="padding: 9px 12px; border-radius: 10px; border: 1px solid #ddd; background: #111; color: #fff;"
            >
                Apply
            </button>
            <button
                    on:click={() => {
                    ageQ = "";
                    ageLimit = 24;
                    agePage = 1;
                    void loadAge();
                }}
                    disabled={ageLoading}
                    style="padding: 9px 12px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
            >
                Reset
            </button>
        </div>
    </div>

    <div style="margin-top: 18px; padding: 14px; border: 1px solid #eee; border-radius: 14px; background: #fff;">
        <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
            <div style="font-weight: 800;">
                {ageMode === "create" ? "Create Age Category" : `Edit Age Category #${ageEditId}`}
            </div>

            <div style="display:flex; gap: 8px;">
                {#if ageMode === "edit"}
                    <button
                            on:click={resetAgeForm}
                            disabled={ageSubmitting}
                            style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
                    >
                        Cancel edit
                    </button>
                {/if}
                <button
                        on:click={submitAge}
                        disabled={ageSubmitting}
                        style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background: #111; color: #fff;"
                >
                    {ageSubmitting ? "Saving…" : ageMode === "create" ? "Create" : "Save"}
                </button>
            </div>
        </div>

        <div style="margin-top: 12px; display:grid; grid-template-columns: repeat(12, 1fr); gap: 10px;">
            <div style="grid-column: span 6; display:grid; gap: 6px;">
                <div style="font-size: 12px; opacity: .7;">Label</div>
                <input
                        bind:value={ageLabel}
                        placeholder="7+"
                        style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                />
            </div>

            <div style="grid-column: span 3; display:grid; gap: 6px;">
                <div style="font-size: 12px; opacity: .7;">Min age</div>
                <input
                        type="number"
                        min="0"
                        bind:value={ageMin}
                        style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                />
            </div>

            <div style="grid-column: span 3; display:grid; gap: 6px;">
                <div style="font-size: 12px; opacity: .7;">Max age</div>
                <input
                        type="number"
                        min="0"
                        bind:value={ageMax}
                        style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                />
            </div>
        </div>
    </div>

    <div style="margin-top: 18px;">
        {#if ageLoading}
            <div style="opacity:.7; padding: 12px 0;">Loading age categories…</div>
        {:else if ageError}
            <div style="padding: 12px; border-radius: 12px; background: #fff3f3; border: 1px solid #ffd1d1;">
                <b>ERROR</b>
                <div style="margin-top: 6px; white-space: pre-wrap;">{ageError}</div>
                <div style="margin-top: 10px; display:flex; gap: 8px;">
                    <button
                            on:click={() => loadAge({ keepPage: true })}
                            style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                    >
                        Retry
                    </button>
                </div>
            </div>
        {:else}
            <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
                <div style="font-size: 13px; opacity:.75;">
                    Page <b>{agePage}</b> • Limit <b>{ageLimit}</b> • Items <b>{ageItems.length}</b>
                </div>
                <div style="display:flex; gap: 8px;">
                    <button
                            on:click={agePrev}
                            disabled={agePage <= 1 || ageLoading}
                            style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                    >
                        Prev
                    </button>
                    <button
                            on:click={ageNext}
                            disabled={ageLoading || ageItems.length === 0}
                            style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                    >
                        Next
                    </button>
                </div>
            </div>

            {#if ageItems.length === 0}
                <div style="margin-top: 12px; padding: 12px; border-radius: 12px; border: 1px dashed #ddd; background:#fafafa;">
                    No age categories found.
                </div>
            {:else}
                <div style="margin-top: 12px; overflow:auto; border: 1px solid #eee; border-radius: 14px; background:#fff;">
                    <table style="width: 100%; border-collapse: collapse; min-width: 720px;">
                        <thead>
                        <tr style="text-align:left; background:#fafafa;">
                            <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Label</th>
                            <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Min</th>
                            <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Max</th>
                            <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Actions</th>
                        </tr>
                        </thead>
                        <tbody>
                        {#each ageItems as it, i (makeKey("age", it.id >= 1 ? it.id : null, `${it.label}-${it.min_age}-${it.max_age}`, i))}
                            <tr style="border-top: 1px solid #eee;">
                                <td style="padding: 10px 12px;">
                                    <div style="font-weight: 700;">{it.label}</div>
                                    <div style="font-size: 12px; opacity:.7;">#{it.id || "?"}</div>
                                </td>
                                <td style="padding: 10px 12px;">{it.min_age}</td>
                                <td style="padding: 10px 12px;">{it.max_age}</td>
                                <td style="padding: 10px 12px;">
                                    <div style="display:flex; gap: 8px; flex-wrap: wrap;">
                                        <button
                                                on:click={() => startEditAge(it)}
                                                style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                                        >
                                            Edit
                                        </button>
                                        <button
                                                on:click={() => deleteAge(it)}
                                                disabled={ageBusyRowId === it.id}
                                                style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                                        >
                                            {ageBusyRowId === it.id ? "Working…" : "Delete"}
                                        </button>
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
{:else}
    <div style="margin-top: 16px; display:flex; gap: 10px; flex-wrap: wrap; align-items:end;">
        <div style="flex:1; min-width: 220px; display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7;">Search</div>
            <input
                    bind:value={eduQ}
                    placeholder="e.g. Memory"
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff; width: 100%;"
            />
        </div>

        <div style="display:grid; gap: 6px;">
            <div style="font-size: 12px; opacity: .7;">Limit</div>
            <input
                    type="number"
                    min="1"
                    max="100"
                    bind:value={eduLimit}
                    style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff; width: 110px;"
            />
        </div>

        <div style="display:flex; gap: 8px; align-items: center;">
            <button
                    on:click={() => loadEdu()}
                    disabled={eduLoading}
                    style="padding: 9px 12px; border-radius: 10px; border: 1px solid #ddd; background: #111; color: #fff;"
            >
                Apply
            </button>
            <button
                    on:click={() => {
                    eduQ = "";
                    eduLimit = 24;
                    eduPage = 1;
                    void loadEdu();
                }}
                    disabled={eduLoading}
                    style="padding: 9px 12px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
            >
                Reset
            </button>
        </div>
    </div>

    <div style="margin-top: 18px; padding: 14px; border: 1px solid #eee; border-radius: 14px; background: #fff;">
        <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
            <div style="font-weight: 800;">
                {eduMode === "create" ? "Create Education Category" : `Edit Education Category #${eduEditId}`}
            </div>

            <div style="display:flex; gap: 8px;">
                {#if eduMode === "edit"}
                    <button
                            on:click={resetEduForm}
                            disabled={eduSubmitting}
                            style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
                    >
                        Cancel edit
                    </button>
                {/if}
                <button
                        on:click={submitEdu}
                        disabled={eduSubmitting}
                        style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background: #111; color: #fff;"
                >
                    {eduSubmitting ? "Saving…" : eduMode === "create" ? "Create" : "Save"}
                </button>
            </div>
        </div>

        <div style="margin-top: 12px; display:grid; grid-template-columns: repeat(12, 1fr); gap: 10px;">
            <div style="grid-column: span 12; display:grid; gap: 6px;">
                <div style="font-size: 12px; opacity: .7;">Name</div>
                <input
                        bind:value={eduName}
                        placeholder="Memory"
                        style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                />
            </div>
        </div>
    </div>

    <div style="margin-top: 18px;">
        {#if eduLoading}
            <div style="opacity:.7; padding: 12px 0;">Loading education categories…</div>
        {:else if eduError}
            <div style="padding: 12px; border-radius: 12px; background: #fff3f3; border: 1px solid #ffd1d1;">
                <b>ERROR</b>
                <div style="margin-top: 6px; white-space: pre-wrap;">{eduError}</div>
                <div style="margin-top: 10px; display:flex; gap: 8px;">
                    <button
                            on:click={() => loadEdu({ keepPage: true })}
                            style="padding: 8px 12px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                    >
                        Retry
                    </button>
                </div>
            </div>
        {:else}
            <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
                <div style="font-size: 13px; opacity:.75;">
                    Page <b>{eduPage}</b> • Limit <b>{eduLimit}</b> • Items <b>{eduItems.length}</b>
                </div>
                <div style="display:flex; gap: 8px;">
                    <button
                            on:click={eduPrev}
                            disabled={eduPage <= 1 || eduLoading}
                            style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                    >
                        Prev
                    </button>
                    <button
                            on:click={eduNext}
                            disabled={eduLoading || eduItems.length === 0}
                            style="padding: 8px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                    >
                        Next
                    </button>
                </div>
            </div>

            {#if eduItems.length === 0}
                <div style="margin-top: 12px; padding: 12px; border-radius: 12px; border: 1px dashed #ddd; background:#fafafa;">
                    No education categories found.
                </div>
            {:else}
                <div style="margin-top: 12px; overflow:auto; border: 1px solid #eee; border-radius: 14px; background:#fff;">
                    <table style="width: 100%; border-collapse: collapse; min-width: 820px;">
                        <thead>
                        <tr style="text-align:left; background:#fafafa;">
                            <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Name</th>
                            <th style="padding: 10px 12px; font-size: 12px; opacity:.7;">Actions</th>
                        </tr>
                        </thead>
                        <tbody>
                        {#each eduItems as it, i (makeKey("edu", it.id >= 1 ? it.id : null, `${it.name}`, i))}
                            <tr style="border-top: 1px solid #eee;">
                                <td style="padding: 10px 12px;">
                                    <div style="font-weight: 700;">{it.name}</div>
                                    <div style="font-size: 12px; opacity:.7;">#{it.id || "?"}</div>
                                </td>
                                <td style="padding: 10px 12px;">
                                    <div style="display:flex; gap: 8px; flex-wrap: wrap;">
                                        <button
                                                on:click={() => startEditEdu(it)}
                                                style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                                        >
                                            Edit
                                        </button>
                                        <button
                                                on:click={() => deleteEdu(it)}
                                                disabled={eduBusyRowId === it.id}
                                                style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#fff;"
                                        >
                                            {eduBusyRowId === it.id ? "Working…" : "Delete"}
                                        </button>
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
{/if}
