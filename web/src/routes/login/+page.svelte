<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import { setToken } from '$lib/api';
	import { goto } from '$app/navigation';
	import { LogIn } from '@lucide/svelte';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function login(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			const res = await fetch('/api/v1/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password }),
			});

			const data = await res.json();

			if (!res.ok) {
				throw new Error(data.error || data.message || `HTTP ${res.status}`);
			}

			setToken(data.api_key);
			await goto('/');
		} catch (err: any) {
			error = err.message || 'Login failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Sign in — Aegis</title>
</svelte:head>

<div class="flex min-h-[60vh] items-center justify-center">
	<Card.Root class="w-full max-w-md border-zinc-800 bg-zinc-900/80 backdrop-blur-sm">
		<Card.Header>
			<Card.Title class="text-zinc-100">Sign in</Card.Title>
			<Card.Description>
				Enter your email and password to access your account.
			</Card.Description>
		</Card.Header>

		<form onsubmit={login} novalidate>
			<Card.Content class="space-y-4">
				<div class="space-y-2">
					<Label for="email">Email</Label>
					<input
						id="email"
						type="text"
						bind:value={email}
						placeholder="you@example.com"
						required
						autocomplete="email"
						class="flex h-9 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1 text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
					/>
				</div>

				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<Label for="password">Password</Label>
						<a
							href="/forgot-password"
							class="text-xs text-zinc-400 hover:text-zinc-300 transition-colors"
						>
							Forgot password?
						</a>
					</div>
					<Input
						id="password"
						type="password"
						placeholder="••••••••"
						bind:value={password}
						required
						autocomplete="current-password"
					/>
				</div>

				{#if error}
					<p class="text-sm text-red-400">{error}</p>
				{/if}
			</Card.Content>

			<Card.Footer class="flex-col gap-3">
				<Button type="submit" class="w-full" disabled={loading}>
					<LogIn class="mr-2 size-4" />
					{loading ? 'Signing in...' : 'Sign in'}
				</Button>
				<p class="text-sm text-zinc-400 text-center">
					Don't have an account?
					<a href="/register" class="text-emerald-400 hover:text-emerald-300 transition-colors">
						Sign up
					</a>
				</p>
			</Card.Footer>
		</form>
	</Card.Root>
</div>
