<script lang="ts">
	import { goto } from '$app/navigation';
	import { setToken } from '$lib/api';
	import * as Card from '$lib/components/ui/card';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import { UserPlus } from '@lucide/svelte';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await fetch('/api/v1/auth/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password }),
			});
			if (!res.ok) {
				const body = await res.json();
				throw new Error(body.error || `Registration failed (${res.status})`);
			}
			const data = await res.json();
			setToken(data.api_key);
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Register — Aegis</title>
</svelte:head>

<div class="flex items-center justify-center min-h-[calc(100vh-10rem)]">
	<Card.Root class="w-full max-w-md border-zinc-800 bg-zinc-900/50">
		<Card.Header>
			<Card.Title class="flex items-center gap-2">
				<UserPlus class="size-5 text-emerald-400" />
				Create account
			</Card.Title>
			<Card.Description>
				Register to start using Aegis.
			</Card.Description>
		</Card.Header>
		<form onsubmit={handleSubmit} novalidate>
			<Card.Content>
				<div class="space-y-4">
					{#if error}
						<p class="text-sm text-red-400 bg-red-400/10 border border-red-400/20 rounded-lg px-3 py-2" role="alert">
							{error}
						</p>
					{/if}
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
						<Label for="password">Password</Label>
						<Input
							id="password"
							type="password"
							bind:value={password}
							placeholder="••••••••"
							required
							minlength={6}
							autocomplete="new-password"
						/>
					</div>
				</div>
			</Card.Content>
			<Card.Footer class="flex-col gap-3">
				<Button type="submit" class="w-full" disabled={loading}>
					{#if loading}
						Registering...
					{:else}
						<UserPlus class="size-4" />
						Register
					{/if}
				</Button>
				<p class="text-sm text-muted-foreground text-center">
					Already have an account?
					<a href="/login" class="text-emerald-400 hover:text-emerald-300 transition-colors">Sign in</a>
				</p>
			</Card.Footer>
		</form>
	</Card.Root>
</div>
