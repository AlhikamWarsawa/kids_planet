<script lang="ts">
    export let kind: "ok" | "err" = "ok";
    export let message = "";
</script>

<div class={`toast ${kind}`}>
    <div class="label">{kind === "ok" ? "OK" : "ERROR"}</div>
    <div class="message">{message}</div>
</div>

<style>
    .toast {
        padding: 10px 12px;
        border-radius: 12px;
        border: 1px solid #ddd;
        font-size: 13px;
        max-width: 420px;
        background: #fff;
    }

    .toast.ok {
        border-color: #bfe7c6;
        background: #f1fff3;
    }

    .toast.err {
        border-color: #ffd1d1;
        background: #fff3f3;
    }

    .label {
        text-transform: uppercase;
        font-size: 11px;
        letter-spacing: 0.4px;
        font-weight: 700;
    }

    .message {
        margin-top: 4px;
        white-space: pre-wrap;
    }
</style>
