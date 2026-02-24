<script lang="ts">
	import { onMount } from 'svelte';
	import { getPublicCategories, type AgeCategoryDTO, type EducationCategoryDTO } from '$lib/api/categories';
	import { listGames } from '$lib/api/games';
	import type { GameListItem } from '$lib/types/game';
	import { isLoggedIn as isPlayerLoggedIn, logout as logoutPlayer } from '$lib/auth/playerAuth';
	import { formatMappedError, mapApiError } from '$lib/api/errorMapper';
	import { formatGameAgeTag, normalizePlayCount } from '$lib/utils/gameDisplay';

	import GameCard from '$lib/components/GameCard.svelte';
	import PickerModal from '$lib/components/PickerModal.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import PlayerShell from '$lib/components/player/PlayerShell.svelte';
	import type { Option } from '$lib/types/picker';

	let ageOptions: Option[] = [];
	let educationOptions: Option[] = [];
	let ageCatalogById: Record<number, AgeCategoryDTO> = {};
	let educationCatalogById: Record<number, EducationCategoryDTO> = {};

	let games: GameListItem[] = [];
	let page = 1;
	let limit = 24;
	let total = 0;
	let canLoadMore = false;

	let loading = false;
	let loadingMore = false;

	let errorInitial: string | null = null;
	let errorMore: string | null = null;

	let selectedAge: number[] = [];
	let selectedCategory: number[] = [];

	let sort: 'newest' | 'popular' = 'newest';
	const sortOptions = [
		{ value: 'newest', label: 'Terbaru' },
		{ value: 'popular', label: 'Popular' }
	] as const;

	let showAgeModal = false;
	let showCatModal = false;
	let loggingOut = false;
	let playerLoggedIn = false;
	let toast: { kind: 'ok' | 'err'; message: string } | null = null;
	let toastTimer: ReturnType<typeof setTimeout> | null = null;

	let reqSeq = 0;
	const playerLoginHref = '/login?next=/player';

	const hasMore = () => canLoadMore;

	function showToast(kind: 'ok' | 'err', message: string) {
		toast = { kind, message };
		if (toastTimer) clearTimeout(toastTimer);
		toastTimer = setTimeout(() => {
			toast = null;
			toastTimer = null;
		}, 2400);
	}

	function normalizeIds(ids: number[]): number[] {
		return [...new Set(ids.filter((id) => Number.isFinite(id) && id >= 1))];
	}

	function parseSelection(detail: unknown): number[] {
		const row = detail && typeof detail === 'object' ? (detail as Record<string, unknown>) : null;
		if (row && Array.isArray(row.ids)) {
			return normalizeIds(row.ids.map((id) => Number(id)));
		}

		const id = Number(row?.id);
		if (Number.isFinite(id) && id >= 1) return [id];
		return [];
	}

	function ageLabelFromCatalog(ageId: number): string | null {
		const category = ageCatalogById[ageId];
		if (!category) return null;
		const raw = category.label?.trim();
		if (raw) return raw;
		return formatGameAgeTag({
			min_age: category.min_age,
			max_age: category.max_age
		});
	}

	function educationLabelFromCatalog(categoryId: number): string | null {
		const category = educationCatalogById[categoryId];
		if (!category) return null;
		const name = category.name?.trim();
		return name || `Category ${categoryId}`;
	}

	function applyCategoryCatalog(ages: AgeCategoryDTO[], education: EducationCategoryDTO[]) {
		const nextAgeMap: Record<number, AgeCategoryDTO> = {};
		for (const category of ages) {
			const id = Number(category.id);
			if (!Number.isFinite(id) || id < 1) continue;
			nextAgeMap[id] = category;
		}

		const nextEducationMap: Record<number, EducationCategoryDTO> = {};
		for (const category of education) {
			const id = Number(category.id);
			if (!Number.isFinite(id) || id < 1) continue;
			nextEducationMap[id] = category;
		}

		ageCatalogById = nextAgeMap;
		educationCatalogById = nextEducationMap;

		ageOptions = Object.values(nextAgeMap)
			.sort((a, b) => a.id - b.id)
			.map((category) => ({
				id: category.id,
				label:
					category.label?.trim() ||
					formatGameAgeTag({
						min_age: category.min_age,
						max_age: category.max_age
					})
			}));

		educationOptions = Object.values(nextEducationMap)
			.sort((a, b) => a.id - b.id)
			.map((category) => ({
				id: category.id,
				label: category.name?.trim() || `Category ${category.id}`
			}));
	}

	async function loadCategoryCatalog() {
		try {
			const catalog = await getPublicCategories();
			applyCategoryCatalog(catalog.age_categories, catalog.education_categories);
		} catch (error) {
			const mapped = mapApiError(error, 'Failed to load categories.');
			showToast('err', formatMappedError(mapped, { includeCode: false, includeRequestId: true }));
		}
	}

	function shouldFallbackToClientFilter(): boolean {
		return selectedAge.length > 1 || selectedCategory.length > 1;
	}

	function hasMoreFromServer(pageNumber: number, pageSize: number, totalRows: number): boolean {
		return pageNumber * pageSize < totalRows;
	}

	function totalLabel(): string {
		if (!shouldFallbackToClientFilter()) return String(total);
		if (hasMore()) return `${games.length}+`;
		return String(games.length);
	}

	function gameMatchesActiveFilters(game: GameListItem): boolean {
		if (selectedAge.length > 0 && !selectedAge.includes(Number(game.age_category_id))) {
			return false;
		}

		if (selectedCategory.length === 0) return true;

		const idsFromList = Array.isArray(game.education_category_ids)
			? game.education_category_ids
			: [];
		const idsFromObjects = Array.isArray(game.education_categories)
			? game.education_categories.map((c) => Number(c.id))
			: [];
		const gameEducationIds = new Set([
			...idsFromList.map((id) => Number(id)),
			...idsFromObjects.map((id) => Number(id))
		]);

		return selectedCategory.some((id) => gameEducationIds.has(id));
	}

	function applyClientFilters(items: GameListItem[]): GameListItem[] {
		if (selectedAge.length === 0 && selectedCategory.length === 0) return items;
		return items.filter((item) => gameMatchesActiveFilters(item));
	}

	function sortGamesForPopular(items: GameListItem[]): GameListItem[] {
		if (sort !== 'popular') return items;
		const hasPlayCount = items.some((item) => item.play_count != null);
		if (!hasPlayCount) return items;

		return [...items].sort((a, b) => {
			const byPlayCount = normalizePlayCount(b.play_count) - normalizePlayCount(a.play_count);
			if (byPlayCount !== 0) return byPlayCount;
			const byDate = Date.parse(b.created_at) - Date.parse(a.created_at);
			if (Number.isFinite(byDate) && byDate !== 0) return byDate;
			return b.id - a.id;
		});
	}

	function dedupeGames(items: GameListItem[]): GameListItem[] {
		const seen = new Set<number>();
		const out: GameListItem[] = [];
		for (const item of items) {
			if (seen.has(item.id)) continue;
			seen.add(item.id);
			out.push(item);
		}
		return out;
	}

	function hydrateFilterOptions(sourceGames: GameListItem[]) {
		const ageMap = new Map<number, string>(ageOptions.map((option) => [option.id, option.label]));

		for (const game of sourceGames) {
			const ageId = Number(game.age_category_id);
			if (!Number.isFinite(ageId) || ageId < 1) continue;
			ageMap.set(ageId, ageLabelFromCatalog(ageId) ?? formatGameAgeTag(game));
		}

		const nextAge = [...ageMap.entries()]
			.sort((a, b) => a[0] - b[0])
			.map(([id, label]) => ({ id, label }));
		if (nextAge.length > 0) ageOptions = nextAge;

		const educationMap = new Map<number, string>(
			educationOptions.map((option) => [option.id, option.label])
		);

			for (const game of sourceGames) {
				const categories = Array.isArray(game.education_categories) ? game.education_categories : [];
				for (const category of categories) {
					const id = Number(category.id);
					if (!Number.isFinite(id) || id < 1) continue;
					const name = String(category.name ?? '').trim();
					educationMap.set(id, name || `Category ${id}`);
				}

			const categoryIds = Array.isArray(game.education_category_ids)
				? game.education_category_ids
				: [];
			for (const categoryId of categoryIds) {
				const id = Number(categoryId);
				if (!Number.isFinite(id) || id < 1 || educationMap.has(id)) continue;
					educationMap.set(
						id,
						educationLabelFromCatalog(id) ?? `Category ${id}`
					);
				}
			}

		const nextEducation = [...educationMap.entries()]
			.sort((a, b) => a[0] - b[0])
			.map(([id, label]) => ({ id, label }));
		if (nextEducation.length > 0) educationOptions = nextEducation;
	}

	function getAgeLabel(ids: number[]) {
		if (ids.length === 0) return 'Pilih Usia';
		if (ids.length === 1) return ageOptions.find((option) => option.id === ids[0])?.label ?? 'Pilih Usia';
		return `${ids.length} usia dipilih`;
	}

	function getCatLabel(ids: number[]) {
		if (ids.length === 0) return 'Pilih Kategori';
		if (ids.length === 1)
			return educationOptions.find((option) => option.id === ids[0])?.label ?? 'Pilih Kategori';
		return `${ids.length} kategori dipilih`;
	}

	$: ageLabel = getAgeLabel(selectedAge);
	$: catLabel = getCatLabel(selectedCategory);

	async function loadInitial(opts: { keepList?: boolean } = {}) {
		const seq = ++reqSeq;
		const fallbackMode = shouldFallbackToClientFilter();
		const queryLimit = fallbackMode ? 100 : limit;

		if (!opts.keepList) loading = true;

		errorInitial = null;
		errorMore = null;

		try {
			const res = await listGames({
				age_category_id: selectedAge.length > 0 ? selectedAge : undefined,
				education_category_id: selectedCategory.length > 0 ? selectedCategory : undefined,
				sort,
				page: 1,
				limit: queryLimit
			});

			if (seq !== reqSeq) return;

			hydrateFilterOptions(res.items);
			games = sortGamesForPopular(dedupeGames(applyClientFilters(res.items)));
			page = res.page;
			limit = res.limit;
			total = fallbackMode ? games.length : res.total;
			canLoadMore = fallbackMode ? false : hasMoreFromServer(res.page, res.limit, res.total);
		} catch (e) {
			if (seq !== reqSeq) return;
			const mapped = mapApiError(e, 'Failed to load games.');
			if (mapped.status === 404) {
				errorInitial = null;
				games = [];
				total = 0;
				page = 1;
				canLoadMore = false;
				return;
			}
			errorInitial = formatMappedError(mapped, {
				includeCode: false,
				includeRequestId: true
			});
			if (mapped.status === 400) {
				showToast('err', errorInitial);
			}

			if (!opts.keepList) {
				games = [];
				total = 0;
				page = 1;
				canLoadMore = false;
			}
		} finally {
			if (seq === reqSeq) loading = false;
		}
	}

	async function loadMore() {
		if (loading || loadingMore || !hasMore()) return;
		if (shouldFallbackToClientFilter()) return;

		const seq = ++reqSeq;
		loadingMore = true;
		errorMore = null;

		try {
			const res = await listGames({
				age_category_id: selectedAge.length > 0 ? selectedAge : undefined,
				education_category_id: selectedCategory.length > 0 ? selectedCategory : undefined,
				sort,
				page: page + 1,
				limit
			});

			if (seq !== reqSeq) return;

			hydrateFilterOptions(res.items);
			games = sortGamesForPopular(dedupeGames([...games, ...applyClientFilters(res.items)]));
			page = res.page;
			limit = res.limit;
			total = res.total;
			canLoadMore = hasMoreFromServer(res.page, res.limit, res.total);
		} catch (e) {
			if (seq !== reqSeq) return;
			const mapped = mapApiError(e, 'Failed to load more games.');
			if (mapped.status === 404) {
				canLoadMore = false;
				errorMore = null;
				return;
			}
			errorMore = formatMappedError(mapped, {
				includeCode: false,
				includeRequestId: true
			});
			if (mapped.status === 400) {
				showToast('err', errorMore);
			}
		} finally {
			if (seq === reqSeq) loadingMore = false;
		}
	}

	function applyFilters(opts: { keepList?: boolean } = {}) {
		page = 1;
		loadInitial({ keepList: opts.keepList ?? true });
	}

	function setSort(next: 'newest' | 'popular') {
		if (sort === next) return;
		sort = next;
		applyFilters({ keepList: false });
	}

	async function handleLogout() {
		if (loggingOut) return;
		loggingOut = true;
		try {
			await logoutPlayer();
		} catch {
		} finally {
			playerLoggedIn = false;
			loggingOut = false;
		}
	}

	onMount(() => {
		playerLoggedIn = isPlayerLoggedIn();
		void Promise.all([loadCategoryCatalog(), loadInitial()]);
	});
