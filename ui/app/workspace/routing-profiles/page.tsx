"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
	useCreateRoutingProfileMutation,
	useDeleteRoutingProfileMutation,
	useExportRoutingProfilesQuery,
	useGetRoutingProfilesQuery,
	useImportRoutingProfilesMutation,
	useSimulateRoutingProfileMutation,
	useUpdateRoutingProfileMutation,
} from "@/lib/store/apis/routingProfilesApi";
import { RoutingProfile, RoutingProfileTarget } from "@/lib/types/routingProfiles";
import { useEffect, useMemo, useState } from "react";
import { RoutingProfileTargetsEditor } from "./components/routingProfileTargetsEditor";

function emptyTarget(): RoutingProfileTarget {
	return { provider: "", model: "", virtual_model: "", priority: 1, enabled: true, capabilities: [], request_types: [] };
}

function defaultCreateProfile(): RoutingProfile {
	return {
		name: "Light Model Alias",
		virtual_provider: "light",
		enabled: true,
		strategy: "ordered_failover",
		targets: [
			{ provider: "anthropic", virtual_model: "light", model: "claude-3-5-haiku-latest", priority: 1, enabled: true, capabilities: ["text"] },
			{ provider: "cerebras", model: "glm-4.7-flash", priority: 2, enabled: true, capabilities: ["text"] },
		],
	};
}

