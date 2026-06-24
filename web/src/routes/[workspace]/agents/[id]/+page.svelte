<script lang="ts">
	import { page } from '$app/stores';
	import Card from '$lib/components/ui/card/card.svelte';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Select from '$lib/components/ui/select';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
	import Separator from '$lib/components/ui/separator/separator.svelte';
	import { api } from '$lib/api';
	import type { Agent, Connection } from '$lib/types';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Trash2, Plus, ArrowLeft } from '@lucide/svelte';

	let agent = $state<Agent | null>(null);
	let connections = $state<Connection[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let newPlatform = $state('telegram');
	let newChatId = $state('');
	let addingConnection = $state(false);
	let deletingMap = $state<Record<string, boolean>>({});

	onMount(() => {
		const aid = $page.params.id!;

		Promise.all([
			api.agents.get(aid),
			api.connections.list(aid).catch(() => [] as Connection[]),
		]).then(([a, c]) => {
			agent = a;
			connections = c;
		}).catch((e: Error) => {
			error = e.message;
		}).finally(() => {
			loading = false;
		});
	});

	async function addConnection(e: SubmitEvent) {
		e.preventDefault();
		if (!newChatId) return;
		addingConnection = true;
		try {
			const conn = await api.connections.create($page.params.id!, { platform: newPlatform, chat_id: newChatId });
			connections = [...connections, conn];
			newChatId = '';
			toast.success('Connection added');
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to add connection');
		} finally {
			addingConnection = false;
		}
	}

	async function deleteConnection(connId: string) {
		deletingMap = { ...deletingMap, [connId]: true };
		try {
			await api.connections.delete(connId);
			connections = connections.filter((c: Connection) => c.id !== connId);
			toast.success('Connection removed');
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to remove connection');
		} finally {
			deletingMap = { ...deletingMap, [connId]: false };
		}
	}
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

		<Separator />

		<section class="space-y-4">
			<h2 class="text-sm font-semibold text-zinc-400 uppercase tracking-wider">Connections</h2>

			<Card class="border-zinc-800 p-6">
				<form onsubmit={addConnection} class="space-y-4">
					<h4 class="text-sm font-medium text-zinc-300">Add Connection</h4>
					<div class="flex gap-3 items-end">
						<div class="space-y-2 flex-1">
							<Label for="platform">Platform</Label>
							<!-- svelte-ignore a11y_no_static_element_interactions -->
							<Select.Root type="single" bind:value={newPlatform}>
								<Select.Trigger class="w-full" id="platform">
									{newPlatform}
								</Select.Trigger>
								<Select.Content>
									<Select.Item value="telegram" label="Telegram" />
									<Select.Item value="discord" label="Discord" />
								</Select.Content>
							</Select.Root>
						</div>
						<div class="space-y-2 flex-1">
							<Label for="chat_id">Chat ID</Label>
							<Input id="chat_id" bind:value={newChatId} placeholder="123456789" required />
						</div>
						<Button type="submit" size="sm" disabled={addingConnection || !newChatId} class="gap-1">
							<Plus class="size-3.5" />
							Add
						</Button>
					</div>
				</form>
			</Card>

			{#if connections.length === 0}
				<Card class="border-zinc-800 p-6 text-center">
					<p class="text-zinc-500 font-mono text-sm">No connections configured</p>
				</Card>
			{:else}
				<div class="space-y-2">
					{#each connections as conn}
						<Card class="border-zinc-800 p-4 flex items-center justify-between">
							<div class="flex items-center gap-3">
								<Badge variant="outline" class="text-zinc-300 border-zinc-700 font-mono text-xs uppercase">
									{conn.platform}
								</Badge>
								<span class="text-sm font-mono text-zinc-400">{conn.chat_id}</span>
							</div>
							<Button
								variant="ghost"
								size="icon-sm"
								disabled={deletingMap[conn.id]}
								onclick={() => deleteConnection(conn.id)}
								class="text-zinc-600 hover:text-red-400"
							>
								<Trash2 class="size-3.5" />
							</Button>
						</Card>
					{/each}
				</div>
			{/if}
		</section>
	</div>
{/if}
