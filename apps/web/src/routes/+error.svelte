<script lang="ts">
	import { page } from '$app/stores';

	export let error: App.Error;
	export let status: number;

	$: message = (error?.message ?? '').trim() || 'Terjadi kesalahan pada aplikasi.';
	$: requestId = ((error as any)?.request_id ?? (error as any)?.requestId ?? '').trim() || null;
</script>

<svelte:head>
	<title>Error {status}</title>
</svelte:head>

<main class="screen">
	<section class="card" role="alert">
		<p class="kicker">Kids Planet</p>
		<h1>Error {status}</h1>
		<p class="message">{message}</p>
		{#if requestId}
			<p class="requestId">Request ID: {requestId}</p>
		{/if}
		<div class="actions">
			<a href="/" class="btn">Ke Beranda</a>
			{#if $page.url.pathname !== '/player'}
				<a href="/player" class="btn primary">Buka Player</a>
			{/if}
		</div>
	</section>
</main>

<style>
	.screen {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 24px 16px;
		background: radial-gradient(120% 120% at 100% 0%, #f5f7fa 0%, #eef2f7 45%, #ffffff 100%);
	}

	.card {
		width: min(100%, 560px);
		border-radius: 18px;
		border: 1px solid #e5e7eb;
		background: #fff;
		padding: 24px;
		box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
	}

	.kicker {
		margin: 0;
		font-size: 12px;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #64748b;
	}

	h1 {
		margin: 6px 0 0;
		font-size: clamp(26px, 5vw, 34px);
		line-height: 1.1;
		color: #111827;
	}

	.message {
		margin: 10px 0 0;
		font-size: 14px;
		line-height: 1.6;
		font-weight: 600;
		color: #334155;
		white-space: pre-wrap;
	}

	.requestId {
		margin: 10px 0 0;
		font-size: 12px;
		font-weight: 700;
		color: #475569;
	}

	.actions {
		margin-top: 18px;
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}

	.btn {
		padding: 10px 14px;
		border-radius: 999px;
		border: 1px solid #d1d5db;
		background: #fff;
		font-weight: 800;
		color: #111827;
		text-decoration: none;
	}

	.btn.primary {
		border-color: #111;
		background: #111;
		color: #fff;
	}
</style>
