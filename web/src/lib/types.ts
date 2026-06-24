export interface Workspace {
	id: string;
	name: string;
	slug: string;
	created_at: string;
	updated_at: string;
}

export interface Daemon {
	id: string;
	user_id: string;
	name: string;
	status: 'online' | 'offline';
	harnesses: string[];
	harness_models: Record<string, string[]>;
	last_seen: string;
	created_at: string;
}

export interface Agent {
	id: string;
	workspace_id: string;
	daemon_id: string;
	name: string;
	harness: string;
	model: string;
	enabled: boolean;
	created_at: string;
	updated_at: string;
}


export interface HealthStatus {
	status: string;
}

export interface DaemonCreateResponse {
	daemon: Daemon;
	token: string;
}

export interface AgentCreateResponse {
	agent: Agent;
	telegram_token: string;
}
