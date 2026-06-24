<script lang="ts">
	import { page } from '$app/stores';
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import { api } from '$lib/api';
	import type { Agent } from '$lib/types';
	import { onMount } from 'svelte';
	import { ArrowLeft } from '@lucide/svelte';

	let agent = $state<Agent | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		const aid = $page.params.id!;
		try {
			agent = await api.agents.get(aid);
		} catch (e: unknown) {
			error = (e as Error).message;
		} finally {
			loading = false;
		}
	});
</script>

{#if loading}
	<div class="space-y-4">
		<Skeleton class="h-10 w-64 rounded-md" />
		<Skeleton class="h-40 rounded-xl" />
	</div>
{:else if error}
	<Card class="border-red-500/20 bg-red-500/5 p-6">
		<p class="text-red-400 text-sm font-mono">{error}</p>
	</Card>
{:else if agent}
	<div class="space-y-8 max-w-2xl">
		<div>
			<a href="/{$page.params.workspace}/agents" class="text-xs text-zinc-500 hover:text-zinc-400 font-mono transition-colors inline-flex items-center gap-1">
				<ArrowLeft class="size-3" /> Agents
			</a>
			<h1 class="text-2xl font-bold tracking-tight mt-2">{agent.name}</h1>
		</div>

		<Card class="border-zinc-800 p-6 space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="font-semibold">Agent Details</h3>
				<Badge variant="outline" class={agent.enabled ? 'text-emerald-400 border-emerald-500/30 bg-emerald-500/5' : 'text-zinc-500 border-zinc-700'}>
					{agent.enabled ? 'enabled' : 'disabled'}
				</Badge>
			</div>
			<div class="grid grid-cols-2 gap-3 text-sm">
				<div>
					<p class="text-zinc-500 text-xs font-mono uppercase">Harness</p>
					<p class="font-medium text-zinc-200">{agent.harness}</p>
				</div>
				<div>
					<p class="text-zinc-500 text-xs font-mono uppercase">Model</p>
					<p class="font-medium text-zinc-200">{agent.model || '—'}</p>
				</div>
				<div>
					<p class="text-zinc-500 text-xs font-mono uppercase">Daemon</p>
					<p class="font-medium text-zinc-200 font-mono text-xs">{agent.daemon_id}</p>
				</div>
				<div>
					<p class="text-zinc-500 text-xs font-mono uppercase">Created</p>
					<p class="font-medium text-zinc-200 font-mono text-xs">{agent.created_at ?? '—'}</p>
				</div>
			</div>
		</Card>

	</div>
{/if}
