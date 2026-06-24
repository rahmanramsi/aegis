<script lang="ts">
	import { page } from '$app/stores';
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import { api } from '$lib/api';
	import type { Daemon, DaemonCreateResponse } from '$lib/types';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Wrench, Plus, Clock, Copy, Check } from '@lucide/svelte';

	let daemons = $state<Daemon[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let newDaemonName = $state('');
	let registerDialogOpen = $state(false);
	let registering = $state(false);

	let tokenValue = $state('');
	let tokenDialogOpen = $state(false);
	let tokenCopied = $state(false);

	onMount(() => {
		api.daemons.list()
			.then((d) => { daemons = d; })
			.catch((e: Error) => { error = e.message; })
			.finally(() => { loading = false; });
	});

	async function registerDaemon(e: SubmitEvent) {
		e.preventDefault();
		if (!newDaemonName) return;
		registering = true;
		try {
			const resp = await api.daemons.create({ name: newDaemonName });
			daemons = [...daemons, resp.daemon];
			tokenValue = resp.token;
			newDaemonName = '';
			registerDialogOpen = false;
			tokenDialogOpen = true;
			tokenCopied = false;
			toast.success('Daemon registered');
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to register daemon');
		} finally {
			registering = false;
		}
	}

	async function copyToken() {
		try {
			await navigator.clipboard.writeText(tokenValue);
			tokenCopied = true;
		} catch {
			// clipboard unavailable
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<a href="/{$page.params.workspace}" class="text-xs text-zinc-500 hover:text-zinc-400 font-mono transition-colors">&larr; Workspace</a>
			<h1 class="text-2xl font-bold tracking-tight mt-2">Daemons</h1>
		</div>
		<Button variant="outline" size="sm" class="gap-2" onclick={() => { registerDialogOpen = true; newDaemonName = ''; }}>
			<Plus class="size-4" />
			Register Daemon
		</Button>
	</div>

	{#if loading}
		<div class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(3) as _}
				<Skeleton class="h-36 rounded-xl" />
			{/each}
		</div>
	{:else if error}
		<Card class="border-red-500/20 bg-red-500/5 p-6">
			<p class="text-red-400 text-sm font-mono">{error}</p>
		</Card>
	{:else if daemons.length === 0}
		<Card class="border-zinc-800 p-8 text-center">
			<Wrench class="size-8 text-zinc-700 mx-auto mb-3" />
			<p class="text-zinc-500 font-mono text-sm">No daemons registered</p>
			<p class="text-zinc-600 text-xs mt-1">Register a daemon to start running agents</p>
		</Card>
	{:else}
		<div class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
			{#each daemons as d}
				<Card class="border-zinc-800 p-6 space-y-3">
					<div class="flex items-start justify-between">
						<h3 class="font-semibold">{d.name}</h3>
						<Badge variant="outline" class={d.status === 'online' ? 'text-emerald-400 border-emerald-500/30 bg-emerald-500/5' : 'text-zinc-500 border-zinc-700'}>
							{d.status}
						</Badge>
					</div>
					<div class="flex items-center gap-1.5 text-xs text-zinc-500">
						<Clock class="size-3" />
						<span class="font-mono">{d.last_seen ?? 'never'}</span>
					</div>
					{#if d.harnesses?.length}
						<div class="flex flex-wrap gap-1.5 pt-1">
							{#each d.harnesses as h}
								<Badge variant="outline" class="text-zinc-500 border-zinc-700 text-xs font-mono">
									{h}
								</Badge>
							{/each}
						</div>
					{/if}
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Register Daemon Dialog -->
<Dialog.Root bind:open={registerDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Register Daemon</Dialog.Title>
			<Dialog.Description>Create a new daemon. A token will be shown once after creation.</Dialog.Description>
		</Dialog.Header>
		<form onsubmit={registerDaemon} class="space-y-4 py-4">
			<div class="space-y-2">
				<Label for="daemon-name">Name</Label>
				<Input id="daemon-name" bind:value={newDaemonName} placeholder="my-daemon" required />
			</div>
			<Dialog.Footer>
				<Dialog.Close>
					<Button type="button" variant="ghost">Cancel</Button>
				</Dialog.Close>
				<Button type="submit" disabled={registering || !newDaemonName}>
					{registering ? 'Registering...' : 'Register'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Token Dialog -->
<Dialog.Root bind:open={tokenDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Daemon Token</Dialog.Title>
			<Dialog.Description>
				Copy this token now. It will not be shown again.
			</Dialog.Description>
		</Dialog.Header>
		<div class="py-4 space-y-4">
			<div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
				<code class="text-xs text-emerald-400 font-mono break-all">{tokenValue}</code>
			</div>
			<Button variant="outline" size="sm" class="gap-2 w-full" onclick={copyToken}>
				{#if tokenCopied}
					<Check class="size-4 text-emerald-400" />
					Copied
				{:else}
					<Copy class="size-4" />
					Copy Token
				{/if}
			</Button>
		</div>
		<Dialog.Footer>
			<Dialog.Close>
				<Button variant="ghost">Close</Button>
			</Dialog.Close>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
