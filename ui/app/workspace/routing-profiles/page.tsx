"use client";

import { useGetRoutingProfilesQuery } from "@/lib/store/apis/routingProfilesApi";

export default function RoutingProfilesPage() {
	const { data = [], isLoading, error } = useGetRoutingProfilesQuery();

	if (isLoading) {
		return <div className="mx-auto w-full max-w-7xl text-sm text-muted-foreground">Loading routing profiles...</div>;
	}

	if (error) {
		return (
			<div className="mx-auto w-full max-w-7xl">
				<div className="rounded-md border border-dashed p-6 text-sm text-muted-foreground">
					Routing profiles API is not available yet. Backend CRUD endpoints are planned next.
				</div>
			</div>
		);
	}

	return (
		<div className="mx-auto w-full max-w-7xl space-y-3">
			<h1 className="text-xl font-semibold">Routing Profiles</h1>
			<p className="text-sm text-muted-foreground">Initial scaffold: {data.length} profiles loaded.</p>
		</div>
	);
}
