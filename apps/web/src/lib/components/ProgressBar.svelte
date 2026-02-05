<script lang="ts">
    export let value: number | null = null;
    export let height = 6;

    $: clamped = value == null ? null : Math.max(0, Math.min(100, value));
</script>

<div
        class="track"
        style={`height: ${height}px;`}
        role="progressbar"
        aria-valuemin="0"
        aria-valuemax="100"
        aria-valuenow={clamped ?? undefined}
        aria-busy={clamped == null ? "true" : "false"}
>
    <div class="fill" class:indeterminate={clamped == null} style={clamped == null ? "" : `width: ${clamped}%;`}></div>
</div>

<style>
    .track {
        width: 100%;
        border-radius: 999px;
        background: #f0f0f0;
        overflow: hidden;
        position: relative;
    }

    .fill {
        height: 100%;
        background: linear-gradient(90deg, #111, #555);
        border-radius: inherit;
        transition: width 150ms ease;
    }

    .fill.indeterminate {
        position: absolute;
        width: 40%;
        left: -40%;
        animation: progress-slide 1.1s ease-in-out infinite;
    }

    @keyframes progress-slide {
        0% {
            left: -40%;
        }
        50% {
            left: 30%;
        }
        100% {
            left: 100%;
        }
    }
</style>
