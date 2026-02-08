<script lang="ts">
    import { onMount } from "svelte";
    import { auth } from "$lib/stores/auth";
    import { ApiError, createApiClient } from "$lib/api/client";
    import Toast from "$lib/components/Toast.svelte";
    import ProgressBar from "$lib/components/ProgressBar.svelte";
    import Spinner from "$lib/components/Spinner.svelte";
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

    const ZIP_MAX_BYTES = 52428800;

    type UploadResult = { object_key: string; etag: string; size: number; game_url: string };
    type LastUploadInfo = UploadResult & { file_name: string; at_ms: number };
    type UploadStage = "idle" | "uploading" | "processing";

    let loading = true;
    let errorMsg: string | null = null;

    let items: AdminGameDTO[] = [];
    let page = 1;
    let limit = 24;
    let total = 0;

    let status: AdminGameStatus | "" = "";
    let q = "";

    let busyRowId: number | null = null;

    let uploadFileById: Record<number, File | null> = {};
    let uploadingById: Record<number, boolean> = {};
    let lastUploadById: Record<number, LastUploadInfo | null> = {};
    let uploadProgressById: Record<number, number | null> = {};
    let uploadStageById: Record<number, UploadStage> = {};
    let uploadErrorById: Record<number, string | null> = {};

    let toast: { kind: "ok" | "err"; message: string } | null = null;
    let toastTimer: any = null;

    function showToast(kind: "ok" | "err", message: string) {
        toast = { kind, message };
        if (toastTimer) clearTimeout(toastTimer);
        toastTimer = setTimeout(() => (toast = null), 2200);
    }

    function setUploadProgress(gameId: number, value: number | null) {
        uploadProgressById = { ...uploadProgressById, [gameId]: value };
    }

    function setUploadStage(gameId: number, stage: UploadStage) {
        uploadStageById = { ...uploadStageById, [gameId]: stage };
    }

    function setUploadError(gameId: number, message: string | null) {
        uploadErrorById = { ...uploadErrorById, [gameId]: message };
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

    function formatMs(ms: number) {
        try {
            if (!Number.isFinite(ms)) return String(ms);
            const d = new Date(ms);
            if (Number.isNaN(d.getTime())) return String(ms);
            return d.toLocaleString();
        } catch {
            return String(ms);
        }
    }

    function formatBytes(n: number) {
        if (!Number.isFinite(n) || n < 0) return String(n);
        const units = ["B", "KB", "MB", "GB"];
        let x = n;
        let i = 0;
        while (x >= 1024 && i < units.length - 1) {
            x = x / 1024;
            i++;
        }
        const dp = i === 0 ? 0 : i === 1 ? 1 : 2;
        return `${x.toFixed(dp)} ${units[i]}`;
    }

    function playableUrlFor(g: AdminGameDTO) {
        const last = lastUploadById[g.id];
        const lastUrl = last?.game_url?.trim();
        if (lastUrl) return lastUrl;
        const gUrl = g.game_url?.trim();
        return gUrl ? gUrl : null;
    }

    function clampInt(n: number, min: number, max: number) {
        if (!Number.isFinite(n)) return min;
        return Math.max(min, Math.min(max, Math.trunc(n)));
    }

    function isZipName(name: string) {
        return name.toLowerCase().endsWith(".zip");
    }

    function validateZipFile(f: File | null): string | null {
        if (!f) return "Please choose a ZIP file.";
        if (!isZipName(f.name || "")) return "Please choose a .zip file.";
        if (!Number.isFinite(f.size) || f.size <= 0) return "That ZIP looks empty. Please try again.";
        if (f.size > ZIP_MAX_BYTES) return `That ZIP is too large. Max ${formatBytes(ZIP_MAX_BYTES)}.`;
        return null;
    }

    function describeUploadError(err: unknown) {
        if (err instanceof ApiError) {
            switch (err.code) {
                case "ZIP_TOO_LARGE":
                    return `That ZIP is too large. Max ${formatBytes(ZIP_MAX_BYTES)}.`;
                case "MISSING_INDEX_HTML":
                    return "We couldn't find an index.html at the root of the ZIP. Move it to the top level and try again.";
                case "INVALID_ZIP":
                    return "That ZIP doesn't look valid. Please export again and try.";
                case "NOT_FOUND":
                    return "We couldn't find this game anymore. Please refresh and try again.";
                case "INTERNAL_ERROR":
                case "INTERNAL_SERVER_ERROR":
                    return "Something went wrong while processing the ZIP. Please try again.";
                case "UNAUTHORIZED":
                    return "Your session expired. Please log in again.";
                case "BAD_REQUEST":
                    return err.message?.trim() || "Upload failed. Please check the ZIP and try again.";
                case "NETWORK_ERROR":
                    return "Network error. Please check your connection and try again.";
                default:
                    return err.message?.trim() || "Upload failed. Please try again.";
            }
        }

        if (err instanceof Error) {
            return "Network error. Please check your connection and try again.";
        }

        return "Upload failed. Please try again.";
    }

    function uploadZipRequest(
        gameId: number,
        file: File,
        opts?: {
            onProgress?: (value: number | null) => void;
            onStage?: (stage: UploadStage) => void;
        }
    ): Promise<UploadResult> {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            xhr.open("POST", `/api/admin/games/${gameId}/upload`);
            xhr.responseType = "json";
            xhr.setRequestHeader("Accept", "application/json");

            const token = (auth as any).getSnapshot?.().token ?? null;
            if (token) xhr.setRequestHeader("Authorization", `Bearer ${token}`);

            if (opts?.onStage) opts.onStage("uploading");

            xhr.upload.onprogress = (evt) => {
                if (!opts?.onProgress) return;
                if (evt.lengthComputable) {
                    const pct = Math.max(0, Math.min(100, Math.round((evt.loaded / evt.total) * 100)));
                    opts.onProgress(pct);
                } else {
                    opts.onProgress(null);
                }
            };

            xhr.upload.onload = () => {
                if (opts?.onStage) opts.onStage("processing");
                if (opts?.onProgress) opts.onProgress(null);
            };

            xhr.onerror = () => {
                reject(new ApiError(0, "NETWORK_ERROR", "Network error"));
            };

            xhr.onabort = () => {
                reject(new ApiError(0, "NETWORK_ERROR", "Upload canceled"));
            };

            xhr.onload = () => {
                const status = xhr.status;
                const json = xhr.response as any;

                if (status >= 200 && status < 300) {
                    const data = json && typeof json === "object" && "data" in json ? json.data : json;
                    resolve(data as UploadResult);
                    return;
                }

                const code = json?.error?.code || (status === 401 ? "UNAUTHORIZED" : "HTTP_ERROR");
                const message = json?.error?.message || xhr.statusText || "Request failed";
                reject(new ApiError(status, code, message));
            };

            const fd = new FormData();
            fd.append("file", file, file.name);
            xhr.send(fd);
        });
    }

    function setUploadFile(gameId: number, f: File | null) {
        uploadFileById = { ...uploadFileById, [gameId]: f };
    }

    function onZipInputChange(gameId: number, e: Event) {
        const input = e.currentTarget as HTMLInputElement | null;
        const f = input?.files && input.files.length > 0 ? input.files[0] : null;

        if (!f) {
            setUploadFile(gameId, null);
            setUploadError(gameId, null);
            setUploadStage(gameId, "idle");
            setUploadProgress(gameId, null);
            return;
        }

        const v = validateZipFile(f);
        if (v) {
            setUploadFile(gameId, null);
            setUploadError(gameId, v);
            setUploadStage(gameId, "idle");
            setUploadProgress(gameId, null);
            if (input) input.value = "";
            showToast("err", v);
            return;
        }

        setUploadFile(gameId, f);
        setUploadError(gameId, null);
        setUploadStage(gameId, "idle");
        setUploadProgress(gameId, null);
    }

    async function doUploadZip(g: AdminGameDTO) {
        if (!g?.id) return;

        if (g.status === "archived") {
            showToast("err", "cannot upload to archived game");
            return;
        }

        const gameId = g.id;
        if (uploadingById[gameId]) return;

        const f = uploadFileById[gameId] ?? null;
        const err = validateZipFile(f);
        if (err) {
            showToast("err", err);
            return;
        }
        if (!f) {
            showToast("err", "file is required");
            return;
        }

        uploadingById = { ...uploadingById, [gameId]: true };
        setUploadError(gameId, null);
        setUploadStage(gameId, "uploading");
        setUploadProgress(gameId, 0);

        try {
            const res = await uploadZipRequest(gameId, f, {
                onProgress: (value) => setUploadProgress(gameId, value),
                onStage: (stage) => setUploadStage(gameId, stage),
            });

            lastUploadById = {
                ...lastUploadById,
                [gameId]: {
                    object_key: res.object_key,
                    etag: res.etag,
                    size: res.size,
                    game_url: res.game_url,
                    file_name: f.name,
                    at_ms: Date.now(),
                },
            };

            setUploadFile(gameId, null);
            showToast("ok", "Upload complete. Game ready to play.");
        } catch (e) {
            const msg = describeUploadError(e);
            setUploadError(gameId, msg);
            showToast("err", msg);
        } finally {
            uploadingById = { ...uploadingById, [gameId]: false };
            setUploadStage(gameId, "idle");
            setUploadProgress(gameId, null);
        }
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

            if ((!Number.isFinite(Number(formAgeCategoryId)) || Number(formAgeCategoryId) < 1) && ageCats.length > 0) {
                formAgeCategoryId = ageCats[0].id as any;
            }
        } catch (e) {
            if (e instanceof ApiError) ageCatsError = `${e.code}: ${e.message}`;
            else ageCatsError = String(e);
        } finally {
            ageCatsLoading = false;
        }
    }

    function normalizeText(v: any) {
        return v == null ? "" : String(v);
    }

    function normalizeNullableText(v: any) {
        if (v == null) return null;
        const s = String(v);
        return s.trim() ? s : null;
    }

    function normalizeBool(v: any) {
        if (typeof v === "boolean") return v;
        if (typeof v === "number") return v === 1;
        if (typeof v === "string") {
            const s = v.trim().toLowerCase();
            return s === "true" || s === "1";
        }
        return false;
    }

    function normalizeStatus(v: any): AdminGameStatus {
        const s = String(v ?? "").trim().toLowerCase();
        if (s === "active" || s === "draft" || s === "archived") return s;
        return "draft";
    }

    function normalizeGame(raw: any): AdminGameDTO | null {
        const id = pickId(raw);
        if (!id) return null;
        return {
            id,
            title: normalizeText(raw?.title ?? raw?.Title),
            slug: normalizeText(raw?.slug ?? raw?.Slug),
            status: normalizeStatus(raw?.status ?? raw?.Status),
            thumbnail: normalizeNullableText(raw?.thumbnail ?? raw?.Thumbnail),
            game_url: normalizeNullableText(raw?.game_url ?? raw?.gameUrl ?? raw?.GameUrl ?? raw?.GameURL),
            age_category_id: clampInt(Number(raw?.age_category_id ?? raw?.AgeCategoryID ?? raw?.ageCategoryId ?? 0), 0, 1_000_000_000),
            free: normalizeBool(raw?.free ?? raw?.Free),
            created_at: normalizeText(raw?.created_at ?? raw?.CreatedAt),
            updated_at: normalizeText(raw?.updated_at ?? raw?.UpdatedAt),
        };
    }

    type Mode = "create" | "edit";
    let mode: Mode = "create";
    let editId: number | null = null;

    let formTitle = "";
    let formSlug = "";
    let formAgeCategoryId: any = 1;
    let formFree = true;

    let submitting = false;

    function resetForm() {
        mode = "create";
        editId = null;
        formTitle = "";
        formSlug = "";
        if (ageCats.length > 0) formAgeCategoryId = ageCats[0].id as any;
        else formAgeCategoryId = 1;
        formFree = true;
    }

    function startEdit(g: AdminGameDTO) {
        mode = "edit";
        editId = g.id;
        formTitle = g.title ?? "";
        formSlug = g.slug ?? "";
        formAgeCategoryId = (g.age_category_id ?? (ageCats[0]?.id ?? 1)) as any;
        formFree = Boolean(g.free);
        showToast("ok", "Edit mode");
        window.scrollTo({ top: 0, behavior: "smooth" });
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
            const rawItems = Array.isArray((data as any)?.items) ? (data as any).items : [];
            items = rawItems.map(normalizeGame).filter(Boolean) as AdminGameDTO[];

            page = clampInt(Number((data as any)?.page ?? page), 1, 1_000_000_000);
            limit = clampInt(Number((data as any)?.limit ?? limit), 1, 100);
            total = clampInt(Number((data as any)?.total ?? 0), 0, 1_000_000_000);

            const ids = new Set(items.map((x) => x.id));
            const nextFiles: Record<number, File | null> = {};
            const nextUploading: Record<number, boolean> = {};
            const nextLast: Record<number, LastUploadInfo | null> = {};
            const nextProgress: Record<number, number | null> = {};
            const nextStage: Record<number, UploadStage> = {};
            const nextError: Record<number, string | null> = {};

            for (const id of ids) {
                nextFiles[id] = uploadFileById[id] ?? null;
                nextUploading[id] = uploadingById[id] ?? false;
                nextLast[id] = lastUploadById[id] ?? null;
                nextProgress[id] = uploadProgressById[id] ?? null;
                nextStage[id] = uploadStageById[id] ?? "idle";
                nextError[id] = uploadErrorById[id] ?? null;
            }
            uploadFileById = nextFiles;
            uploadingById = nextUploading;
            lastUploadById = nextLast;
            uploadProgressById = nextProgress;
            uploadStageById = nextStage;
            uploadErrorById = nextError;
        } catch (e) {
            if (e instanceof ApiError) errorMsg = `${e.code}: ${e.message}`;
            else errorMsg = String(e);
        } finally {
            loading = false;
        }
    }

    function totalPages() {
        const perPage = limit > 0 ? limit : 24;
        return Math.max(1, Math.ceil((total || 0) / perPage));
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
        <div style="font-size: 13px; opacity:.7;">Create, edit, publish, and upload ZIP</div>
    </div>

    {#if toast}
        <Toast kind={toast.kind} message={toast.message} />
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

<!-- Create / Edit form -->
<div style="margin-top: 18px; padding: 14px; border: 1px solid #eee; border-radius: 14px; background: #fff;">
    <div style="display:flex; align-items:center; justify-content:space-between; gap: 12px; flex-wrap: wrap;">
        <div>
            <div style="font-weight: 800;">
                {mode === "create" ? "Create Game" : `Edit Game #${editId}`}
            </div>
            <div style="font-size: 12px; opacity:.7; margin-top: 2px;">
                Title + Slug + Age Category + Free
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
                    disabled={submitting || ageCatsLoading || !!ageCatsError || ageCats.length === 0}
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
                    disabled={ageCatsLoading || !!ageCatsError || ageCats.length === 0}
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

<!-- List -->
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
                <table style="width: 100%; border-collapse: collapse; min-width: 1040px;">
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

                            <td style="padding: 10px 12px;">{g.age_category_id}</td>

                            <td style="padding: 10px 12px; font-size: 13px; opacity:.8;">
                                {formatDate(g.updated_at)}
                            </td>

                            <td style="padding: 10px 12px;">
                                <div style="display:grid; gap: 10px;">
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

                                    <div style="display:grid; gap: 6px;">
                                        <div style="font-size: 12px; opacity:.7;">
                                            Upload ZIP (max {formatBytes(ZIP_MAX_BYTES)})
                                        </div>

                                        <div style="display:flex; gap: 8px; align-items:center; flex-wrap: wrap;">
                                            <input
                                                    type="file"
                                                    accept=".zip"
                                                    disabled={(uploadingById[g.id] ?? false) || g.status === "archived"}
                                                    on:change={(e) => onZipInputChange(g.id, e)}
                                                    style="max-width: 260px;"
                                            />

                                            <button
                                                    on:click={() => doUploadZip(g)}
                                                    disabled={(uploadingById[g.id] ?? false) || !(uploadFileById[g.id] ?? null) || g.status === "archived"}
                                                    style="padding: 7px 10px; border-radius: 10px; border: 1px solid #ddd; background:#111; color:#fff;"
                                            >
                                                {(uploadingById[g.id] ?? false) ? "Uploading…" : "Upload ZIP"}
                                            </button>

                                            {#if uploadFileById[g.id]}
                                                    <span style="font-size: 12px; opacity:.75;">
                                                        {(uploadFileById[g.id] as File).name} • {formatBytes((uploadFileById[g.id] as File).size)}
                                                    </span>
                                            {/if}
                                        </div>

                                        {#if uploadingById[g.id]}
                                            <div style="display:grid; gap: 6px; max-width: 320px;">
                                                <div style="display:flex; gap: 6px; align-items:center; font-size: 12px; opacity:.8;">
                                                    <Spinner size={14} />
                                                    {#if (uploadStageById[g.id] ?? "idle") === "processing"}
                                                        <span>Processing ZIP…</span>
                                                    {:else}
                                                        <span>
                                                            {(uploadProgressById[g.id] ?? null) != null
                                                                ? `Uploading ${uploadProgressById[g.id]}%`
                                                                : "Uploading…"}
                                                        </span>
                                                    {/if}
                                                </div>
                                                <ProgressBar
                                                        value={(uploadStageById[g.id] ?? "idle") === "processing"
                                                            ? null
                                                            : (uploadProgressById[g.id] ?? null)}
                                                />
                                            </div>
                                        {/if}

                                        {#if uploadErrorById[g.id]}
                                            <div
                                                    style="
                                                    font-size: 12px;
                                                    color:#b42318;
                                                    background:#fff3f3;
                                                    border:1px solid #ffd1d1;
                                                    border-radius: 10px;
                                                    padding: 6px 8px;
                                                "
                                            >
                                                {uploadErrorById[g.id]}
                                            </div>
                                        {/if}

                                        <div style="font-size: 12px; opacity:.8; line-height: 1.35;">
                                            <div>
                                                <b>Playable</b>:
                                                {#if playableUrlFor(g)}
                                                    <a
                                                            href={playableUrlFor(g)}
                                                            target="_blank"
                                                            rel="noreferrer"
                                                            style="font-family: ui-monospace, SFMono-Regular, Menlo, monospace;"
                                                    >
                                                        {playableUrlFor(g)}
                                                    </a>
                                                {:else}
                                                    <span>Game not available yet.</span>
                                                {/if}
                                            </div>
                                            {#if lastUploadById[g.id]}
                                                <div>
                                                    <b>Last</b>:
                                                    {(lastUploadById[g.id] as LastUploadInfo).file_name}
                                                    • {formatBytes((lastUploadById[g.id] as LastUploadInfo).size)}
                                                    • {formatMs((lastUploadById[g.id] as LastUploadInfo).at_ms)}
                                                </div>
                                                <div style="font-family: ui-monospace, SFMono-Regular, Menlo, monospace;">
                                                    key: {(lastUploadById[g.id] as LastUploadInfo).object_key}
                                                </div>
                                                <div style="font-family: ui-monospace, SFMono-Regular, Menlo, monospace;">
                                                    etag: {(lastUploadById[g.id] as LastUploadInfo).etag}
                                                </div>
                                            {/if}
                                        </div>

                                        {#if g.status === "archived"}
                                            <div style="font-size: 12px; color:#b42318;">
                                                Upload disabled for archived games.
                                            </div>
                                        {/if}
                                    </div>
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
