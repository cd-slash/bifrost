"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	Plus,
	Search,
	ChevronUp,
	ChevronDown,
	Trash2,
	Play,
	Edit2,
	Copy,
} from "lucide-react";
import { useState, useMemo } from "react";
import {
	useCreateRoutingProfileMutation,
	useDeleteRoutingProfileMutation,
	useExportRoutingProfilesQuery,
	useGetRoutingProfilesQuery,
	useSimulateRoutingProfileMutation,
	useUpdateRoutingProfileMutation,
} from "@/lib/store/apis/routingProfilesApi";
import { RoutingProfile, RoutingProfileTarget, RoutingProfileStrategy } from "@/lib/types/routingProfiles";
import { toast } from "sonner";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";

const STRATEGY_OPTIONS: { value: RoutingProfileStrategy; label: string; description: string }[] = [
	{ value: "ordered_failover", label: "Ordered Failover", description: "Use targets in order until one succeeds" },
	{ value: "weighted", label: "Weighted", description: "Distribute requests based on weight" },
];

function emptyTarget(): RoutingProfileTarget {
	return { provider: "", model: "", virtual_model: "", priority: 1, enabled: true, capabilities: [], request_types: [] };
}

interface ProfileFormData {
	name: string;
	virtual_provider: string;
	strategy: RoutingProfileStrategy;
	enabled: boolean;
	targets: RoutingProfileTarget[];
}

function emptyForm(): ProfileFormData {
	return {
		name: "",
		virtual_provider: "",
		strategy: "ordered_failover",
		enabled: true,
		targets: [emptyTarget()],
	};
}

