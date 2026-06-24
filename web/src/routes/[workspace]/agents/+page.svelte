<script lang="ts">
	import { page } from '$app/stores';
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import { api } from '$lib/api';
	import type { Agent } from '$lib/types';
	import { onMount } from 'svelte';
	import { Bot, Plus } from '@lucide/svelte';

	let agents = $state<Agent[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(() => {
		const wid = $page.params.workspace!;
		api.agents.list(wid)
			.then((a) => { agents = a; })
			.catch((e: Error) => { error = e.message; })
			.finally(() => { loading = false; });
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<a href="/{$page.params.workspace}" class="text-xs text-zinc-500 hover:text-zinc-400 font-mono transition-colors">&larr; Workspace</a>
			<h1 class="text-2xl font-bold tracking-tight mt-2">Agents</h1>
		</div>
		<a href="/{$page.params.workspace}/agents/new">
			<Button variant="outline" size="sm" class="gap-2">
				<Plus class="size-4" />
				New Agent
			</Button>
		</a>
	</div>

	{#if loading}
		<div class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(3) as _}
				<Skeleton class="h-32 rounded-xl" />
			{/each}
		</div>
	{:else if error}
		<Card class="border-red-500/20 bg-red-500/5 p-6">
			<p class="text-red-400 text-sm font-mono">{error}</p>
		</Card>
	{:else if agents.length === 0}
		<Card class="border-zinc-800 p-8 text-center">
			<Bot class="size-8 text-zinc-700 mx-auto mb-3" />
			<p class="text-zinc-500 font-mono text-sm">No agents configured</p>
			<p class="text-zinc-600 text-xs mt-1">Create an agent to start orchestrating</p>
		</Card>
	{:else}
		<div class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
			{#each agents as agent}
				<a href="/{$page.params.workspace}/agents/{agent.id}" class="block group">
					<Card class="border-zinc-800 hover:border-zinc-700 transition-colors p-6 cursor-pointer h-full">
						<div class="flex items-start justify-between mb-3">
							<h3 class="font-semibold text-zinc-100 group-hover:text-emerald-400 transition-colors">{agent.name}</h3>
							<Badge variant="outline" class={agent.enabled ? 'text-emerald-400 border-emerald-500/30 bg-emerald-500/5' : 'text-zinc-500 border-zinc-700'}>
								{agent.enabled ? 'enabled' : 'disabled'}
							</Badge>
						</div>
						<div class="space-y-1 text-xs font-mono">
							<p class="text-zinc-500">harness <span class="text-zinc-300">{agent.harness}</span></p>
							{#if agent.model}
								<p class="text-zinc-500">model <span class="text-zinc-300">{agent.model}</span></p>
							{/if}
						</div>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</div>
