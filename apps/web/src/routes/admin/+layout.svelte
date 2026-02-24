<script lang="ts">
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { page } from "$app/stores";
	import { auth } from "$lib/stores/auth";
	import { formatMappedError, mapApiError } from "$lib/api/errorMapper";

	let booting = true;
	let ready = false;
	let bootError: string | null = null;

	$: pathname = $page.url.pathname;

	function isPublicAdminRoute(p: string) {
		return p === "/admin/login";
	}

	async function boot() {
		booting = true;
		ready = false;
		bootError = null;

		try {
			if (isPublicAdminRoute(pathname)) {
				return;
			}

			await auth.loadFromStorage({ fetchMe: false });

			if (!$auth.token) {
				await goto("/admin/login");
				return;
			}

			await auth.fetchMe($auth.token);

			if (!$auth.token || !$auth.me) {
				await goto("/admin/login");
				return;
			}

			ready = true;
		} catch (e) {
			bootError = formatMappedError(mapApiError(e, "Failed to validate admin session"), {
				includeCode: false,
				includeRequestId: true
			});
		} finally {
			booting = false;
		}
	}

	function logout() {
		auth.clear();
		goto("/admin/login");
	}

	onMount(() => {
		void boot();
	});

	$: if (pathname) {
		void boot();
	}
</script>


<svelte:head>
	<title>Admin</title>
</svelte:head>

{#if pathname === "/admin/login"}
	<slot />
{:else if booting}
	<main style="max-width: 760px; margin: 60px auto; padding: 0 16px; font-family: system-ui, sans-serif;">
		<h1 style="margin: 0 0 8px;">Admin</h1>
		<p style="opacity: .75;">Checking session…</p>
	</main>

{:else if bootError}
	<main style="max-width: 760px; margin: 60px auto; padding: 0 16px; font-family: system-ui, sans-serif;">
		<h1 style="margin: 0 0 8px;">Admin</h1>

		<div style="margin-top: 12px; padding: 12px; border-radius: 12px; background: #fff3f3; border: 1px solid #ffd1d1;">
			<b>BOOT ERROR</b>
			<div style="margin-top: 6px; white-space: pre-wrap;">{bootError}</div>
		</div>

		<div style="display:flex; gap: 8px; margin-top: 12px; flex-wrap: wrap;">
			<button on:click={boot} style="padding: 8px 12px;">Retry</button>
			<button on:click={logout} style="padding: 8px 12px;">Logout</button>
		</div>
	</main>

{:else if ready}
	<div style="min-height: 100vh; display: flex; font-family: system-ui, sans-serif;">
		<aside style="width: 240px; border-right: 1px solid #eee; padding: 16px; background: #fafafa;">
			<div style="display:flex; align-items:center; justify-content:space-between; gap: 10px;">
				<div>
					<div style="font-weight: 800; letter-spacing: .2px;">Kids Planet</div>
					<div style="font-size: 12px; opacity: .7;">Admin</div>
				</div>
				<button
						on:click={logout}
						style="padding: 6px 10px; border-radius: 10px; border: 1px solid #ddd; background: #fff;"
				>
					Logout
				</button>
			</div>

			<div style="margin-top: 14px; font-size: 12px; opacity: .75;">
				Signed in as<br />
				<b>{$auth.me?.email}</b>
			</div>

			<nav style="margin-top: 16px; display: grid; gap: 6px;">
				<a
						href="/admin/dashboard"
						style="padding: 10px 10px; border-radius: 10px; text-decoration:none;
            background: {$page.url.pathname.startsWith('/admin/dashboard') ? '#111' : 'transparent'};
            color: {$page.url.pathname.startsWith('/admin/dashboard') ? '#fff' : '#111'};"
				>
					Dashboard
				</a>

				<a
						href="/admin/games"
						style="padding: 10px 10px; border-radius: 10px; text-decoration:none;
            background: {$page.url.pathname.startsWith('/admin/games') ? '#111' : 'transparent'};
            color: {$page.url.pathname.startsWith('/admin/games') ? '#fff' : '#111'};"
				>
					Games
				</a>

					<a
							href="/admin/categories"
							style="padding: 10px 10px; border-radius: 10px; text-decoration:none;
	            background: {$page.url.pathname.startsWith('/admin/categories') ? '#111' : 'transparent'};
	            color: {$page.url.pathname.startsWith('/admin/categories') ? '#fff' : '#111'};"
					>
						Categories
					</a>
				</nav>
			</aside>

		<main style="flex: 1; padding: 20px;">
			<slot />
		</main>
	</div>

{:else}
	<main style="max-width: 760px; margin: 60px auto; padding: 0 16px; font-family: system-ui, sans-serif;">
		<h1 style="margin: 0 0 8px;">Admin</h1>
		<p style="opacity: .75;">Redirecting…</p>
	</main>
{/if}