</script>

<svelte:head>
	<title>Kids Planet Player</title>
	<meta
		name="description"
		content="Kids Planet player area. Browse educational games by age, category, and sorting, then start playing."
	/>
</svelte:head>

<PlayerShell>
	{#if toast}
		<Toast kind={toast.kind} message={toast.message} />
	{/if}

	<svelte:fragment slot="left">
		{#if playerLoggedIn}
			<button class="player-pill" type="button" on:click={handleLogout} disabled={loggingOut}>
				{loggingOut ? 'Logging out...' : 'Logout'}
			</button>
		{:else}
			<a class="player-pill" href={playerLoginHref}>Login</a>
		{/if}
		<a class="player-pill" href="/player/history">History</a>
	</svelte:fragment>

	<svelte:fragment slot="right">
		<div class="sortGroup" role="group" aria-label="Sort games">
			{#each sortOptions as opt}
				<button
					class="player-pill sortPill"
					class:active={sort === opt.value}
					type="button"
					on:click={() => setSort(opt.value)}
					aria-pressed={sort === opt.value}
					disabled={loading}
				>
					{opt.label}
				</button>
			{/each}
		</div>

		<button
			class="player-pill"
			class:active={selectedAge.length > 0}
			type="button"
			on:click={() => (showAgeModal = true)}
		>
			{ageLabel}
		</button>

		<button
			class="player-pill"
			class:active={selectedCategory.length > 0}
			type="button"
			on:click={() => (showCatModal = true)}
		>
			{catLabel}
		</button>
	</svelte:fragment>

	{#if loading}
		<div class="grid" aria-busy="true" aria-live="polite">
			{#each Array(8) as _, i (i)}
				<div class="box skeleton" aria-hidden="true"></div>
			{/each}
		</div>
	{:else if errorInitial && games.length === 0}
		<div class="state error" role="alert">
			<div class="state-title">Gagal load games</div>
			<div class="state-sub">{errorInitial}</div>
			<div class="actions">
				<button class="player-pill" type="button" on:click={() => loadInitial()}>Retry</button>
			</div>
		</div>
	{:else if games.length === 0}
		<div class="empty">
			<div class="empty-title">Belum ada game</div>
			<div class="empty-sub">Coba ubah filter/sort atau cek lagi nanti.</div>
			{#if errorInitial}
				<div class="miniErr" role="alert">{errorInitial}</div>
				<button class="player-pill" type="button" on:click={() => loadInitial()}>Retry</button>
			{/if}
		</div>
	{:else}
		{#if errorInitial}
			<div class="banner" role="alert">
				<span>{errorInitial}</span>
				<button class="player-pill sm" type="button" on:click={() => loadInitial({ keepList: true })}>
					Retry
				</button>
			</div>
		{/if}

		<div class="grid">
			{#each games as game (game.id)}
				<GameCard {game} />
			{/each}
		</div>

		<div class="footer">
			<div class="meta">
				Showing <b>{games.length}</b> of <b>{totalLabel()}</b>
			</div>

			<div class="footerRight">
				{#if errorMore}
					<div class="moreErr" role="alert">
						{errorMore}
						<button class="player-pill sm" type="button" on:click={loadMore} disabled={loadingMore}>
							Retry
						</button>
					</div>
				{/if}

				{#if hasMore()}
					<button class="player-pill" on:click={loadMore} disabled={loadingMore}>
						{loadingMore ? 'Loading...' : 'Load more'}
					</button>
				{/if}
			</div>
		</div>
	{/if}

	<PickerModal
		open={showAgeModal}
		title="Pilih Usia"
		options={ageOptions}
		selectedIds={selectedAge}
		multiple
		on:close={() => (showAgeModal = false)}
		on:clear={() => {
			selectedAge = [];
			showAgeModal = false;
			applyFilters();
		}}
		on:select={(e) => {
			selectedAge = parseSelection(e.detail);
			showAgeModal = false;
			applyFilters();
		}}
	/>

	<PickerModal
		open={showCatModal}
		title="Pilih Kategori"
		options={educationOptions}
		selectedIds={selectedCategory}
		multiple
		on:close={() => (showCatModal = false)}
		on:clear={() => {
			selectedCategory = [];
			showCatModal = false;
			applyFilters();
		}}
		on:select={(e) => {
			selectedCategory = parseSelection(e.detail);
			showCatModal = false;
			applyFilters();
		}}
	/>
</PlayerShell>

<style>
	.sortGroup {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		align-items: center;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 16px;
		min-width: 0;
	}

	@media (min-width: 720px) {
		.grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
			gap: 20px;
		}
	}

	@media (min-width: 1024px) {
		.grid {
			grid-template-columns: repeat(4, minmax(0, 1fr));
			gap: 24px;
		}
	}

	@media (min-width: 1280px) {
		.grid {
			grid-template-columns: repeat(5, minmax(0, 1fr));
		}
	}

	.box {
		height: 160px;
		border: 4px solid #666;
		border-radius: 12px;
		background: #fff;
		box-sizing: border-box;
	}

	.skeleton {
		position: relative;
		overflow: hidden;
	}

	.skeleton::after {
		content: '';
		position: absolute;
		inset: 0;
		transform: translateX(-100%);
		background: linear-gradient(
			90deg,
			rgba(255, 255, 255, 0) 0%,
			rgba(0, 0, 0, 0.06) 50%,
			rgba(255, 255, 255, 0) 100%
		);
		animation: shimmer 1.4s infinite;
	}

	@keyframes shimmer {
		100% {
			transform: translateX(100%);
		}
	}

	.footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 18px;
		gap: 12px;
		flex-wrap: wrap;
	}

	.footerRight {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
		justify-content: flex-end;
	}

	.meta {
		color: #222;
		opacity: 0.7;
		font-size: 13px;
	}

	.state {
		margin: 0 0 14px;
		padding: 14px 16px;
		border-radius: 12px;
		border: 2px solid #666;
		background: #fff;
		color: #222;
		font-weight: 800;
		box-sizing: border-box;
		max-width: 720px;
	}

	.state.error {
		border-color: #ef4444;
		color: #991b1b;
		background: #fff;
	}

	.state-title {
		font-weight: 900;
		margin-bottom: 6px;
	}

	.state-sub {
		opacity: 0.85;
		font-size: 13px;
	}

	.actions {
		margin-top: 10px;
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}

	.banner {
		margin: 0 0 12px;
		padding: 12px 14px;
		border-radius: 12px;
		border: 2px solid #ef4444;
		background: #fff;
		color: #991b1b;
		font-weight: 900;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
	}

	.moreErr {
		display: flex;
		gap: 10px;
		align-items: center;
		flex-wrap: wrap;
		padding: 8px 10px;
		border-radius: 12px;
		border: 2px solid #ef4444;
		color: #991b1b;
		font-weight: 900;
		font-size: 12px;
		background: #fff;
	}

	.empty {
		border: 4px solid #666;
		border-radius: 12px;
		padding: 18px;
		background: #fff;
		box-sizing: border-box;
		max-width: 720px;
	}

	.empty-title {
		font-weight: 800;
		margin-bottom: 6px;
		color: #222;
	}

	.empty-sub {
		opacity: 0.7;
		color: #222;
		font-size: 13px;
	}

	.miniErr {
		margin-top: 12px;
		padding: 10px 12px;
		border-radius: 12px;
		border: 2px solid #ef4444;
		color: #991b1b;
		font-weight: 900;
		font-size: 12px;
		background: #fff;
	}
</style>
