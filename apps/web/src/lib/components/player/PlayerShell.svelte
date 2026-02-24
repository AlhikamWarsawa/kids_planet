<script lang="ts">
    export let showTopbar = true;
</script>

<main class="screen">
    <div class="container">
        {#if showTopbar}
            <div class="topbar">
                <div class="left">
                    <slot name="left" />
                </div>
                <div class="right">
                    <slot name="right" />
                </div>
            </div>
        {/if}

        <slot />
    </div>
</main>

<style>
    .screen {
        min-height: calc(100vh - 80px);
        padding: 20px 16px 28px;
        font-family:
            system-ui,
            -apple-system,
            "Segoe UI",
            Roboto,
            Arial,
            sans-serif;
        background: #fff;
        overflow-x: hidden;
    }

    .container {
        max-width: 1200px;
        margin: 0 auto;
    }

    .topbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        margin-bottom: 18px;
        flex-wrap: wrap;
    }

    .left {
        display: flex;
        gap: 10px;
        align-items: center;
        flex: 0 0 auto;
        min-width: 0;
    }

    .right {
        display: flex;
        gap: 12px;
        align-items: center;
        flex-wrap: wrap;
        justify-content: flex-end;
        flex: 1 1 280px;
        min-width: 0;
    }

    :global(.player-pill) {
        padding: 12px 18px;
        border-radius: 999px;
        border: 4px solid #666;
        background: #fff;
        font-weight: 700;
        cursor: pointer;
        text-decoration: none;
        color: #222;
        box-sizing: border-box;
        max-width: 100%;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    :global(.player-pill:hover) {
        background: #f5f5f5;
    }

    :global(.player-pill:disabled) {
        opacity: 0.6;
        cursor: not-allowed;
    }

    :global(.player-pill.active) {
        background: #222;
        color: #fff;
        border-color: #222;
    }

    :global(.player-pill.sm) {
        padding: 8px 12px;
        border-width: 3px;
        font-weight: 800;
        font-size: 12px;
    }

    @media (min-width: 720px) {
        .screen {
            padding: 24px 22px 34px;
        }
    }

    @media (min-width: 1024px) {
        .screen {
            padding: 26px 26px 40px;
        }
    }
</style>
