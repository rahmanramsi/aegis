import type { Workspace, Daemon, DaemonCreateResponse, Agent, AgentCreateResponse, Connection, HealthStatus } from '$lib/types';

const BASE = '/api/v1';
const TOKEN_KEY = 'aegis_api_key';

let _token = '';
let _onUnauthorized: (() => void) | null = null;

export function setToken(token: string) {
	_token = token;
	try {
		if (token) localStorage.setItem(TOKEN_KEY, token);
		else localStorage.removeItem(TOKEN_KEY);
	} catch {}
}

export function getToken(): string {
	if (_token) return _token;
	try {
		_token = localStorage.getItem(TOKEN_KEY) || '';
	} catch {
		_token = '';
	}
	return _token;
}

export function onUnauthorized(handler: () => void) {
	_onUnauthorized = handler;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(options?.headers as Record<string, string>),
	};

	const token = getToken();
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const res = await fetch(BASE + path, {
		...options,
		headers,
	});

	if (res.status === 401 && _onUnauthorized) {
		_onUnauthorized();
	}

	if (!res.ok) {
		let msg = '';
		try {
			const body = await res.json();
			msg = body.error || res.statusText;
		} catch {
			msg = res.statusText || `HTTP ${res.status}`;
		}
		throw new Error(msg || `HTTP ${res.status}`);
	}
	if (res.status === 204) return undefined as unknown as T;
	return res.json() as Promise<T>;
}

export const api = {
	workspaces: {
		list: (): Promise<Workspace[]> => request('/workspaces'),
		create: (data: { name: string; slug: string }): Promise<Workspace> =>
			request('/workspaces', { method: 'POST', body: JSON.stringify(data) }),
		get: (id: string): Promise<Workspace> => request(`/workspaces/${id}`),
		update: (id: string, data: { name: string; slug: string }): Promise<Workspace> =>
			request(`/workspaces/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
		delete: (id: string): Promise<void> => request(`/workspaces/${id}`, { method: 'DELETE' }),
	},
	daemons: {
		list: (wid: string): Promise<Daemon[]> => request(`/workspaces/${wid}/daemons`),
		create: (wid: string, data: { name: string }): Promise<DaemonCreateResponse> =>
			request(`/workspaces/${wid}/daemons`, { method: 'POST', body: JSON.stringify(data) }),
		get: (id: string): Promise<Daemon> => request(`/daemons/${id}`),
		delete: (id: string): Promise<void> => request(`/daemons/${id}`, { method: 'DELETE' }),
	},
	agents: {
		list: (wid: string): Promise<Agent[]> => request(`/workspaces/${wid}/agents`),
		create: (wid: string, data: { name: string; daemon_id: string; harness: string; model?: string; personality?: string; telegram_token?: string }): Promise<AgentCreateResponse> =>
			request(`/workspaces/${wid}/agents`, { method: 'POST', body: JSON.stringify(data) }),
		get: (id: string): Promise<Agent> => request(`/agents/${id}`),
		update: (id: string, data: { name?: string; harness?: string; model?: string; enabled?: boolean }): Promise<Agent> =>
			request(`/agents/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
		delete: (id: string): Promise<void> => request(`/agents/${id}`, { method: 'DELETE' }),
	},
	connections: {
		list: (aid: string): Promise<Connection[]> => request(`/agents/${aid}/connections`),
		create: (aid: string, data: { platform: string; chat_id: string }): Promise<Connection> =>
			request(`/agents/${aid}/connections`, { method: 'POST', body: JSON.stringify(data) }),
		delete: (id: string): Promise<void> => request(`/connections/${id}`, { method: 'DELETE' }),
	},
	health: (): Promise<HealthStatus> => request('/health'),
};
