"use client";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
	useCreateRoutingProfileMutation,
	useDeleteRoutingProfileMutation,
	useGetRoutingProfilesQuery,
	useUpdateRoutingProfileMutation,
} from "@/lib/store/apis/routingProfilesApi";
import { RoutingProfile } from "@/lib/types/routingProfiles";
import { useMemo, useState } from "react";

const defaultProfileDraft = {
	name: "Light Model Alias",
	virtual_provider: "light",
	enabled: true,
	strategy: "ordered_failover",
	targets: [
		{ provider: "cerebras", virtual_model: "light", model: "glm-4.7-flash", priority: 1, enabled: true },
		{ provider: "anthropic", virtual_model: "light", model: "claude-3-5-haiku-latest", priority: 2, enabled: true },
	],
};

function prettyJson(value: unknown): string {
	return JSON.stringify(value, null, 2);
}

export default function RoutingProfilesPage() {
	const { data = [], isLoading, error, isFetching } = useGetRoutingProfilesQuery();
	const [createRoutingProfile, { isLoading: isCreating }] = useCreateRoutingProfileMutation();
	const [updateRoutingProfile, { isLoading: isUpdating }] = useUpdateRoutingProfileMutation();
	const [deleteRoutingProfile, { isLoading: isDeleting }] = useDeleteRoutingProfileMutation();

	const [createDraft, setCreateDraft] = useState<string>(prettyJson(defaultProfileDraft));
	const [createError, setCreateError] = useState<string>("");
	const [rowErrors, setRowErrors] = useState<Record<string, string>>({});
	const [editingRows, setEditingRows] = useState<Record<string, string>>({});

	const sortedProfiles = useMemo(
		() => [...data].sort((a, b) => (a.virtual_provider || "").localeCompare(b.virtual_provider || "")),
		[data]
	);

	const onCreate = async () => {
		setCreateError("");
		try {
			const parsed = JSON.parse(createDraft) as Partial<RoutingProfile>;
			await createRoutingProfile(parsed).unwrap();
			setCreateDraft(prettyJson(defaultProfileDraft));
		} catch (err: any) {
			setCreateError(err?.data?.error?.message || err?.message || "Failed to create routing profile");
		}
	};

	const onSave = async (id: string) => {
		const draft = editingRows[id];
		if (!draft) {
			return;
		}
		setRowErrors((prev) => ({ ...prev, [id]: "" }));
		try {
			const parsed = JSON.parse(draft) as Partial<RoutingProfile>;
			await updateRoutingProfile({ id, data: parsed }).unwrap();
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
				<p className="text-sm text-muted-foreground">
					Manage virtual provider/model profiles. Use aliases like <code>light/light</code>.
				</p>
			</div>

			<div className="space-y-2 rounded-md border p-4">
				<p className="text-sm font-medium">Create profile (JSON)</p>
				<Textarea value={createDraft} onChange={(e) => setCreateDraft(e.target.value)} rows={10} className="font-mono text-xs" />
				<div className="flex items-center gap-2">
					<Button onClick={onCreate} disabled={isCreating || isUpdating || isDeleting || isFetching}>
						{isCreating ? "Creating..." : "Create profile"}
					</Button>
					{createError ? <span className="text-xs text-destructive">{createError}</span> : null}
				</div>
			</div>

			<div className="space-y-3">
				{sortedProfiles.length === 0 ? (
					<div className="rounded-md border border-dashed p-6 text-sm text-muted-foreground">No routing profiles yet.</div>
				) : null}
				{sortedProfiles.map((profile) => {
					const id = profile.id || `${profile.virtual_provider}-${profile.name}`;
					const draft = editingRows[id] ?? prettyJson(profile);
					return (
						<div key={id} className="space-y-2 rounded-md border p-4">
							<div className="flex items-center justify-between gap-2">
								<p className="text-sm font-medium">{profile.name} ({profile.virtual_provider})</p>
								<div className="flex items-center gap-2">
									<Button
										variant="outline"
										onClick={() => onSave(id)}
										disabled={!profile.id || isCreating || isUpdating || isDeleting || isFetching}
									>
										Save
									</Button>
									<Button
										variant="destructive"
										onClick={() => profile.id && onDelete(profile.id)}
										disabled={!profile.id || isCreating || isUpdating || isDeleting || isFetching}
									>
										Delete
									</Button>
								</div>
							</div>
							<Textarea
								value={draft}
								onChange={(e) => setEditingRows((prev) => ({ ...prev, [id]: e.target.value }))}
								rows={10}
								className="font-mono text-xs"
							/>
							{rowErrors[id] ? <p className="text-xs text-destructive">{rowErrors[id]}</p> : null}
						</div>
					);
				})}
			</div>
		</div>
	);
}
