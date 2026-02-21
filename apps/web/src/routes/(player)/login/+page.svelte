<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { ApiError } from '$lib/api/client';
	import { isLoggedIn, login, register } from '$lib/auth/playerAuth';

	let email = '';
	let pin = '';
	let loading = false;
	let loadingMode: 'login' | 'register' | null = null;
	let errorMsg: string | null = null;

	$: nextUrl = $page.url.searchParams.get('next') || '/';

	function validate(): string | null {
		const normalizedEmail = email.trim().toLowerCase();
		if (!normalizedEmail || !normalizedEmail.includes('@')) {
			return 'Email is invalid';
		}

		if (!/^\d{6}$/.test(pin.trim())) {
			return 'PIN must be exactly 6 digits';
		}

		return null;
	}

	async function submit(mode: 'login' | 'register') {
		errorMsg = validate();
		if (errorMsg) return;

		loading = true;
		loadingMode = mode;
		errorMsg = null;
		const normalizedEmail = email.trim().toLowerCase();
		const normalizedPin = pin.trim();

		try {
			if (mode === 'login') {
				await login(normalizedEmail, normalizedPin);
			} else {
				await register(normalizedEmail, normalizedPin);
			}

			await goto(nextUrl);
		} catch (e) {
			if (e instanceof ApiError) {
				errorMsg = e.message || 'Authentication failed';
			} else {
				errorMsg = 'Authentication failed';
			}
		} finally {
			loading = false;
			loadingMode = null;
		}
	}

	onMount(() => {
		if (isLoggedIn()) {
			void goto(nextUrl);
		}
	});
</script>

<svelte:head>
	<title>Player Login</title>
</svelte:head>

<main class="screen">
	<section class="card">
		<header class="head">
			<p class="kicker">Kids Planet</p>
			<h1>Player Login</h1>
			<p class="subtitle">Use your email and 6-digit PIN to continue playing.</p>
		</header>

		<form class="form" on:submit|preventDefault={() => submit('login')}>
			<label class="field">
				<span>Email</span>
				<input
					class="input"
					bind:value={email}
					type="email"
					autocomplete="email"
					placeholder="player@email.com"
					disabled={loading}
					required
				/>
			</label>

			<label class="field">
				<span>PIN (6 digits)</span>
				<input
					class="input"
					bind:value={pin}
					type="password"
					inputmode="numeric"
					autocomplete="one-time-code"
					minlength="6"
					maxlength="6"
					pattern="[0-9]{6}"
					placeholder="123456"
					disabled={loading}
					required
				/>
			</label>

			<p class="hint">PIN must be exactly 6 numeric digits.</p>

			{#if errorMsg}
				<p class="alert" role="alert">{errorMsg}</p>
			{/if}

			<div class="actions">
				<button class="pill primary" type="submit" disabled={loading}>
					{loadingMode === 'login' ? 'Logging in...' : 'Login'}
				</button>

				<button class="pill" type="button" on:click={() => submit('register')} disabled={loading}>
					{loadingMode === 'register' ? 'Creating account...' : 'Register'}
				</button>
			</div>
		</form>
	</section>
</main>

<style>
	.screen {
		min-height: calc(100vh - 80px);
		padding: 24px 16px;
		display: grid;
		place-items: center;
		font-family:
			system-ui,
			-apple-system,
			'Segoe UI',
			Roboto,
			Arial,
			sans-serif;
		background: radial-gradient(120% 120% at 100% 0%, #f5f7fa 0%, #eef2f7 45%, #ffffff 100%);
	}

	.card {
    	width: min(100%, 520px);
    	border: 1.5px solid #e5e7eb;
    	border-radius: 20px;
    	background: #fff;
    	box-shadow: 0 10px 25px rgba(15, 23, 42, 0.06);
    	padding: 28px 24px;
    }

	.head {
		display: grid;
		gap: 6px;
	}

	.kicker {
		margin: 0;
		color: #4b5563;
		font-size: 12px;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	h1 {
		margin: 0;
		color: #222;
		font-size: clamp(24px, 4vw, 30px);
		line-height: 1.1;
		font-weight: 900;
	}

	.subtitle {
		margin: 0;
		color: #4b5563;
		font-size: 13px;
		line-height: 1.5;
		font-weight: 600;
	}

	.form {
		margin-top: 24px;
		display: grid;
		gap: 16px;
	}

	.field {
		display: grid;
		gap: 6px;
	}

	.field span {
		color: #222;
		font-size: 13px;
		font-weight: 900;
	}

	.input {
    	width: 100%;
    	border: 1.5px solid #d1d5db;
    	border-radius: 12px;
    	padding: 12px 14px;
    	font-size: 15px;
    	font-weight: 600;
    	color: #111;
    	background: #fff;
    	outline: none;
    	box-sizing: border-box;
    	transition:
    		border-color 0.15s ease,
    		box-shadow 0.15s ease;
    }

    .input:focus {
    	border-color: #222;
    	box-shadow: 0 0 0 3px rgba(34, 34, 34, 0.08);
    }

	.input::placeholder {
		color: #9ca3af;
	}

	.input:disabled {
		cursor: not-allowed;
		opacity: 0.75;
		background: #f8fafc;
	}

	.hint {
		margin: -6px 0 0;
		color: #4b5563;
		font-size: 12px;
		font-weight: 700;
	}

	.alert {
		margin: 0;
		border: 3px solid #ef4444;
		border-radius: 12px;
		padding: 11px 14px;
		color: #991b1b;
		font-size: 13px;
		font-weight: 800;
		background: #fff;
	}

	.actions {
		margin-top: 4px;
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}

	.pill {
    	padding: 13px 18px;
    	border-radius: 999px;
    	border: 1.5px solid #d1d5db;
    	background: #fff;
    	font-size: 14px;
    	font-weight: 800;
    	color: #222;
    	cursor: pointer;
    	transition:
    		background-color 0.15s ease,
    		color 0.15s ease,
    		border-color 0.15s ease,
    		transform 0.15s ease;
    }

    .pill.primary {
    	background: #222;
    	border-color: #222;
    	color: #fff;
    }

	.pill:hover:not(:disabled) {
		background: #f5f5f5;
		transform: translateY(-1px);
	}

	.pill:disabled {
		opacity: 0.65;
		cursor: not-allowed;
	}

	.pill.primary:hover:not(:disabled) {
		background: #111;
		border-color: #111;
	}

	@media (min-width: 640px) {
		.screen {
			padding: 32px 24px;
		}

		.card {
			padding: 32px 28px;
		}
	}

	@media (max-width: 420px) {
		.actions {
			grid-template-columns: 1fr;
		}
	}
</style>