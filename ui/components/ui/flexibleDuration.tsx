// Flexible duration component with number input + unit selector
// Allows users to set any arbitrary duration (e.g., 4 hours, 90 minutes)

import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import React from "react";

export const durationUnits = [
	{ label: "Minutes", value: "m" },
	{ label: "Hours", value: "h" },
	{ label: "Days", value: "d" },
	{ label: "Weeks", value: "w" },
	{ label: "Months", value: "M" },
	{ label: "Years", value: "y" },
] as const;

export type DurationUnit = (typeof durationUnits)[number]["value"];

export interface FlexibleDuration {
	value: number;
	unit: DurationUnit;
}

// Parse a duration string (e.g., "4h", "90m") into FlexibleDuration
export function parseFlexibleDuration(durationStr: string): FlexibleDuration | null {
	if (!durationStr || durationStr.length < 2) return null;

	const match = durationStr.match(/^(\d+)([mhdwMy])$/);
	if (!match) return null;

	const value = parseInt(match[1], 10);
	const unit = match[2] as DurationUnit;

	return { value, unit };
}

// Convert FlexibleDuration to string format (e.g., "4h")
export function formatFlexibleDuration(duration: FlexibleDuration): string {
	return `${duration.value}${duration.unit}`;
}

// Get human-readable label for a duration (e.g., "4 hours")
export function getDurationLabel(durationStr: string): string {
	const parsed = parseFlexibleDuration(durationStr);
	if (!parsed) return durationStr;

	const unitLabel = durationUnits.find((u) => u.value === parsed.unit)?.label || parsed.unit;
	return `${parsed.value} ${unitLabel.toLowerCase()}${parsed.value !== 1 ? "s" : ""}`;
}

interface FlexibleDurationInputProps {
	id: string;
	label: string;
	value: string; // Duration string format (e.g., "4h", "90m")
	onChange: (value: string) => void;
	minValue?: number;
	maxValue?: number;
	placeholder?: string;
	labelClassName?: string;
	disabled?: boolean;
}

export function FlexibleDurationInput({
	id,
	label,
	value,
	onChange,
	minValue = 1,
	maxValue = 999,
	placeholder = "1",
	labelClassName,
	disabled = false,
}: FlexibleDurationInputProps) {
	const parsed = parseFlexibleDuration(value);
	const numberValue = parsed?.value?.toString() || "";
	const unitValue = parsed?.unit || "h";

	const handleNumberChange = (inputValue: string) => {
		if (inputValue === "") {
			onChange("");
			return;
		}

		const num = parseInt(inputValue, 10);
		if (isNaN(num)) return;

		const clampedNum = Math.max(minValue, Math.min(maxValue, num));
		onChange(`${clampedNum}${unitValue}`);
	};

	const handleUnitChange = (newUnit: string) => {
		const num = parsed?.value || 1;
		onChange(`${num}${newUnit}`);
	};

	return (
		<div className="flex w-full items-start gap-3">
			<div className="flex-1 space-y-2">
				<Label htmlFor={id} className={labelClassName}>
					{label}
				</Label>
				<Input
					id={id}
					placeholder={placeholder}
					value={numberValue}
					onChange={(e) => handleNumberChange(e.target.value)}
					type="number"
					min={minValue}
					max={maxValue}
					disabled={disabled}
					className="w-full"
				/>
			</div>
			<div className="w-32 space-y-2">
				<Label htmlFor={`${id}-unit`} className={labelClassName}>
					Unit
				</Label>
				<Select value={unitValue} onValueChange={handleUnitChange} disabled={disabled}>
					<SelectTrigger id={`${id}-unit`} className="w-full">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{durationUnits.map((unit) => (
							<SelectItem key={unit.value} value={unit.value}>
								{unit.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>
		</div>
	);
}

// Legacy-compatible wrapper that maintains the same interface as NumberAndSelect
// but uses the new flexible duration format
interface FlexibleDurationWrapperProps {
	id: string;
	label: string;
	value: string;
	selectValue: string; // This is the combined duration string now
	onChangeNumber: (value: string) => void;
	onChangeSelect: (value: string) => void;
	options?: { label: string; value: string }[]; // Legacy - ignored
	labelClassName?: string;
	disabled?: boolean;
}

export function FlexibleDurationWrapper({
	id,
	label,
	value,
	selectValue,
	onChangeNumber,
	onChangeSelect,
	labelClassName,
	disabled,
}: FlexibleDurationWrapperProps) {
	// Combine the value and selectValue into a single duration string
	const combinedValue = value && selectValue ? `${value}${selectValue}` : selectValue || "";

	const handleChange = (newValue: string) => {
		const parsed = parseFlexibleDuration(newValue);
		if (parsed) {
			onChangeNumber(parsed.value.toString());
			onChangeSelect(parsed.unit);
		} else {
			onChangeNumber("");
			onChangeSelect("");
		}
	};

	return (
		<FlexibleDurationInput
			id={id}
			label={label}
			value={combinedValue}
			onChange={handleChange}
			labelClassName={labelClassName}
			disabled={disabled}
		/>
	);
}

// DurationNumberInput - Just the number part of a duration input
interface DurationNumberInputProps {
	id: string;
	value: string;
	onChange: (value: string) => void;
	disabled?: boolean;
	minValue?: number;
	maxValue?: number;
}

export function DurationNumberInput({
	id,
	value,
	onChange,
	disabled = false,
	minValue = 1,
	maxValue = 999,
}: DurationNumberInputProps) {
	const parsed = parseFlexibleDuration(value);
	const numberValue = parsed?.value?.toString() || "";
	const unitValue = parsed?.unit || "h";

	const handleNumberChange = (inputValue: string) => {
		if (inputValue === "") {
			onChange("");
			return;
		}

		const num = parseInt(inputValue, 10);
		if (isNaN(num)) return;

		const clampedNum = Math.max(minValue, Math.min(maxValue, num));
		onChange(`${clampedNum}${unitValue}`);
	};

	return (
		<Input
			id={id}
			placeholder="1"
			value={numberValue}
			onChange={(e) => handleNumberChange(e.target.value)}
			type="number"
			min={minValue}
			max={maxValue}
			disabled={disabled}
			className="w-full"
		/>
	);
}

// DurationUnitSelect - Just the unit selector part of a duration input
interface DurationUnitSelectProps {
	id: string;
	value: string;
	onChange: (value: string) => void;
	disabled?: boolean;
}

export function DurationUnitSelect({
	id,
	value,
	onChange,
	disabled = false,
}: DurationUnitSelectProps) {
	const parsed = parseFlexibleDuration(value);
	const unitValue = parsed?.unit || "h";

	const handleUnitChange = (newUnit: string) => {
		const num = parsed?.value || 1;
		onChange(`${num}${newUnit}`);
	};

	return (
		<Select value={unitValue} onValueChange={handleUnitChange} disabled={disabled}>
			<SelectTrigger id={id} className="w-full">
				<SelectValue />
			</SelectTrigger>
			<SelectContent>
				{durationUnits.map((unit) => (
					<SelectItem key={unit.value} value={unit.value}>
						{unit.label}
					</SelectItem>
				))}
			</SelectContent>
		</Select>
	);
}

export default FlexibleDurationInput;
