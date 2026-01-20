<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api, ApiError } from "$lib/api/client";
    import { auth } from "$lib/stores/auth";

    type LoginResponse = {
        access_token: string;
        token_type: string;
        expires_in: number;
    };

    let email = "";
    let password = "";

    let loading = false;
    let errorMsg: string | null = null;

    let checkingSession = true;

    function validate() {
        const e = email.trim().toLowerCase();
        if (!e || !e.includes("@")) return "Email tidak valid";
        if (!password || password.length < 4) return "Password minimal 4 karakter";
        return null;
    }

    async function checkAlreadyLoggedIn() {
        checkingSession = true;
        try {
            await auth.loadFromStorage({ fetchMe: false });

            if (!$auth.token) return;

            await auth.fetchMe($auth.token);

            if ($auth.token && $auth.me) {
                await goto("/admin/dashboard");
            }
        } catch {
        } finally {
            checkingSession = false;
        }
    }

    onMount(() => {
        void checkAlreadyLoggedIn();
    });

    async function submit() {
        errorMsg = validate();
        if (errorMsg) return;

        loading = true;
        errorMsg = null;

        try {
            const data = await api.post<LoginResponse>("/auth/admin/login", {
                email: email.trim(),
                password
            });

            if (!data?.access_token) {
                errorMsg = "Login gagal: token tidak ditemukan";
                return;
            }

            await auth.setToken(data.access_token);
            await goto("/admin/dashboard");
        } catch (e) {
            if (e instanceof ApiError) errorMsg = e.message || "Login gagal";
            else errorMsg = "Login gagal";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>Admin Login</title>
</svelte:head>

<main
        style="
    min-height: calc(100vh - 80px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    font-family: system-ui, -apple-system, 'Segoe UI', Roboto, Arial, sans-serif;
    background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
  "
>
    <div style="width: 100%; max-width: 440px;">
        <div style="text-align: center; margin-bottom: 24px;">
            <h1
                    style="
          margin: 0;
          font-size: 26px;
          font-weight: 700;
          letter-spacing: -0.5px;
          color: #1a1a1a;
        "
            >
                Admin Login
            </h1>
        </div>

        <form
                on:submit|preventDefault={submit}
                style="
        background: #fff;
        border: 1px solid #e5e5e5;
        border-radius: 20px;
        padding: 28px 24px;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
      "
        >
            <div style="display: grid; gap: 8px; margin-bottom: 16px;">
                <label
                        style="
            font-weight: 600;
            font-size: 13px;
            color: #333;
            letter-spacing: 0.2px;
          "
                >
                    Email
                </label>
                <input
                        bind:value={email}
                        type="email"
                        autocomplete="email"
                        placeholder="admin@kidsplanet.com"
                        disabled={loading || checkingSession}
                        style="
            width: 100%;
            padding: 12px 14px;
            border-radius: 12px;
            border: 1.5px solid #e0e0e0;
            outline: none;
            background: #fafafa;
            font-size: 14px;
            transition: all 0.2s ease;
            box-sizing: border-box;
          "
                        on:focus={(e) => {
            e.currentTarget.style.borderColor = '#111';
            e.currentTarget.style.background = '#fff';
          }}
                        on:blur={(e) => {
            e.currentTarget.style.borderColor = '#e0e0e0';
            e.currentTarget.style.background = '#fafafa';
          }}
                />
            </div>

            <div style="display: grid; gap: 8px; margin-bottom: 16px;">
                <label
                        style="
            font-weight: 600;
            font-size: 13px;
            color: #333;
            letter-spacing: 0.2px;
          "
                >
                    Password
                </label>
                <input
                        bind:value={password}
                        type="password"
                        autocomplete="current-password"
                        placeholder="********"
                        disabled={loading || checkingSession}
                        style="
            width: 100%;
            padding: 12px 14px;
            border-radius: 12px;
            border: 1.5px solid #e0e0e0;
            outline: none;
            background: #fafafa;
            font-size: 14px;
            transition: all 0.2s ease;
            box-sizing: border-box;
          "
                        on:focus={(e) => {
            e.currentTarget.style.borderColor = '#111';
            e.currentTarget.style.background = '#fff';
          }}
                        on:blur={(e) => {
            e.currentTarget.style.borderColor = '#e0e0e0';
            e.currentTarget.style.background = '#fafafa';
          }}
                />
            </div>

            {#if errorMsg}
                <div
                        style="
            margin-bottom: 16px;
            padding: 12px 14px;
            border-radius: 12px;
            background: #fef2f2;
            border: 1.5px solid #fecaca;
            color: #991b1b;
            font-size: 13px;
          "
                >
                    <div style="font-weight: 700; margin-bottom: 4px;">Login gagal</div>
                    <div style="opacity: 0.9; line-height: 1.4;">{errorMsg}</div>
                </div>
            {/if}

            <button
                    type="submit"
                    disabled={loading || checkingSession}
                    style="
          width: 100%;
          padding: 13px 16px;
          border-radius: 12px;
          border: none;
          background: #111;
          color: #fff;
          font-weight: 700;
          font-size: 14px;
          letter-spacing: 0.3px;
          cursor: pointer;
          opacity: ${loading ? '0.7' : '1'};
          transition: all 0.2s ease;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        "
                    on:mouseenter={(e) => {
          if (!loading) e.currentTarget.style.transform = 'translateY(-1px)';
        }}
                    on:mouseleave={(e) => {
          e.currentTarget.style.transform = 'translateY(0)';
        }}
            >
                {checkingSession ? "Checking session..." : loading ? "Signing in..." : "Sign in"}
            </button>
        </form>
    </div>
</main>