export default function RoutingProfilesPage() {
	const [virtualProviderFilter, setVirtualProviderFilter] = useState<string>("");
	const [createSheetOpen, setCreateSheetOpen] = useState(false);
	const [editSheetOpen, setEditSheetOpen] = useState(false);
	const [detailSheetOpen, setDetailSheetOpen] = useState(false);
	const [selectedProfileId, setSelectedProfileId] = useState<string | null>(null);
	const [editingProfile, setEditingProfile] = useState<RoutingProfile | null>(null);
	const [formData, setFormData] = useState<ProfileFormData>(emptyForm());

	const { data = [], isLoading, error } = useGetRoutingProfilesQuery(
		virtualProviderFilter ? { virtualProvider: virtualProviderFilter } : undefined
	);
	const { data: exportData } = useExportRoutingProfilesQuery(undefined, { skip: !detailSheetOpen });
	const [createRoutingProfile, { isLoading: isCreating }] = useCreateRoutingProfileMutation();
	const [updateRoutingProfile, { isLoading: isUpdating }] = useUpdateRoutingProfileMutation();
	const [deleteRoutingProfile] = useDeleteRoutingProfileMutation();
	const [simulateRoutingProfile, { isLoading: isSimulating }] = useSimulateRoutingProfileMutation();

	const [simulateDraft, setSimulateDraft] = useState<string>(JSON.stringify({ model: "light/light", request_type: "chat" }, null, 2));
	const [simulateResult, setSimulateResult] = useState<string>("");
	const [simulateError, setSimulateError] = useState<string>("");

	const sortedProfiles = useMemo(
		() => [...data].sort((a, b) => (a.virtual_provider || "").localeCompare(b.virtual_provider || "")),
		[data]
	);

	const selectedProfile = useMemo(
		() => sortedProfiles.find((p) => p.id === selectedProfileId) || null,
		[sortedProfiles, selectedProfileId]
	);

	const handleCreateOpen = () => {
		setFormData(emptyForm());
		setCreateSheetOpen(true);
	};

	const handleEdit = (profile: RoutingProfile) => {
		setEditingProfile(profile);
		setFormData({
			name: profile.name,
			virtual_provider: profile.virtual_provider,
			strategy: profile.strategy || "ordered_failover",
			enabled: profile.enabled,
			targets: profile.targets.length > 0 ? profile.targets.map(t => ({...t})) : [emptyTarget()],
		});
		setDetailSheetOpen(false);
		setEditSheetOpen(true);
	};

	const handleViewDetails = (profileId: string) => {
		setSelectedProfileId(profileId);
		setSimulateResult("");
		setSimulateError("");
		setDetailSheetOpen(true);
	};

	const handleSaveCreate = async () => {
		if (!formData.name.trim()) {
			toast.error("Profile name is required");
			return;
		}
		if (!formData.virtual_provider.trim()) {
			toast.error("Virtual provider is required");
			return;
		}
		try {
			await createRoutingProfile(formData).unwrap();
			toast.success("Profile created successfully");
			setCreateSheetOpen(false);
			setFormData(emptyForm());
		} catch (err: any) {
			toast.error(err?.data?.error?.message || err?.message || "Failed to create profile");
		}
	};

	const handleSaveEdit = async () => {
		if (!editingProfile?.id) return;
		if (!formData.name.trim()) {
			toast.error("Profile name is required");
			return;
		}
		if (!formData.virtual_provider.trim()) {
			toast.error("Virtual provider is required");
			return;
		}
		try {
			await updateRoutingProfile({ id: editingProfile.id, data: formData }).unwrap();
			toast.success("Profile updated successfully");
			setEditSheetOpen(false);
			setEditingProfile(null);
		} catch (err: any) {
			toast.error(err?.data?.error?.message || err?.message || "Failed to update profile");
		}
	};

	const handleDelete = async (profileId: string) => {
		if (!confirm("Are you sure you want to delete this profile?")) return;
		try {
			await deleteRoutingProfile(profileId).unwrap();
			toast.success("Profile deleted successfully");
			if (selectedProfileId === profileId) {
				setDetailSheetOpen(false);
				setSelectedProfileId(null);
			}
		} catch (err: any) {
			toast.error(err?.data?.error?.message || err?.message || "Failed to delete profile");
		}
	};

	const handleSimulate = async () => {
		setSimulateError("");
		setSimulateResult("");
		try {
			const parsed = JSON.parse(simulateDraft) as { model: string; request_type?: string; capabilities?: string[] };
			const response = await simulateRoutingProfile(parsed).unwrap();
			setSimulateResult(JSON.stringify(response, null, 2));
		} catch (err: any) {
			setSimulateError(err?.data?.error?.message || err?.message || "Failed to simulate routing");
		}
	};

	const moveTarget = (index: number, direction: 'up' | 'down') => {
		const newTargets = [...formData.targets];
		const newIndex = direction === 'up' ? index - 1 : index + 1;
		if (newIndex < 0 || newIndex >= newTargets.length) return;
		[newTargets[index], newTargets[newIndex]] = [newTargets[newIndex], newTargets[index]];
		newTargets.forEach((t, i) => { t.priority = i + 1; });
		setFormData({ ...formData, targets: newTargets });
	};

	const updateTarget = (index: number, field: keyof RoutingProfileTarget, value: any) => {
		const newTargets = [...formData.targets];
		newTargets[index] = { ...newTargets[index], [field]: value };
		setFormData({ ...formData, targets: newTargets });
	};

	const addTarget = () => {
		setFormData({
			...formData,
			targets: [...formData.targets, { ...emptyTarget(), priority: formData.targets.length + 1 }],
		});
	};

	const removeTarget = (index: number) => {
		const newTargets = formData.targets.filter((_, i) => i !== index);
		newTargets.forEach((t, i) => { t.priority = i + 1; });
		setFormData({ ...formData, targets: newTargets });
	};

	const copyToClipboard = (text: string) => {
		navigator.clipboard.writeText(text);
		toast.success("Copied to clipboard");
	};

	if (isLoading) {
		return <div className="mx-auto w-full max-w-7xl text-sm text-muted-foreground">Loading routing profiles...</div>;
	}

	if (error) {
		return (
			<div className="mx-auto w-full max-w-7xl">
				<div className="rounded-md border border-dashed p-6 text-sm text-muted-foreground">Failed to load routing profiles.</div>
			</div>
		);
	}

	return (
		<div className="mx-auto w-full max-w-7xl space-y-6">
			<div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
				<div>
					<h1 className="text-2xl font-semibold">Routing Profiles</h1>
					<p className="text-sm text-muted-foreground">Manage virtual provider/model profiles for intelligent request routing.</p>
				</div>
				<Button onClick={handleCreateOpen} className="gap-2">
					<Plus className="h-4 w-4" />
					Create Profile
				</Button>
			</div>

			<div className="flex items-center gap-2">
				<div className="relative flex-1 max-w-sm">
					<Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						placeholder="Filter by virtual provider..."
						value={virtualProviderFilter}
						onChange={(e) => setVirtualProviderFilter(e.target.value)}
						className="pl-9"
					/>
				</div>
			</div>

			{sortedProfiles.length === 0 ? (
				<Card>
					<CardContent className="flex flex-col items-center justify-center py-12">
						<p className="text-muted-foreground mb-4">No routing profiles yet.</p>
						<Button onClick={handleCreateOpen} variant="outline">
							<Plus className="mr-2 h-4 w-4" />
							Create your first profile
						</Button>
					</CardContent>
				</Card>
			) : (
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{sortedProfiles.map((profile) => (
						<Card key={profile.id} className="cursor-pointer transition-colors hover:bg-accent/50" onClick={() => handleViewDetails(profile.id!)}>
							<CardHeader className="pb-3">
								<div className="flex items-start justify-between">
									<div className="space-y-1">
										<CardTitle className="text-base">{profile.name}</CardTitle>
										<CardDescription className="font-mono text-xs">{profile.virtual_provider}</CardDescription>
									</div>
									<Badge variant={profile.enabled ? "default" : "secondary"}>
										{profile.enabled ? "Active" : "Disabled"}
									</Badge>
								</div>
							</CardHeader>
							<CardContent>
								<div className="flex items-center gap-4 text-sm text-muted-foreground">
									<div className="flex items-center gap-1">
										<span className="font-medium">{profile.targets.length}</span>
										<span>targets</span>
									</div>
									<Separator orientation="vertical" className="h-4" />
									<div className="capitalize">{profile.strategy?.replace("_", " ") || "ordered failover"}</div>
								</div>
								{profile.targets.length > 0 && (
									<div className="mt-3 flex flex-wrap gap-1">
										{profile.targets.slice(0, 3).map((t) => (
											<Badge key={`${t.provider}-${t.model}`} variant="outline" className="text-xs">
												{t.provider}/{t.model || "any"}
											</Badge>
										))}
										{profile.targets.length > 3 && (
											<Badge variant="outline" className="text-xs">
												+{profile.targets.length - 3} more
											</Badge>
										)}
									</div>
								)}
							</CardContent>
						</Card>
					))}
				</div>
			)}

			<Sheet open={createSheetOpen} onOpenChange={setCreateSheetOpen}>
				<SheetContent className="flex w-full flex-col min-w-[500px] overflow-y-auto">
					<SheetHeader>
						<SheetTitle>Create New Profile</SheetTitle>
						<SheetDescription>Define a routing profile with targets in priority order.</SheetDescription>
					</SheetHeader>
					<div className="flex flex-1 flex-col gap-6 py-4">
						<div className="space-y-2">
							<Label htmlFor="create-name">Profile Name</Label>
							<Input
								id="create-name"
								placeholder="e.g., Light Model Alias"
								value={formData.name}
								onChange={(e) => setFormData({ ...formData, name: e.target.value })}
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="create-vp">Virtual Provider</Label>
							<Input
								id="create-vp"
								placeholder="e.g., light"
								value={formData.virtual_provider}
								onChange={(e) => setFormData({ ...formData, virtual_provider: e.target.value })}
							/>
							<p className="text-xs text-muted-foreground">Use aliases like <code>light/light</code> in your requests.</p>
						</div>
						<div className="space-y-2">
							<Label htmlFor="create-strategy">Routing Strategy</Label>
							<Select
								value={formData.strategy}
								onValueChange={(v) => setFormData({ ...formData, strategy: v as RoutingProfileStrategy })}
							>
								<SelectTrigger id="create-strategy">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{STRATEGY_OPTIONS.map((opt) => (
										<SelectItem key={opt.value} value={opt.value}>
											<div className="flex flex-col">
												<span>{opt.label}</span>
												<span className="text-xs text-muted-foreground">{opt.description}</span>
											</div>
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="flex items-center justify-between rounded-lg border p-4">
							<div className="space-y-0.5">
								<Label>Enable Profile</Label>
								<p className="text-sm text-muted-foreground">Active profiles will route requests</p>
							</div>
							<Switch
								checked={formData.enabled}
								onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
							/>
						</div>
						<Separator />
						<div className="space-y-3">
							<div className="flex items-center justify-between">
								<Label>Targets (in priority order)</Label>
								<Button variant="outline" size="sm" onClick={addTarget} className="gap-1">
									<Plus className="h-3 w-3" />
									Add
								</Button>
							</div>
							<div className="space-y-2">
								{formData.targets.map((target, idx) => (
									<div key={`create-target-${idx}-${target.provider}`} className="flex items-center gap-2 rounded-md border p-2">
										<div className="flex flex-col gap-1">
											<Button
												variant="ghost"
												size="sm"
												className="h-6 px-1"
												disabled={idx === 0}
												onClick={() => moveTarget(idx, 'up')}
											>
												<ChevronUp className="h-3 w-3" />
											</Button>
											<Button
												variant="ghost"
												size="sm"
												className="h-6 px-1"
												disabled={idx === formData.targets.length - 1}
												onClick={() => moveTarget(idx, 'down')}
											>
												<ChevronDown className="h-3 w-3" />
											</Button>
										</div>
										<div className="flex flex-1 gap-2">
											<Input
												placeholder="Provider"
												value={target.provider}
												onChange={(e) => updateTarget(idx, 'provider', e.target.value)}
												className="flex-1"
											/>
											<Input
												placeholder="Model (optional)"
												value={target.model || ""}
												onChange={(e) => updateTarget(idx, 'model', e.target.value)}
												className="flex-1"
											/>
											<Input
												placeholder="Virtual model"
												value={target.virtual_model || ""}
												onChange={(e) => updateTarget(idx, 'virtual_model', e.target.value)}
												className="flex-1"
											/>
										</div>
										<Button variant="ghost" size="sm" className="h-8 px-2 text-destructive" onClick={() => removeTarget(idx)}>
											<Trash2 className="h-4 w-4" />
										</Button>
									</div>
								))}
							</div>
						</div>
						<div className="flex justify-end gap-2 pt-4">
							<Button variant="outline" onClick={() => setCreateSheetOpen(false)}>Cancel</Button>
							<Button onClick={handleSaveCreate} disabled={isCreating}>Create Profile</Button>
						</div>
					</div>
				</SheetContent>
			</Sheet>

			<Sheet open={editSheetOpen} onOpenChange={setEditSheetOpen}>
				<SheetContent className="flex w-full flex-col min-w-[500px] overflow-y-auto">
					<SheetHeader>
						<SheetTitle>Edit Profile</SheetTitle>
						<SheetDescription>Update the routing profile configuration.</SheetDescription>
					</SheetHeader>
					<div className="flex flex-1 flex-col gap-6 py-4">
						<div className="space-y-2">
							<Label htmlFor="edit-name">Profile Name</Label>
							<Input
								id="edit-name"
								placeholder="e.g., Light Model Alias"
								value={formData.name}
								onChange={(e) => setFormData({ ...formData, name: e.target.value })}
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="edit-vp">Virtual Provider</Label>
							<Input
								id="edit-vp"
								placeholder="e.g., light"
								value={formData.virtual_provider}
								onChange={(e) => setFormData({ ...formData, virtual_provider: e.target.value })}
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="edit-strategy">Routing Strategy</Label>
							<Select
								value={formData.strategy}
								onValueChange={(v) => setFormData({ ...formData, strategy: v as RoutingProfileStrategy })}
							>
								<SelectTrigger id="edit-strategy">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{STRATEGY_OPTIONS.map((opt) => (
										<SelectItem key={opt.value} value={opt.value}>
											<div className="flex flex-col">
												<span>{opt.label}</span>
												<span className="text-xs text-muted-foreground">{opt.description}</span>
											</div>
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
						<div className="flex items-center justify-between rounded-lg border p-4">
							<div className="space-y-0.5">
								<Label>Enable Profile</Label>
								<p className="text-sm text-muted-foreground">Active profiles will route requests</p>
							</div>
							<Switch
								checked={formData.enabled}
								onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
							/>
						</div>
						<Separator />
						<div className="space-y-3">
							<div className="flex items-center justify-between">
								<Label>Targets (in priority order)</Label>
								<Button variant="outline" size="sm" onClick={addTarget} className="gap-1">
									<Plus className="h-3 w-3" />
									Add
								</Button>
							</div>
							<div className="space-y-2">
								{formData.targets.map((target, idx) => (
									<div key={`edit-target-${idx}-${target.provider}`} className="flex items-center gap-2 rounded-md border p-2">
										<div className="flex flex-col gap-1">
											<Button
												variant="ghost"
												size="sm"
												className="h-6 px-1"
												disabled={idx === 0}
												onClick={() => moveTarget(idx, 'up')}
											>
												<ChevronUp className="h-3 w-3" />
											</Button>
											<Button
												variant="ghost"
												size="sm"
												className="h-6 px-1"
												disabled={idx === formData.targets.length - 1}
												onClick={() => moveTarget(idx, 'down')}
											>
												<ChevronDown className="h-3 w-3" />
											</Button>
										</div>
										<div className="flex flex-1 gap-2">
											<Input
												placeholder="Provider"
												value={target.provider}
												onChange={(e) => updateTarget(idx, 'provider', e.target.value)}
												className="flex-1"
											/>
											<Input
												placeholder="Model (optional)"
												value={target.model || ""}
												onChange={(e) => updateTarget(idx, 'model', e.target.value)}
												className="flex-1"
											/>
											<Input
												placeholder="Virtual model"
												value={target.virtual_model || ""}
												onChange={(e) => updateTarget(idx, 'virtual_model', e.target.value)}
												className="flex-1"
											/>
										</div>
										<Button variant="ghost" size="sm" className="h-8 px-2 text-destructive" onClick={() => removeTarget(idx)}>
											<Trash2 className="h-4 w-4" />
										</Button>
									</div>
								))}
							</div>
						</div>
						<div className="flex justify-end gap-2 pt-4">
							<Button variant="outline" onClick={() => setEditSheetOpen(false)}>Cancel</Button>
							<Button onClick={handleSaveEdit} disabled={isUpdating}>Save Changes</Button>
						</div>
					</div>
				</SheetContent>
			</Sheet>

			<Sheet open={detailSheetOpen} onOpenChange={setDetailSheetOpen}>
				<SheetContent className="flex w-full flex-col min-w-[600px] overflow-y-auto">
					{selectedProfile && (
						<>
							<SheetHeader>
								<div className="flex items-center justify-between">
									<div>
										<SheetTitle>{selectedProfile.name}</SheetTitle>
										<SheetDescription className="font-mono">{selectedProfile.virtual_provider}</SheetDescription>
									</div>
									<Badge variant={selectedProfile.enabled ? "default" : "secondary"}>
										{selectedProfile.enabled ? "Active" : "Disabled"}
									</Badge>
								</div>
							</SheetHeader>
							<div className="flex flex-1 flex-col gap-4 py-4">
								<div className="flex gap-2">
									<Button variant="outline" onClick={() => handleEdit(selectedProfile)} className="gap-2">
										<Edit2 className="h-4 w-4" />
										Edit
									</Button>
									<Button variant="destructive" onClick={() => handleDelete(selectedProfile.id!)} className="gap-2">
										<Trash2 className="h-4 w-4" />
										Delete
									</Button>
								</div>
								<Tabs defaultValue="targets" className="flex flex-1 flex-col">
									<TabsList>
										<TabsTrigger value="targets">Targets</TabsTrigger>
										<TabsTrigger value="simulate">Simulate</TabsTrigger>
										<TabsTrigger value="export">Export JSON</TabsTrigger>
									</TabsList>
									<TabsContent value="targets" className="flex-1 space-y-4">
										<div className="rounded-md border">
											<div className="grid grid-cols-6 gap-2 bg-muted p-2 text-sm font-medium">
												<div>#</div>
												<div className="col-span-2">Provider</div>
												<div>Model</div>
												<div>Virtual Model</div>
												<div>Enabled</div>
											</div>
											{selectedProfile.targets.map((target, idx) => (
												<div key={`detail-target-${idx}-${target.provider}`} className="grid grid-cols-6 gap-2 border-t p-2 text-sm">
													<div className="flex items-center">
														<Badge variant="outline">{idx + 1}</Badge>
													</div>
													<div className="col-span-2 flex items-center font-mono">{target.provider}</div>
													<div className="flex items-center font-mono text-muted-foreground">{target.model || "—"}</div>
													<div className="flex items-center font-mono text-muted-foreground">{target.virtual_model || "—"}</div>
													<div className="flex items-center">
														<Badge variant={target.enabled ? "default" : "secondary"}>
															{target.enabled ? "Yes" : "No"}
														</Badge>
													</div>
												</div>
											))}
											{selectedProfile.targets.length === 0 && (
												<div className="p-4 text-center text-muted-foreground">No targets defined</div>
											)}
										</div>
									</TabsContent>
									<TabsContent value="simulate" className="space-y-4">
										<div className="space-y-2">
											<Label>Request (JSON)</Label>
											<Textarea
												value={simulateDraft}
												onChange={(e) => setSimulateDraft(e.target.value)}
												rows={6}
												className="font-mono text-xs"
											/>
											<Button onClick={handleSimulate} disabled={isSimulating} className="gap-2">
												<Play className="h-4 w-4" />
												{isSimulating ? "Simulating..." : "Run Simulation"}
											</Button>
											{simulateError && (
												<div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
													{simulateError}
												</div>
											)}
											{simulateResult && (
												<div className="space-y-2">
													<div className="flex items-center justify-between">
														<Label>Result</Label>
														<Button variant="ghost" size="sm" onClick={() => copyToClipboard(simulateResult)}>
															<Copy className="h-4 w-4" />
														</Button>
													</div>
													<Textarea readOnly value={simulateResult} rows={12} className="font-mono text-xs" />
												</div>
											)}
										</div>
									</TabsContent>
									<TabsContent value="export" className="space-y-4">
										<div className="space-y-2">
											<div className="flex items-center justify-between">
												<Label>Profile JSON</Label>
												<Button variant="ghost" size="sm" onClick={() => copyToClipboard(JSON.stringify(exportData || {}, null, 2))}>
													<Copy className="mr-2 h-4 w-4" />
													Copy
												</Button>
											</div>
											<Textarea
												readOnly
												value={JSON.stringify(exportData || {}, null, 2)}
												rows={20}
												className="font-mono text-xs"
											/>
										</div>
									</TabsContent>
								</Tabs>
							</div>
						</>
					)}
				</SheetContent>
			</Sheet>
		</div>
	);
}
