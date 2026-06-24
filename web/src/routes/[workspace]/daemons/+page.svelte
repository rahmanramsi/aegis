<script lang="ts">
	import { page } from '$app/stores';
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import { api, getToken } from '$lib/api';
	import type { Daemon } from '$lib/types';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Wrench, Plus, Clock, Copy, Check, Terminal, Key, Globe } from '@lucide/svelte';

	let daemons = $state<Daemon[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let addDialogOpen = $state(false);
	let daemonName = $state('');
	let apiKeyCopied = $state(false);
	let commandCopied = $state(false);

	onMount(() => {
		api.daemons.list()
			.then((d) => { daemons = d; })
			.catch((e: Error) => { error = e.message; })
			.finally(() => { loading = false; });
	});
	function getCommand(): string {
		const key = getToken();
		const name = daemonName.trim() || 'aegis-agent';
		return `AEGIS_API_KEY=${key} AEGIS_DAEMON_NAME=${name} AEGIS_GATEWAY_URL=ws://localhost:8080/ws/daemon ./aegis-agent`;
	}

	async function copyApiKey() {
		try {
			await navigator.clipboard.writeText(getToken());
			apiKeyCopied = true;
		} catch {
			toast.error('Failed to copy');
		}
	}

	async function copyCommand() {
		try {
			await navigator.clipboard.writeText(getCommand());
			commandCopied = true;
		} catch {
			toast.error('Failed to copy');
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<a href="/{$page.params.workspace}" class="text-xs text-zinc-500 hover:text-zinc-400 font-mono transition-colors">&larr; Workspace</a>
			<h1 class="text-2xl font-bold tracking-tight mt-2">Daemons</h1>
		</div>
		<Button variant="outline" size="sm" class="gap-2" onclick={() => { addDialogOpen = true; daemonName = ''; apiKeyCopied = false; commandCopied = false; }}>
			<Plus class="size-4" />
			Add Daemon
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
			<p class="text-zinc-500 font-mono text-sm">No daemons connected</p>
			<p class="text-zinc-600 text-xs mt-1">Run the agent daemon binary to connect</p>
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

<!-- Add Daemon Dialog -->
<Dialog.Root bind:open={addDialogOpen}>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>Add Daemon</Dialog.Title>
			<Dialog.Description>
				Run the agent daemon binary on any machine. It will auto-register and appear here.
			</Dialog.Description>
		</Dialog.Header>
		<div class="py-4 space-y-4">
			<!-- Daemon Name -->
			<div class="space-y-2">
				<Label for="daemon-name">Daemon Name <span class="text-zinc-600 font-normal">(optional)</span></Label>
				<Input id="daemon-name" bind:value={daemonName} placeholder="aegis-agent" />
			</div>

			<!-- API Key -->
			<div class="space-y-2">
				<Label class="flex items-center gap-1.5">
					<Key class="size-3.5" />
					Your API Key
				</Label>
				<div class="flex gap-2">
					<code class="flex-1 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-emerald-400 font-mono break-all">
						{getToken()}
					</code>
					<Button variant="outline" size="sm" class="gap-1.5 shrink-0" onclick={copyApiKey}>
						{#if apiKeyCopied}
							<Check class="size-3.5 text-emerald-400" />
							Copied
						{:else}
							<Copy class="size-3.5" />
							Copy
						{/if}
					</Button>
				</div>
			</div>

			<!-- Run Command -->
			<div class="space-y-2">
				<Label class="flex items-center gap-1.5">
					<Terminal class="size-3.5" />
					Run this command
				</Label>
				<div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
					<code class="text-xs text-zinc-300 font-mono break-all select-all">
						{getCommand()}
					</code>
				</div>
				<Button variant="outline" size="sm" class="gap-1.5 w-full" onclick={copyCommand}>
					{#if commandCopied}
						<Check class="size-3.5 text-emerald-400" />
						Copied
					{:else}
						<Copy class="size-3.5" />
						Copy Command
					{/if}
				</Button>
			</div>

			<div class="flex items-start gap-2 text-xs text-zinc-500 bg-zinc-900 rounded-lg p-3">
				<Globe class="size-3.5 mt-0.5 shrink-0" />
				<span>Make sure the binary is on the target machine and the gateway URL is reachable.</span>
			</div>
		</div>
		<Dialog.Footer>
			<Dialog.Close>
				<Button variant="ghost">Done</Button>
			</Dialog.Close>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
