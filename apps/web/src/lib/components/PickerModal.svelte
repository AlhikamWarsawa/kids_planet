<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import type { Option } from '$lib/types/picker';

    export let open = false;
    export let title = '';
    export let options: Option[] = [];
    export let selectedId: number | null = null;

    const dispatch = createEventDispatcher<{
        close: void;
        select: { id: number; label: string };
        clear: void;
    }>();

    function close() {
        dispatch('close');
    }

    function clearAll() {
        dispatch('clear');
        close();
    }

    function selectOpt(opt: Option) {
        dispatch('select', { id: opt.id, label: opt.label });
        close();
    }

    function onBackdrop(e: MouseEvent) {
        if (e.currentTarget === e.target) close();
    }

    onMount(() => {
        const onKeydown = (e: KeyboardEvent) => {
            if (!open) return;
            if (e.key === 'Escape') close();
        };
        window.addEventListener('keydown', onKeydown);
        return () => window.removeEventListener('keydown', onKeydown);
    });
</script>

{#if open}
    <div
            class="overlay"
            on:click={onBackdrop}
            aria-hidden="false"
    >
        <div
                class="modal"
                role="dialog"
                aria-modal="true"
                aria-label={title}
                tabindex="0"
                on:click|stopPropagation
        >
            <div class="modal-head">
                <div class="modal-title">{title}</div>
                <button class="pill ghost" type="button" on:click={clearAll}>
                    All
                </button>
            </div>

            <div class="grid">
                {#each options as opt (opt.id)}
                    <button
                            type="button"
                            class="card {selectedId === opt.id ? 'active' : ''}"
                            on:click={() => selectOpt(opt)}
                    >
                        {opt.label}
                    </button>
                {/each}
            </div>

            <div class="modal-foot">
                <button class="pill" type="button" on:click={close}>
                    Close
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .overlay {
        position: fixed;
        inset: 0;
        background: rgba(0,0,0,0.12);
        display: grid;
        place-items: center;
        padding: 24px;
        z-index: 50;
    }

    .modal {
        width: min(820px, 100%);
        background: #fff;
        border: 3px solid #444;
        border-radius: 14px;
        box-shadow: 0 30px 90px rgba(0,0,0,0.18);
        padding: 18px;
    }

    .modal:focus {
        outline: 3px solid rgba(17,17,17,0.2);
        outline-offset: 2px;
    }

    .modal-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 14px;
    }

    .modal-title {
        flex: 1;
        text-align: center;
        font-weight: 700;
    }

    .grid {
        display: grid;
        gap: 18px;
        grid-template-columns: repeat(3, 1fr);
        padding: 10px;
        border: 3px solid #444;
        border-radius: 12px;
    }

    @media (max-width: 720px) {
        .grid { grid-template-columns: repeat(2, 1fr); }
    }

    .card {
        height: 140px;
        border-radius: 10px;
        border: 3px solid #555;
        background: #fff;
        font-weight: 600;
        cursor: pointer;
    }

    .card.active {
        outline: 3px solid #111;
    }

    .modal-foot {
        display: flex;
        justify-content: center;
        margin-top: 14px;
    }

    .pill {
        padding: 12px 22px;
        border-radius: 999px;
        border: 3px solid #666;
        background: #fff;
        font-weight: 600;
        cursor: pointer;
    }

    .ghost {
        border-color: #bbb;
    }
</style>