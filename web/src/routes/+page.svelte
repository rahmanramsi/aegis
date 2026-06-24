<script lang="ts">
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import Separator from '$lib/components/ui/separator/separator.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { api } from '$lib/api';
	import type { Workspace, HealthStatus } from '$lib/types';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { Activity, Layers, Plus } from '@lucide/svelte';

	let health = $state<HealthStatus | null>(null);
	let workspaces = $state<Workspace[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let createOpen = $state(false);
	let wsName = $state('');
	let wsSlug = $state('');

	onMount(() => {
		loadData();
	});

	async function loadData() {
		loading = true;
		try {
			[health, workspaces] = await Promise.all([
				api.health().catch(() => null),
				api.workspaces.list().catch(() => []),
			]);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function slugify(name: string): string {
		return name.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
	}

	async function createWorkspace(e: SubmitEvent) {
		e.preventDefault();
		if (!wsName) return;
		try {
			const ws = await api.workspaces.create({ name: wsName, slug: wsSlug || slugify(wsName) });
			workspaces = [...workspaces, ws];
			wsName = '';
			wsSlug = '';
			createOpen = false;
			toast.success('Workspace created');
		} catch (err: any) {
			toast.error(err.message || 'Failed to create workspace');
		}
	}
</script>

<div class="space-y-8">
	<section>
		<div class="flex items-center gap-3 mb-4">
			<div class="size-10 rounded-lg bg-emerald-500/10 flex items-center justify-center">
				<Activity class="size-5 text-emerald-400" />
			</div>
			<div>
				<h1 class="text-2xl font-bold text-zinc-100">Aegis</h1>
				<p class="text-sm text-zinc-500">Your agents, your harness, your rules.</p>
			</div>
		</div>

		<div class="flex items-center gap-2 mt-4">
			<Badge variant={health ? 'default' : 'secondary'}>
				{health ? 'Gateway online' : 'Gateway offline'}
			</Badge>
			<span class="text-xs text-zinc-600">v0.1.0</span>
		</div>
	</section>

	<Separator />

	<section>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-sm font-semibold text-zinc-400 uppercase tracking-wider">Workspaces</h2>
			<Button variant="outline" size="sm" onclick={() => createOpen = true}>
				<Plus class="size-4" data-icon="inline-start" />
				New Workspace
			</Button>
		</div>

		{#if loading}
			<div class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
				{#each Array(3) as _}
					<Card class="p-6"><Skeleton class="h-6 w-2/3 mb-2" /><Skeleton class="h-4 w-1/2" /></Card>
				{/each}
			</div>
		{:else if error}
			<Card class="border-red-500/20 bg-red-500/5 p-6">
				<p class="text-red-400 text-sm font-mono">{error}</p>
			</Card>
		{:else if workspaces.length === 0}
			<Card class="border-zinc-800 p-8 text-center">
				<Layers class="size-8 text-zinc-700 mx-auto mb-3" />
				<p class="text-zinc-400 text-sm">No workspaces yet</p>
				<p class="text-zinc-600 text-xs mt-1">Create your first workspace to get started.</p>
			</Card>
		{:else}
			<div class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
				{#each workspaces as ws}
					<button class="text-left cursor-pointer" onclick={() => goto(`/${ws.id}`)}>
						<Card class="p-6 border-zinc-800 hover:border-zinc-700 transition-colors">
							<div class="flex items-center gap-2 mb-1">
								<Layers class="size-4 text-zinc-500" />
								<h3 class="font-semibold text-zinc-100">{ws.name}</h3>
							</div>
							<p class="text-xs text-zinc-500 font-mono">{ws.slug}</p>
						</Card>
					</button>
				{/each}
			</div>
		{/if}
	</section>
</div>

<!-- Create Workspace Dialog -->
<Dialog.Root bind:open={createOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Create Workspace</Dialog.Title>
			<Dialog.Description>Workspaces isolate your agents and configurations.</Dialog.Description>
		</Dialog.Header>
		<form onsubmit={createWorkspace} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="ws-name">Name</Label>
				<Input id="ws-name" bind:value={wsName} placeholder="My Workspace" required
					oninput={() => wsSlug = slugify(wsName)} />
			</div>
			<div class="flex flex-col gap-2">
				<Label for="ws-slug">Slug</Label>
				<Input id="ws-slug" bind:value={wsSlug} placeholder="my-workspace" />
				<p class="text-xs text-zinc-500">Used in URLs. Auto-generated from name.</p>
			</div>
			<Dialog.Footer>
				<Button type="button" variant="outline" onclick={() => createOpen = false}>Cancel</Button>
				<Button type="submit" disabled={!wsName}>Create</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