export default function RoutingProfilesPage() {
	const [virtualProviderFilter, setVirtualProviderFilter] = useState<string>("");
	const [showExport, setShowExport] = useState<boolean>(false);
	const { data = [], isLoading, error, isFetching } = useGetRoutingProfilesQuery(
		virtualProviderFilter ? { virtualProvider: virtualProviderFilter } : undefined
	);
	const { data: exportData } = useExportRoutingProfilesQuery(undefined, { skip: !showExport });
	const [createRoutingProfile, { isLoading: isCreating }] = useCreateRoutingProfileMutation();
	const [updateRoutingProfile, { isLoading: isUpdating }] = useUpdateRoutingProfileMutation();
	const [deleteRoutingProfile, { isLoading: isDeleting }] = useDeleteRoutingProfileMutation();
	const [simulateRoutingProfile, { isLoading: isSimulating }] = useSimulateRoutingProfileMutation();
	const [importRoutingProfiles, { isLoading: isImporting }] = useImportRoutingProfilesMutation();

	const [createForm, setCreateForm] = useState<RoutingProfile>(defaultCreateProfile());
	const [profileForms, setProfileForms] = useState<Record<string, RoutingProfile>>({});
	const [simulateDraft, setSimulateDraft] = useState<string>(JSON.stringify({ model: "light/light", request_type: "chat" }, null, 2));
	const [importDraft, setImportDraft] = useState<string>(JSON.stringify({ routing_profiles: [defaultCreateProfile()] }, null, 2));
	const [simulateResult, setSimulateResult] = useState<string>("");
	const [simulateError, setSimulateError] = useState<string>("");
	const [importError, setImportError] = useState<string>("");
	const [createError, setCreateError] = useState<string>("");
	const [rowErrors, setRowErrors] = useState<Record<string, string>>({});

	const sortedProfiles = useMemo(
		() => [...data].sort((a, b) => (a.virtual_provider || "").localeCompare(b.virtual_provider || "")),
		[data]
	);

	useEffect(() => {
		setProfileForms((prev) => {
			const next = { ...prev };
			for (const profile of sortedProfiles) {
				if (!profile.id) {
					continue;
				}
				if (!next[profile.id]) {
					next[profile.id] = JSON.parse(JSON.stringify(profile));
				}
			}
			return next;
		});
	}, [sortedProfiles]);

	const onCreate = async () => {
		setCreateError("");
		try {
			await createRoutingProfile(createForm).unwrap();
			setCreateForm(defaultCreateProfile());
		} catch (err: any) {
			setCreateError(err?.data?.error?.message || err?.message || "Failed to create routing profile");
		}
	};

	const onSave = async (id: string) => {
		const profile = profileForms[id];
		if (!profile) {
			return;
		}
		setRowErrors((prev) => ({ ...prev, [id]: "" }));
		try {
			await updateRoutingProfile({ id, data: profile }).unwrap();
		} catch (err: any) {
			setRowErrors((prev) => ({
				...prev,
				[id]: err?.data?.error?.message || err?.message || "Failed to update routing profile",
			}));
		}
	};

	const onDelete = async (id: string) => {
		setRowErrors((prev) => ({ ...prev, [id]: "" }));
		try {
			await deleteRoutingProfile(id).unwrap();
		} catch (err: any) {
			setRowErrors((prev) => ({
				...prev,
				[id]: err?.data?.error?.message || err?.message || "Failed to delete routing profile",
			}));
		}
	};

	const onSimulate = async () => {
		setSimulateError("");
		setSimulateResult("");
		try {
			const parsed = JSON.parse(simulateDraft) as { model: string; request_type?: string; capabilities?: string[] };
			const response = await simulateRoutingProfile(parsed).unwrap();
			setSimulateResult(JSON.stringify(response, null, 2));
		} catch (err: any) {
			setSimulateError(err?.data?.error?.message || err?.message || "Failed to simulate routing profile");
		}
	};

	const onImport = async () => {
		setImportError("");
		try {
			const parsed = JSON.parse(importDraft) as { routing_profiles?: RoutingProfile[]; plugin?: unknown };
			await importRoutingProfiles(parsed).unwrap();
		} catch (err: any) {
			setImportError(err?.data?.error?.message || err?.message || "Failed to import routing profiles");
		}
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
		<div className="mx-auto w-full max-w-7xl space-y-4">
			<div>
				<h1 className="text-xl font-semibold">Routing Profiles</h1>
				<p className="text-sm text-muted-foreground">Manage virtual provider/model profiles. Use aliases like <code>light/light</code>.</p>
				<div className="mt-3 max-w-sm">
					<Input placeholder="Filter by virtual provider" value={virtualProviderFilter} onChange={(e) => setVirtualProviderFilter(e.target.value)} />
				</div>
			</div>

			<div className="space-y-3 rounded-md border p-4">
				<p className="text-sm font-medium">Create profile</p>
				<div className="grid grid-cols-1 gap-2 md:grid-cols-4">
					<Input value={createForm.name} onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })} placeholder="Profile name" />
					<Input value={createForm.virtual_provider} onChange={(e) => setCreateForm({ ...createForm, virtual_provider: e.target.value })} placeholder="Virtual provider" />
					<Input value={createForm.strategy || "ordered_failover"} onChange={(e) => setCreateForm({ ...createForm, strategy: e.target.value as RoutingProfile["strategy"] })} placeholder="Strategy" />
					<label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={createForm.enabled} onChange={(e) => setCreateForm({ ...createForm, enabled: e.target.checked })} /> Enabled</label>
				</div>
				<div className="space-y-2">
					<p className="text-xs text-muted-foreground">Targets</p>
					<RoutingProfileTargetsEditor
						targets={createForm.targets}
						onChange={(targets) => setCreateForm({ ...createForm, targets })}
						idPrefix="create"
					/>
				</div>
				<div className="flex items-center gap-2">
					<Button onClick={onCreate} disabled={isCreating || isUpdating || isDeleting || isFetching}>{isCreating ? "Creating..." : "Create profile"}</Button>
					{createError ? <span className="text-xs text-destructive">{createError}</span> : null}
				</div>
			</div>

			<div className="space-y-2 rounded-md border p-4">
				<div className="flex items-center gap-2">
					<Button variant="outline" onClick={() => setShowExport((prev) => !prev)}>{showExport ? "Hide export" : "Show export JSON"}</Button>
				</div>
				{showExport ? <Textarea readOnly value={JSON.stringify(exportData || { plugin: {} }, null, 2)} rows={10} className="font-mono text-xs" /> : null}
			</div>

			<div className="space-y-2 rounded-md border p-4">
				<p className="text-sm font-medium">Simulate route (JSON)</p>
				<Textarea value={simulateDraft} onChange={(e) => setSimulateDraft(e.target.value)} rows={6} className="font-mono text-xs" />
				<div className="flex items-center gap-2">
					<Button variant="outline" onClick={onSimulate} disabled={isSimulating}>{isSimulating ? "Simulating..." : "Simulate"}</Button>
					{simulateError ? <span className="text-xs text-destructive">{simulateError}</span> : null}
				</div>
				{simulateResult ? <Textarea readOnly value={simulateResult} rows={10} className="font-mono text-xs" /> : null}
			</div>

			<div className="space-y-2 rounded-md border p-4">
				<p className="text-sm font-medium">Import profiles (JSON)</p>
				<Textarea value={importDraft} onChange={(e) => setImportDraft(e.target.value)} rows={8} className="font-mono text-xs" />
				<div className="flex items-center gap-2">
					<Button variant="outline" onClick={onImport} disabled={isImporting}>{isImporting ? "Importing..." : "Import"}</Button>
					{importError ? <span className="text-xs text-destructive">{importError}</span> : null}
				</div>
			</div>

			<div className="space-y-3">
				{sortedProfiles.length === 0 ? <div className="rounded-md border border-dashed p-6 text-sm text-muted-foreground">No routing profiles yet.</div> : null}
				{sortedProfiles.map((profile) => {
					if (!profile.id) {
						return null;
					}
					const editable = profileForms[profile.id] || profile;
					return (
						<div key={profile.id} className="space-y-2 rounded-md border p-4">
							<div className="grid grid-cols-1 gap-2 md:grid-cols-4">
								<Input value={editable.name} onChange={(e) => setProfileForms((prev) => ({ ...prev, [profile.id!]: { ...editable, name: e.target.value } }))} />
								<Input value={editable.virtual_provider} onChange={(e) => setProfileForms((prev) => ({ ...prev, [profile.id!]: { ...editable, virtual_provider: e.target.value } }))} />
								<Input value={editable.strategy || "ordered_failover"} onChange={(e) => setProfileForms((prev) => ({ ...prev, [profile.id!]: { ...editable, strategy: e.target.value as RoutingProfile["strategy"] } }))} />
								<label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={editable.enabled} onChange={(e) => setProfileForms((prev) => ({ ...prev, [profile.id!]: { ...editable, enabled: e.target.checked } }))} /> Enabled</label>
							</div>
							<RoutingProfileTargetsEditor
								targets={editable.targets}
								onChange={(targets) => setProfileForms((prev) => ({ ...prev, [profile.id!]: { ...editable, targets } }))}
								idPrefix={`edit-${profile.id}`}
							/>
							<div className="flex items-center gap-2">
								<Button variant="outline" onClick={() => onSave(profile.id!)} disabled={isCreating || isUpdating || isDeleting || isFetching}>Save</Button>
								<Button variant="destructive" onClick={() => onDelete(profile.id!)} disabled={isCreating || isUpdating || isDeleting || isFetching}>Delete</Button>
								{rowErrors[profile.id] ? <span className="text-xs text-destructive">{rowErrors[profile.id]}</span> : null}
							</div>
						</div>
					);
				})}
			</div>
		</div>
	);
}
