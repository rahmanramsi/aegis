<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Card from '$lib/components/ui/card/card.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Select from '$lib/components/ui/select';
	import { api } from '$lib/api';
	import type { Daemon, Agent, AgentCreateResponse } from '$lib/types';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Copy, Check } from '@lucide/svelte';

	let daemons = $state<Daemon[]>([]);
	let name = $state('');
	let daemonId = $state('');
	let harness = $state('');
	let model = $state('');
	let telegramToken = $state('');
	let submitting = $state(false);
	let createdAgent = $state<Agent | null>(null);
	let createdToken = $state('');
	let copied = $state(false);

	const selectedDaemon = $derived(daemons.find((d: Daemon) => d.id === daemonId));
	const availableHarnesses = $derived(selectedDaemon?.harnesses ?? []);

	onMount(() => {
		const wid = $page.params.workspace!;
		api.daemons.list(wid)
			.then((d) => { daemons = d; })
			.catch(() => { toast.error('Failed to load daemons'); });
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!name || !daemonId || !harness) return;
		submitting = true;
		try {
			const wid = $page.params.workspace!;
			const res = await api.agents.create(wid, {
				name,
				daemon_id: daemonId,
				harness,
				model: model || undefined,
				telegram_token: telegramToken || undefined,
			});
			createdAgent = res.agent;
			createdToken = res.telegram_token;
			toast.success('Agent created');
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to create agent');
		} finally {
			submitting = false;
		}
	}

	async function copyToken() {
		try {
			await navigator.clipboard.writeText(createdToken);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {}
	}

	function done() {
		goto(`/${$page.params.workspace}/agents`);
	}
</script>

<svelte:head>
	<title>New Agent — Aegis</title>
</svelte:head>

<div class="max-w-lg mx-auto space-y-6">
	<div>
		<h1 class="text-2xl font-bold">Create Agent</h1>
		<p class="text-zinc-400 mt-1">Connect an AI agent to a messaging platform</p>
	</div>

	{#if createdAgent}
		<Card class="border-emerald-800 bg-emerald-950/50 p-6 space-y-4">
			<div class="flex items-center gap-2 text-emerald-400">
				<Check class="h-5 w-5" />
				<span class="font-semibold">Agent "{createdAgent.name}" created</span>
			</div>
			{#if createdToken}
				<div class="space-y-2">
					<p class="text-sm text-zinc-400">Telegram token (save this now — it won't be shown again):</p>
					<div class="flex items-center gap-2">
						<code class="flex-1 bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-300 break-all font-mono">
							{createdToken}
						</code>
						<Button variant="outline" size="sm" onclick={copyToken}>
							{#if copied}
								<Check class="h-4 w-4 text-emerald-400" />
							{:else}
								<Copy class="h-4 w-4" />
							{/if}
						</Button>
					</div>
					<p class="text-xs text-emerald-600">Add this token to AEGIS_TELEGRAM_TOKENS env and restart the gateway.</p>
				</div>
			{/if}
			<Button class="w-full" onclick={done}>Go to agents</Button>
		</Card>
	{:else}
		<Card class="border-zinc-800 p-6">
			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2">
					<Label for="name">Name</Label>
					<Input id="name" placeholder="e.g. Support Bot" bind:value={name} required />
				</div>

				<div class="space-y-2">
					<Label for="daemon">Daemon</Label>
					<Select.Root bind:value={daemonId}>
						<Select.Trigger class="w-full">
							<Select.Value placeholder="Select daemon..." />
						</Select.Trigger>
						<Select.Content>
							{#each daemons as d}
								<Select.Item value={d.id}>{d.name} ({d.status})</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class="space-y-2">
					<Label for="harness">Harness</Label>
					<Select.Root bind:value={harness} disabled={!daemonId}>
						<Select.Trigger class="w-full">
							<Select.Value placeholder="Select harness..." />
						</Select.Trigger>
						<Select.Content>
							{#each availableHarnesses as h}
								<Select.Item value={h}>{h}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class="space-y-2">
					<Label for="model">Model (optional)</Label>
					<Input id="model" placeholder="e.g. claude-sonnet-4-20250514" bind:value={model} />
				</div>

				<div class="space-y-2">
					<Label for="tg-token">Telegram Bot Token (optional)</Label>
					<Input id="tg-token" type="password" placeholder="Bot token from @BotFather" bind:value={telegramToken} />
					<p class="text-xs text-zinc-500">Stored encrypted. Only shown once on creation.</p>
				</div>

				<div class="flex gap-2 justify-end pt-2">
					<Button type="button" variant="outline" href={`/${$page.params.workspace}/agents`}>Cancel</Button>
					<Button type="submit" disabled={submitting}>
						{submitting ? 'Creating...' : 'Create Agent'}
					</Button>
				</div>
			</form>
		</Card>
	{/if}
</div>
