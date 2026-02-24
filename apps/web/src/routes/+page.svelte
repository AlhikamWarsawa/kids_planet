<script lang="ts">
	import PlayerShell from '$lib/components/player/PlayerShell.svelte';
	import favicon from '$lib/assets/icon.png';
</script>

<svelte:head>
	<title>Kids Planet</title>
	<meta
			name="description"
			content="Selamat datang di Kids Planet. Tempat bermain game edukasi anak dengan pengalaman yang aman dan menyenangkan."
	/>

	<link rel="icon" type="image/png" href={favicon} />

	<link rel="apple-touch-icon" href={favicon} />
</svelte:head>

<PlayerShell showTopbar={false}>
	<section class="hero" aria-label="Welcome">
		<div class="orb orbA" aria-hidden="true"></div>
		<div class="orb orbB" aria-hidden="true"></div>

		<div class="heroBody">
			<p class="kicker">Kids Planet</p>
			<h1>Selamat datang di Kids Planet</h1>
			<p>
				Pilih game edukasi yang seru, bermain dengan aman, dan lanjutkan kapan saja dari halaman
				player.
			</p>

			<div class="actions">
				<a class="player-pill cta" href="/player">Mulai Bermain</a>
			</div>

			<div class="highlights">
				<div class="block">Game edukasi pilihan</div>
				<div class="block">Filter usia & kategori</div>
				<div class="block">Riwayat bermain otomatis</div>
			</div>
		</div>

		<div class="illustration" aria-hidden="true">
			<img src={favicon} alt="" />
		</div>
	</section>
</PlayerShell>

<style>
	.hero {
		display: grid;
		grid-template-columns: 1.3fr 0.9fr;
		gap: 24px;
		align-items: center;
		padding: 28px;
		border: 3px solid #5a5a5a;
		border-radius: 24px;
		background: #fff;
		overflow: hidden;
		position: relative;
	}

	.heroBody {
		display: grid;
		align-content: center;
		gap: 12px;
		z-index: 1;
	}

	.kicker {
		margin: 0;
		font-size: 12px;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #475569;
	}

	h1 {
		margin: 0;
		font-size: clamp(30px, 5vw, 46px);
		line-height: 1.1;
		color: #0f172a;
	}

	.subtitle {
		margin: 0;
		font-size: clamp(17px, 2.6vw, 22px);
		font-weight: 700;
		color: #1d4ed8;
	}

	p {
		margin: 0;
		font-size: 15px;
		line-height: 1.6;
		font-weight: 600;
		color: #334155;
		max-width: 56ch;
	}

	.actions {
		margin-top: 8px;
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}

	.cta {
		background: #111;
		border-color: #111;
		color: #fff;
	}

	.cta:hover {
		background: #000;
	}

	.highlights {
		margin-top: 8px;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 10px;
	}

	.block {
		border-radius: 12px;
		border: 1px solid rgba(51, 65, 85, 0.15);
		background: rgba(255, 255, 255, 0.78);
		padding: 10px;
		font-size: 12px;
		font-weight: 700;
		color: #334155;
		text-align: center;
	}

	.illustration {
		display: grid;
		place-items: center;
		z-index: 1;
	}

	.illustration img {
		width: min(220px, 64vw);
		height: auto;
		filter: drop-shadow(0 12px 24px rgba(15, 23, 42, 0.2));
	}

	.orb {
		position: absolute;
		border-radius: 999px;
		filter: blur(2px);
	}

	.orbA {
		width: 140px;
		height: 140px;
		background: rgba(255, 255, 255, 0.35);
		top: -34px;
		right: 120px;
	}

	.orbB {
		width: 180px;
		height: 180px;
		background: rgba(255, 255, 255, 0.28);
		bottom: -82px;
		left: -42px;
	}

	@media (max-width: 900px) {
		.hero {
			grid-template-columns: 1fr;
			padding: 20px;
			gap: 18px;
		}

		.highlights {
			grid-template-columns: 1fr;
		}

		.illustration {
			justify-content: start;
		}
	}
</style>
