<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Card from '$lib/components/ui/card/card.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Select from '$lib/components/ui/select';
	import { api } from '$lib/api';
	import type { Daemon } from '$lib/types';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let daemons = $state<Daemon[]>([]);
	let name = $state('');
	let daemonId = $state('');
	let harness = $state('');
	let model = $state('');
	let submitting = $state(false);

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
			await api.agents.create(wid, { name, daemon_id: daemonId, harness, model: model || undefined });
			toast.success('Agent created');
			goto(`/${wid}/agents`);
		} catch (err: unknown) {
			toast.error(err instanceof Error ? err.message : 'Failed to create agent');
		} finally {
			submitting = false;
		}
	}
</script>

<div class="max-w-lg mx-auto space-y-6">
	<div>
		<a href="/{$page.params.workspace}/agents" class="text-xs text-zinc-500 hover:text-zinc-400 font-mono transition-colors">&larr; Agents</a>
		<h1 class="text-2xl font-bold tracking-tight mt-2">New Agent</h1>
	</div>

	<Card class="border-zinc-800 p-6">
		<form onsubmit={handleSubmit} class="space-y-5">
			<div class="space-y-2">
				<Label for="name">Name</Label>
				<Input id="name" bind:value={name} placeholder="my-agent" required />
			</div>

			<div class="space-y-2">
				<Label for="daemon">Daemon</Label>
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<Select.Root type="single" bind:value={daemonId}>
					<Select.Trigger class="w-full" id="daemon">
						{selectedDaemon?.name ?? 'Select daemon...'}
					</Select.Trigger>
					<Select.Content>
						{#each daemons as d}
							<Select.Item value={d.id} label={d.name} />
						{/each}
					</Select.Content>
				</Select.Root>
			</div>

			<div class="space-y-2">
				<Label for="harness">Harness</Label>
				<Select.Root type="single" bind:value={harness} disabled={!daemonId}>
					<Select.Trigger class="w-full" id="harness">
						{harness || 'Select harness...'}
					</Select.Trigger>
					<Select.Content>
						{#each availableHarnesses as h}
							<Select.Item value={h} label={h} />
						{/each}
					</Select.Content>
				</Select.Root>
			</div>

			<div class="space-y-2">
				<Label for="model">Model <span class="text-zinc-600">(optional)</span></Label>
				<Input id="model" bind:value={model} placeholder="gpt-4" />
			</div>

			<div class="flex gap-3 pt-2">
				<Button type="submit" disabled={submitting || !name || !daemonId || !harness}>
					{submitting ? 'Creating...' : 'Create Agent'}
				</Button>
				<a href="/{$page.params.workspace}/agents">
					<Button type="button" variant="ghost">Cancel</Button>
				</a>
			</div>
		</form>
	</Card>
</div>
