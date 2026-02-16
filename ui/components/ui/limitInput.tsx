// LimitInput component - Separates limit value from reset duration
// Allows setting arbitrary limit values (e.g., 1,000,000 tokens) with flexible reset periods (e.g., 5 hours)

import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { DurationNumberInput, DurationUnitSelect } from "./flexibleDuration";
import React from "react";

interface LimitInputProps {
	id: string;
	label: string;
	limitValue: string;
	durationValue: string;
	onChangeLimit: (value: string) => void;
	onChangeDuration: (value: string) => void;
	placeholder?: string;
	labelClassName?: string;
	disabled?: boolean;
	// Type of limit for appropriate placeholder
	limitType?: "tokens" | "requests" | "currency";
}

export function LimitInput({
	id,
	label,
	limitValue,
	durationValue,
	onChangeLimit,
	onChangeDuration,
	placeholder = "Enter amount...",
	labelClassName,
	disabled = false,
	limitType = "tokens",
}: LimitInputProps) {
	// Get appropriate placeholder based on limit type
	const getPlaceholder = () => {
		switch (limitType) {
			case "tokens":
				return "e.g., 1000000";
			case "requests":
				return "e.g., 1000";
			case "currency":
				return "e.g., 100.00";
			default:
				return placeholder;
		}
	};

	// Get helper text based on limit type
	const getHelperText = () => {
		switch (limitType) {
			case "tokens":
				return "Maximum number of tokens allowed";
			case "requests":
				return "Maximum number of requests allowed";
			case "currency":
				return "Maximum spend in USD";
			default:
				return "";
		}
	};

	return (
		<div className="space-y-4">
			{/* Row 1: All Labels aligned at top */}
			<div className="flex w-full gap-3">
				{/* Limit Label - 1/2 width */}
				<div className="w-1/2">
					<Label htmlFor={id} className={labelClassName}>
						{label}
					</Label>
				</div>
				{/* Reset Period Label - 1/4 width */}
				<div className="w-1/4">
					<Label htmlFor={`${id}-duration`} className="text-xs font-normal">
						Reset Period
					</Label>
				</div>
				{/* Unit Label - 1/4 width */}
				<div className="w-1/4">
					<Label htmlFor={`${id}-unit`} className="text-xs font-normal">
						Unit
					</Label>
				</div>
			</div>
			
			{/* Row 2: All Inputs aligned */}
			<div className="flex w-full items-start gap-3">
				{/* Limit Value Input - 1/2 width */}
				<div className="w-1/2 space-y-1">
					<Input
						id={id}
						type="text"
						inputMode="numeric"
						placeholder={getPlaceholder()}
						value={limitValue}
						onChange={(e) => {
							// Allow only numeric input
							const value = e.target.value.replace(/[^0-9.]/g, "");
							onChangeLimit(value);
						}}
						disabled={disabled}
						className="w-full"
					/>
					{getHelperText() && (
						<p className="text-muted-foreground text-xs">{getHelperText()}</p>
					)}
				</div>

				{/* Reset Duration Input - 1/4 width */}
				<div className="w-1/4">
					<DurationNumberInput
						id={`${id}-duration`}
						value={durationValue}
						onChange={onChangeDuration}
						disabled={disabled}
					/>
				</div>

				{/* Unit Selector - 1/4 width */}
				<div className="w-1/4">
					<DurationUnitSelect
						id={`${id}-unit`}
						value={durationValue}
						onChange={onChangeDuration}
						disabled={disabled}
					/>
				</div>
			</div>
		</div>
	);
}

export default LimitInput;
