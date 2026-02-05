<script lang="ts">
    export let size = 14;
    export let color = "#555";
</script>

<span class="spinner" style={`--spinner-size: ${size}px; --spinner-color: ${color};`}></span>

<style>
    .spinner {
        display: inline-block;
        width: var(--spinner-size);
        height: var(--spinner-size);
        border: 2px solid var(--spinner-color);
        border-top-color: transparent;
        border-radius: 999px;
        animation: spin 0.7s linear infinite;
    }

    @keyframes spin {
        from {
            transform: rotate(0deg);
        }
        to {
            transform: rotate(360deg);
        }
    }
</style>
