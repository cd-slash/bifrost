"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { getDurationLabel } from "@/lib/constants/governance";
import { Budget, Customer, ModelConfig, ProviderGovernance, RateLimit, Team, VirtualKey } from "@/lib/types/governance";
import { RefreshCw } from "lucide-react";
import Link from "next/link";

interface LimitItem {
	id: string;
	name: string;
	type: "provider" | "api_key" | "model" | "team" | "customer";
	budget?: Budget;
	rateLimit?: RateLimit;
	url: string;
}

interface BudgetsAndLimitsViewProps {
	providers: ProviderGovernance[];
	virtualKeys: VirtualKey[];
	modelConfigs: ModelConfig[];
	teams: Team[];
	customers: Customer[];
	budgets: Record<string, Budget>;
	rateLimits: Record<string, RateLimit>;
	onRefresh: () => void;
}

export default function BudgetsAndLimitsView({
	providers,
	virtualKeys,
	modelConfigs,
	teams,
	customers,
	budgets,
	rateLimits,
	onRefresh,
}: BudgetsAndLimitsViewProps) {
	// Build a consolidated list of all limit items
	const buildLimitItems = (): LimitItem[] => {
		const items: LimitItem[] = [];

		// Provider limits
		providers.forEach((provider) => {
			if (provider.budget || provider.rate_limit) {
				items.push({
					id: `provider-${provider.provider}`,
					name: provider.provider,
					type: "provider",
					budget: provider.budget,
					rateLimit: provider.rate_limit,
					url: "/workspace/providers",
				});
			}
		});

		// API Key (Virtual Key) limits
		virtualKeys.forEach((vk) => {
			if (vk.budget || vk.rate_limit) {
				items.push({
					id: `vk-${vk.id}`,
					name: vk.name,
					type: "api_key",
					budget: vk.budget,
					rateLimit: vk.rate_limit,
					url: "/workspace/virtual-keys",
				});
			}
		});

		// Model limits
		modelConfigs.forEach((mc) => {
			if (mc.budget || mc.rate_limit) {
				const displayName = mc.provider ? `${mc.model_name} (${mc.provider})` : mc.model_name;
				items.push({
					id: `model-${mc.id}`,
					name: displayName,
					type: "model",
					budget: mc.budget,
					rateLimit: mc.rate_limit,
					url: "/workspace/model-limits",
				});
			}
		});

		// Team limits
		teams.forEach((team) => {
			if (team.budget) {
				items.push({
					id: `team-${team.id}`,
					name: team.name,
					type: "team",
					budget: team.budget,
					url: "/workspace/governance",
				});
			}
		});

		// Customer limits
		customers.forEach((customer) => {
			if (customer.budget) {
				items.push({
					id: `customer-${customer.id}`,
					name: customer.name,
					type: "customer",
					budget: customer.budget,
					url: "/workspace/governance",
				});
			}
		});

		return items;
	};

	const allItems = buildLimitItems();

	// Filter items by type
	const providerItems = allItems.filter((i) => i.type === "provider");
	const apiKeyItems = allItems.filter((i) => i.type === "api_key");
	const modelItems = allItems.filter((i) => i.type === "model");
	const teamItems = allItems.filter((i) => i.type === "team");
	const customerItems = allItems.filter((i) => i.type === "customer");

	// Calculate usage percentage
	const getBudgetUsagePercent = (budget?: Budget): number => {
		if (!budget || budget.max_limit === 0) return 0;
		return Math.min(100, (budget.current_usage / budget.max_limit) * 100);
	};

	const getTokenUsagePercent = (rateLimit?: RateLimit): number => {
		if (!rateLimit || !rateLimit.token_max_limit || rateLimit.token_max_limit === 0) return 0;
		return Math.min(100, (rateLimit.token_current_usage / rateLimit.token_max_limit) * 100);
	};

	const getRequestUsagePercent = (rateLimit?: RateLimit): number => {
		if (!rateLimit || !rateLimit.request_max_limit || rateLimit.request_max_limit === 0) return 0;
		return Math.min(100, (rateLimit.request_current_usage / rateLimit.request_max_limit) * 100);
	};

	// Get color based on usage
	const getUsageColor = (percent: number): string => {
		if (percent >= 100) return "bg-red-500";
		if (percent >= 80) return "bg-amber-500";
		return "bg-green-500";
	};

	// Get type label
	const getTypeLabel = (type: string): string => {
		const labels: Record<string, string> = {
			provider: "Provider",
			api_key: "API Key",
			model: "Model",
			team: "Team",
			customer: "Customer",
		};
		return labels[type] || type;
	};

	// Render limit table
	const renderLimitTable = (items: LimitItem[]) => {
		if (items.length === 0) {
			return (
				<Card>
					<CardContent className="flex h-32 items-center justify-center">
						<p className="text-muted-foreground text-sm">No limits configured</p>
					</CardContent>
				</Card>
			);
		}

		return (
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead className="w-[200px]">Name</TableHead>
						<TableHead>Type</TableHead>
						<TableHead>Budget</TableHead>
						<TableHead>Token Limit</TableHead>
						<TableHead>Request Limit</TableHead>
						<TableHead className="text-right">Actions</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{items.map((item) => {
						const budgetPercent = getBudgetUsagePercent(item.budget);
						const tokenPercent = getTokenUsagePercent(item.rateLimit);
						const requestPercent = getRequestUsagePercent(item.rateLimit);

						return (
							<TableRow key={item.id}>
								<TableCell className="font-medium">{item.name}</TableCell>
								<TableCell>
									<span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-1 text-xs font-medium text-slate-800 dark:bg-slate-800 dark:text-slate-200">
										{getTypeLabel(item.type)}
									</span>
								</TableCell>
								<TableCell>
									{item.budget ? (
										<div className="space-y-1">
											<div className="flex items-center justify-between text-xs">
												<span>
													${item.budget.current_usage.toFixed(2)} / ${item.budget.max_limit.toFixed(2)}
												</span>
												<span className="text-muted-foreground">
													Resets {getDurationLabel(item.budget.reset_duration)}
												</span>
											</div>
											<Progress value={budgetPercent} className="h-2" indicatorClassName={getUsageColor(budgetPercent)} />
											{budgetPercent >= 100 && (
												<span className="text-xs text-red-500 font-medium">Budget exceeded</span>
											)}
										</div>
									) : (
										<span className="text-muted-foreground text-sm">-</span>
									)}
								</TableCell>
								<TableCell>
									{item.rateLimit?.token_max_limit ? (
										<div className="space-y-1">
											<div className="flex items-center justify-between text-xs">
												<span>
													{item.rateLimit.token_current_usage.toLocaleString()} /{" "}
													{item.rateLimit.token_max_limit.toLocaleString()} tokens
												</span>
												<span className="text-muted-foreground">
													Resets {getDurationLabel(item.rateLimit.token_reset_duration || "")}
												</span>
											</div>
											<Progress value={tokenPercent} className="h-2" indicatorClassName={getUsageColor(tokenPercent)} />
											{tokenPercent >= 100 && (
												<span className="text-xs text-red-500 font-medium">Limit exceeded</span>
											)}
										</div>
									) : (
										<span className="text-muted-foreground text-sm">-</span>
									)}
								</TableCell>
								<TableCell>
									{item.rateLimit?.request_max_limit ? (
										<div className="space-y-1">
											<div className="flex items-center justify-between text-xs">
												<span>
													{item.rateLimit.request_current_usage.toLocaleString()} /{" "}
													{item.rateLimit.request_max_limit.toLocaleString()} requests
												</span>
												<span className="text-muted-foreground">
													Resets {getDurationLabel(item.rateLimit.request_reset_duration || "")}
												</span>
											</div>
											<Progress value={requestPercent} className="h-2" indicatorClassName={getUsageColor(requestPercent)} />
											{requestPercent >= 100 && (
												<span className="text-xs text-red-500 font-medium">Limit exceeded</span>
											)}
										</div>
									) : (
										<span className="text-muted-foreground text-sm">-</span>
									)}
								</TableCell>
								<TableCell className="text-right">
									<Button variant="ghost" size="sm" asChild>
										<Link href={item.url}>Manage</Link>
									</Button>
								</TableCell>
							</TableRow>
						);
					})}
				</TableBody>
			</Table>
		);
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-bold tracking-tight">Budgets & Limits</h1>
					<p className="text-muted-foreground">
						Manage and monitor all budgets, rate limits, and quotas across providers, API keys, models, teams, and customers.
					</p>
				</div>
				<Button variant="outline" size="sm" onClick={onRefresh}>
					<RefreshCw className="mr-2 h-4 w-4" />
					Refresh
				</Button>
			</div>

			{/* Summary Cards */}
			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">Providers</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{providerItems.length}</div>
						<p className="text-xs text-muted-foreground">
							{providerItems.filter((i) => i.budget).length} budgets,{" "}
							{providerItems.filter((i) => i.rateLimit).length} rate limits
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">API Keys</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{apiKeyItems.length}</div>
						<p className="text-xs text-muted-foreground">
							{apiKeyItems.filter((i) => i.budget).length} budgets,{" "}
							{apiKeyItems.filter((i) => i.rateLimit).length} rate limits
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">Models</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{modelItems.length}</div>
						<p className="text-xs text-muted-foreground">
							{modelItems.filter((i) => i.budget).length} budgets,{" "}
							{modelItems.filter((i) => i.rateLimit).length} rate limits
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">Teams</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{teamItems.length}</div>
						<p className="text-xs text-muted-foreground">
							{teamItems.filter((i) => i.budget).length} budgets configured
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">Customers</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{customerItems.length}</div>
						<p className="text-xs text-muted-foreground">
							{customerItems.filter((i) => i.budget).length} budgets configured
						</p>
					</CardContent>
				</Card>
			</div>

			{/* Provider Precedence Notice */}
			<Card className="bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800">
				<CardHeader className="pb-2">
					<CardTitle className="text-sm font-medium text-amber-800 dark:text-amber-200">
						Provider Limit Precedence
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="text-sm text-amber-700 dark:text-amber-300">
						Provider-level limits take precedence over API key, model, team, and customer limits. If a provider limit
						is exceeded, all requests to that provider will be blocked regardless of other limits.
					</p>
				</CardContent>
			</Card>

			{/* Tabs for different limit types */}
			<Tabs defaultValue="all" className="w-full">
				<TabsList className="grid w-full grid-cols-6">
					<TabsTrigger value="all">All ({allItems.length})</TabsTrigger>
					<TabsTrigger value="providers">Providers ({providerItems.length})</TabsTrigger>
					<TabsTrigger value="api_keys">API Keys ({apiKeyItems.length})</TabsTrigger>
					<TabsTrigger value="models">Models ({modelItems.length})</TabsTrigger>
					<TabsTrigger value="teams">Teams ({teamItems.length})</TabsTrigger>
					<TabsTrigger value="customers">Customers ({customerItems.length})</TabsTrigger>
				</TabsList>

				<TabsContent value="all" className="mt-4">
					<Card>
						<CardHeader>
							<CardTitle>All Limits</CardTitle>
							<CardDescription>Comprehensive view of all configured budgets and rate limits</CardDescription>
						</CardHeader>
						<CardContent>{renderLimitTable(allItems)}</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="providers" className="mt-4">
					<Card>
						<CardHeader>
							<CardTitle>Provider Limits</CardTitle>
							<CardDescription>Budgets and rate limits configured at the provider level</CardDescription>
						</CardHeader>
						<CardContent>{renderLimitTable(providerItems)}</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="api_keys" className="mt-4">
					<Card>
						<CardHeader>
							<CardTitle>API Key Limits</CardTitle>
							<CardDescription>Budgets and rate limits for individual API keys</CardDescription>
						</CardHeader>
						<CardContent>{renderLimitTable(apiKeyItems)}</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="models" className="mt-4">
					<Card>
						<CardHeader>
							<CardTitle>Model Limits</CardTitle>
							<CardDescription>Budgets and rate limits configured for specific models</CardDescription>
						</CardHeader>
						<CardContent>{renderLimitTable(modelItems)}</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="teams" className="mt-4">
					<Card>
						<CardHeader>
							<CardTitle>Team Budgets</CardTitle>
							<CardDescription>Budgets configured for teams</CardDescription>
						</CardHeader>
						<CardContent>{renderLimitTable(teamItems)}</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="customers" className="mt-4">
					<Card>
						<CardHeader>
							<CardTitle>Customer Budgets</CardTitle>
							<CardDescription>Budgets configured for customers</CardDescription>
						</CardHeader>
						<CardContent>{renderLimitTable(customerItems)}</CardContent>
					</Card>
				</TabsContent>
			</Tabs>
		</div>
	);
}
