<script lang="ts">
	import { page } from '$app/stores';
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import { api } from '$lib/api';
	import type { Workspace, Daemon, Agent } from '$lib/types';
	import { onMount } from 'svelte';
	import { Bot, Wrench } from '@lucide/svelte';

	let workspace = $state<Workspace | null>(null);
	let daemons = $state<Daemon[]>([]);
	let agents = $state<Agent[]>([]);

	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(() => {
		const wid = $page.params.workspace!;
		Promise.all([
			api.workspaces.get(wid),
			api.daemons.list().catch(() => [] as Daemon[]),
			api.agents.list(wid).catch(() => [] as Agent[]),
		]).then(([w, d, a]) => {
			workspace = w;
			daemons = d;
			agents = a;
		}).catch((e: Error) => {
			error = e.message;
		}).finally(() => {
			loading = false;
		});
	});
</script>

{#if loading}
	<div class="space-y-4">
		<Skeleton class="h-10 w-64 rounded-md" />
		<Skeleton class="h-6 w-48 rounded-md" />
		<div class="grid gap-4 grid-cols-1 md:grid-cols-2 mt-6">
			<Skeleton class="h-40 rounded-xl" />
			<Skeleton class="h-40 rounded-xl" />
		</div>
	</div>
{:else if error}
	<Card class="border-red-500/20 bg-red-500/5 p-6">
		<p class="text-red-400 text-sm font-mono">{error}</p>
	</Card>
{:else if workspace}
	<div class="space-y-8">
		<section>
			<a href="/" class="text-xs text-zinc-500 hover:text-zinc-400 font-mono transition-colors">&larr; All Workspaces</a>
			<h1 class="text-2xl font-bold tracking-tight mt-2">{workspace.name}</h1>
			<p class="text-sm text-zinc-500 font-mono mt-1">{workspace.slug}</p>
		</section>

		<div class="grid gap-6 grid-cols-1 md:grid-cols-2">
			<a href="/{$page.params.workspace}/daemons" class="block group">
				<Card class="border-zinc-800 hover:border-zinc-700 transition-colors p-6 cursor-pointer h-full">
					<div class="flex items-center gap-3 mb-3">
						<div class="size-8 rounded-md bg-amber-500/10 flex items-center justify-center">
							<Wrench class="size-4 text-amber-400" />
						</div>
						<h3 class="font-semibold">Daemons</h3>
					</div>
					<p class="text-3xl font-bold text-zinc-100 font-mono">{daemons.length}</p>
					<p class="text-xs text-zinc-500 mt-1">
						{daemons.filter((d: Daemon) => d.status === 'online').length} online
					</p>
				</Card>
			</a>

			<a href="/{$page.params.workspace}/agents" class="block group">
				<Card class="border-zinc-800 hover:border-zinc-700 transition-colors p-6 cursor-pointer h-full">
					<div class="flex items-center gap-3 mb-3">
						<div class="size-8 rounded-md bg-emerald-500/10 flex items-center justify-center">
							<Bot class="size-4 text-emerald-400" />
						</div>
						<h3 class="font-semibold">Agents</h3>
					</div>
					<p class="text-3xl font-bold text-zinc-100 font-mono">{agents.length}</p>
					<p class="text-xs text-zinc-500 mt-1">
						{agents.filter((a: Agent) => a.enabled).length} enabled
					</p>
				</Card>
			</a>
		</div>

		<section>
			<h2 class="text-sm font-semibold text-zinc-400 uppercase tracking-wider mb-4">Recent Daemons</h2>
			{#if daemons.length === 0}
				<Card class="border-zinc-800 p-6 text-center">
					<p class="text-zinc-500 font-mono text-sm">No daemons registered</p>
				</Card>
			{:else}
				<div class="space-y-2">
					{#each daemons.slice(0, 5) as d}
						<Card class="border-zinc-800 p-4 flex items-center justify-between">
							<div>
								<p class="font-medium text-sm">{d.name}</p>
								<p class="text-xs text-zinc-500 font-mono mt-0.5">
									{#each d.harnesses ?? [] as h, i}
										<span class="text-zinc-600">{h}{i < (d.harnesses?.length ?? 0) - 1 ? ', ' : ''}</span>
									{/each}
									{#if !d.harnesses?.length}
										<span class="text-zinc-700">no harnesses</span>
									{/if}
								</p>
							</div>
							<Badge variant="outline" class={d.status === 'online' ? 'text-emerald-400 border-emerald-500/30 bg-emerald-500/5' : 'text-zinc-500 border-zinc-700'}>
								{d.status}
							</Badge>
						</Card>
					{/each}
				</div>
			{/if}
		</section>
	</div>
{/if}
