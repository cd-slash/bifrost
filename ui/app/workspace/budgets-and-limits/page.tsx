"use client";

import FullPageLoader from "@/components/fullPageLoader";
import { getErrorMessage, useGetAllLimitsQuery, useLazyGetCoreConfigQuery } from "@/lib/store";
import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import BudgetsAndLimitsView from "./views/budgetsAndLimitsView";

export default function BudgetsAndLimitsPage() {
	const [governanceEnabled, setGovernanceEnabled] = useState<boolean | null>(null);
	const hasGovernanceAccess = useRbac(RbacResource.Governance, RbacOperation.View);

	const [triggerGetConfig] = useLazyGetCoreConfigQuery();

	// Use regular query with skip and polling
	const {
		data: limitsData,
		error: limitsError,
		isLoading: limitsLoading,
		refetch,
	} = useGetAllLimitsQuery(undefined, {
		skip: !governanceEnabled || !hasGovernanceAccess,
		pollingInterval: 5000,
	});

	const isLoading = limitsLoading || governanceEnabled === null;

	useEffect(() => {
		triggerGetConfig({ fromDB: true })
			.then((res) => {
				if (res.data?.client_config?.enable_governance) {
					setGovernanceEnabled(true);
				} else {
					setGovernanceEnabled(false);
					toast.error("Governance is not enabled. Please enable it in the config.");
				}
			})
			.catch((err) => {
				console.error("Failed to fetch config:", err);
				setGovernanceEnabled(false);
				toast.error(getErrorMessage(err) || "Failed to load configuration");
			});
	}, [triggerGetConfig]);

	// Handle query errors
	useEffect(() => {
		if (limitsError) {
			toast.error(`Failed to load limits data: ${getErrorMessage(limitsError)}`);
		}
	}, [limitsError]);

	if (isLoading) {
		return <FullPageLoader />;
	}

	return (
		<div className="mx-auto w-full max-w-7xl">
			<BudgetsAndLimitsView
				providers={limitsData?.providers || []}
				virtualKeys={limitsData?.virtual_keys || []}
				modelConfigs={limitsData?.model_configs || []}
				teams={limitsData?.teams || []}
				customers={limitsData?.customers || []}
				budgets={limitsData?.budgets || {}}
				rateLimits={limitsData?.rate_limits || {}}
				onRefresh={refetch}
			/>
		</div>
	);
}
