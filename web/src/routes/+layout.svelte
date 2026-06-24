<script lang="ts">
	import '../app.css';
	import { Toaster } from '$lib/components/ui/sonner';
	import { Terminal, LogOut } from '@lucide/svelte';
	import { setToken, getToken, onUnauthorized } from '$lib/api';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	let { children } = $props();

	const PROTECTED_ROUTES = ['/login', '/register'];

	function isAuthRoute(): boolean {
		return PROTECTED_ROUTES.includes($page.url.pathname);
	}

	function logout() {
		setToken('');
		goto('/login');
	}

	$effect(() => {
		if (!getToken() && !isAuthRoute()) {
			goto('/login');
		}
	});

	$effect(() => {
		onUnauthorized(() => {
			setToken('');
			goto('/login');
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
				{#if getToken()}
					<button
						onclick={logout}
						class="ml-2 px-2 py-1.5 rounded-md text-zinc-500 hover:text-red-400 hover:bg-zinc-800/50 transition-colors"
						title="Logout"
					>
						<LogOut class="size-4" />
					</button>
				{/if}
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


<Toaster />
