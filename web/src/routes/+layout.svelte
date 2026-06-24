<script lang="ts">
	import '../app.css';
	import { Toaster } from '$lib/components/ui/sonner';
	import * as Dialog from '$lib/components/ui/dialog';
	import Button from '$lib/components/ui/button/button.svelte';
	import { Terminal, KeyRound } from '@lucide/svelte';
	import { setToken, getToken, onUnauthorized } from '$lib/api';

	let { children } = $props();

	let keyOpen = $state(false);
	let keyInput = $state('');
	let hasKey = $state(false);

	function initKey() {
		hasKey = !!getToken();
	}

	function saveKey() {
		setToken(keyInput.trim());
		hasKey = !!keyInput.trim();
		keyOpen = false;
		keyInput = '';
	}

	function clearKey() {
		setToken('');
		hasKey = false;
		keyOpen = false;
		keyInput = '';
	}

	$effect(() => {
		initKey();
		onUnauthorized(() => {
			keyOpen = true;
		});
	});
</script>

<svelte:head>
	<title>Aegis</title>
</svelte:head>

<div class="min-h-screen bg-zinc-950 text-zinc-100 flex flex-col">
	<header class="border-b border-zinc-800 bg-zinc-900/50 backdrop-blur-sm sticky top-0 z-50">
		<div class="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
			<a href="/" class="flex items-center gap-2 font-mono text-emerald-400 hover:text-emerald-300 transition-colors">
				<Terminal class="size-5" />
				<span class="font-bold tracking-tight">aegis</span>
			</a>
			<nav class="flex items-center gap-1 text-sm">
				<a href="/" class="px-3 py-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 transition-colors">
					Workspaces
				</a>
				<button
					onclick={() => { keyOpen = true; keyInput = getToken(); }}
					class="ml-2 px-2 py-1.5 rounded-md {hasKey ? 'text-emerald-400' : 'text-zinc-500'} hover:text-zinc-100 hover:bg-zinc-800/50 transition-colors"
					title="API Key"
				>
					<KeyRound class="size-4" />
				</button>
			</nav>
		</div>
	</header>

	<main class="flex-1 max-w-7xl mx-auto w-full px-4 py-6">
		{@render children()}
	</main>

	<footer class="border-t border-zinc-800 py-3 text-center">
		<p class="text-xs text-zinc-600 font-mono">aegis v0.1.0</p>
	</footer>
</div>

<!-- API Key Dialog -->
<Dialog.Root bind:open={keyOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>API Key</Dialog.Title>
			<Dialog.Description>
				Enter the gateway API key to authorize mutating operations. Leave empty if no key is configured.
			</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-4">
			<input
				type="password"
				bind:value={keyInput}
				placeholder="Enter API key..."
				class="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none focus:border-emerald-500/50"
				onkeydown={(e) => { if (e.key === 'Enter') saveKey(); }}
			/>
		</div>
		<Dialog.Footer class="flex gap-2">
			<Button variant="outline" onclick={clearKey}>Clear</Button>
			<Button onclick={saveKey}>Save</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<Toaster />
