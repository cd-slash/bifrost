"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { RoutingProfileTarget } from "@/lib/types/routingProfiles";
import React from "react";

function splitCSV(value: string): string[] {
	return value
		.split(",")
		.map((v) => v.trim())
		.filter(Boolean);
}

function joinCSV(value?: string[]): string {
	return (value || []).join(", ");
}

function emptyTarget(): RoutingProfileTarget {
	return { provider: "", model: "", virtual_model: "", priority: 1, enabled: true, capabilities: [], request_types: [] };
}

interface Props {
	targets: RoutingProfileTarget[];
	onChange: (targets: RoutingProfileTarget[]) => void;
	idPrefix: string;
}

export function RoutingProfileTargetsEditor({ targets, onChange, idPrefix }: Props) {
	const updateTarget = (index: number, patch: Partial<RoutingProfileTarget>) => {
		const nextTargets = [...targets];
		nextTargets[index] = { ...nextTargets[index], ...patch };
		onChange(nextTargets);
	};

	return (
		<div className="space-y-2">
			{targets.map((target, idx) => (
				<div key={`${idPrefix}-${target.provider}-${target.model}-${target.virtual_model}-${idx}`} className="grid grid-cols-1 gap-2 rounded border p-2 md:grid-cols-7">
					<Input value={target.provider} onChange={(e) => updateTarget(idx, { provider: e.target.value })} placeholder="Provider" aria-label="Provider" />
					<Input value={target.virtual_model || ""} onChange={(e) => updateTarget(idx, { virtual_model: e.target.value })} placeholder="Virtual model" aria-label="Virtual model" />
					<Input value={target.model || ""} onChange={(e) => updateTarget(idx, { model: e.target.value })} placeholder="Model" aria-label="Model" />
					<Input value={String(target.priority || 0)} onChange={(e) => updateTarget(idx, { priority: Number(e.target.value || "0") })} placeholder="Priority" aria-label="Priority" />
					<Input value={joinCSV(target.capabilities)} onChange={(e) => updateTarget(idx, { capabilities: splitCSV(e.target.value) })} placeholder="Capabilities csv" aria-label="Capabilities" />
					<Input value={joinCSV(target.request_types)} onChange={(e) => updateTarget(idx, { request_types: splitCSV(e.target.value) })} placeholder="Request types csv" aria-label="Request types" />
					<Button variant="outline" onClick={() => onChange(targets.filter((_, t) => t !== idx))}>Remove</Button>
				</div>
			))}
			<Button variant="outline" onClick={() => onChange([...targets, emptyTarget()])}>Add target</Button>
		</div>
	);
}
