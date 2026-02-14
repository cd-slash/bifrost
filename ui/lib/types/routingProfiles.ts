export type RoutingProfileStrategy = "ordered_failover" | "weighted";

export interface RoutingProfileRateHint {
	request_percent_threshold?: number;
	token_percent_threshold?: number;
	budget_percent_threshold?: number;
}

export interface RoutingProfileTarget {
	provider: string;
	model?: string;
	priority?: number;
	weight?: number;
	request_types?: string[];
	capabilities?: string[];
	enabled: boolean;
	rate_limit?: RoutingProfileRateHint;
}

export interface RoutingProfile {
	id?: string;
	name: string;
	description?: string;
	virtual_provider: string;
	virtual_model?: string;
	enabled: boolean;
	strategy?: RoutingProfileStrategy;
	targets: RoutingProfileTarget[];
}

export interface GetRoutingProfilesResponse {
	profiles: RoutingProfile[];
	count: number;
}
