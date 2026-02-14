// Governance-related constants

// Legacy duration options - maintained for backward compatibility
// These will be gradually deprecated in favor of flexible duration input
export const resetDurationOptions = [
	{ label: "Every Minute", value: "1m" },
	{ label: "Every 5 Minutes", value: "5m" },
	{ label: "Every 15 Minutes", value: "15m" },
	{ label: "Every 30 Minutes", value: "30m" },
	{ label: "Hourly", value: "1h" },
	{ label: "Every 6 Hours", value: "6h" },
	{ label: "Daily", value: "1d" },
	{ label: "Weekly", value: "1w" },
	{ label: "Monthly", value: "1M" },
];

export const budgetDurationOptions = [
	{ label: "Hourly", value: "1h" },
	{ label: "Daily", value: "1d" },
	{ label: "Weekly", value: "1w" },
	{ label: "Monthly", value: "1M" },
];

// Map of duration values to short labels for display
export const resetDurationLabels: Record<string, string> = {
	"1m": "Every Minute",
	"5m": "Every 5 Minutes",
	"15m": "Every 15 Minutes",
	"30m": "Every 30 Minutes",
	"1h": "Hourly",
	"6h": "Every 6 Hours",
	"1d": "Daily",
	"1w": "Weekly",
	"1M": "Monthly",
};

// Duration units for flexible duration input
export const durationUnits = [
	{ label: "Minutes", value: "m", seconds: 60 },
	{ label: "Hours", value: "h", seconds: 3600 },
	{ label: "Days", value: "d", seconds: 86400 },
	{ label: "Weeks", value: "w", seconds: 604800 },
	{ label: "Months", value: "M", seconds: 2592000 }, // 30 days
	{ label: "Years", value: "y", seconds: 31536000 }, // 365 days
] as const;

export type DurationUnit = (typeof durationUnits)[number]["value"];

export interface FlexibleDuration {
	value: number;
	unit: DurationUnit;
}

/**
 * Parse a duration string into a FlexibleDuration object
 * Supports format: "<number><unit>" where unit is m, h, d, w, M, or y
 * Examples: "4h", "90m", "7d", "1w", "3M", "2y"
 */
export function parseFlexibleDuration(durationStr: string): FlexibleDuration | null {
	if (!durationStr || typeof durationStr !== "string" || durationStr.length < 2) {
		return null;
	}

	const match = durationStr.match(/^(\d+)([mhdwMy])$/);
	if (!match) return null;

	const value = parseInt(match[1], 10);
	const unit = match[2] as DurationUnit;

	// Validate unit is valid
	const validUnit = durationUnits.find((u) => u.value === unit);
	if (!validUnit) return null;

	return { value, unit };
}

/**
 * Convert a FlexibleDuration to a duration string
 * Example: { value: 4, unit: "h" } => "4h"
 */
export function formatFlexibleDuration(duration: FlexibleDuration): string {
	if (!duration || typeof duration.value !== "number" || !duration.unit) {
		return "";
	}
	return `${duration.value}${duration.unit}`;
}

/**
 * Get a human-readable label for a duration string
 * Examples:
 *   "4h" => "4 hours"
 *   "1m" => "1 minute"
 *   "90m" => "90 minutes"
 */
export function getDurationLabel(durationStr: string): string {
	if (!durationStr) return "";

	const parsed = parseFlexibleDuration(durationStr);
	if (!parsed) return durationStr;

	const unitInfo = durationUnits.find((u) => u.value === parsed.unit);
	if (!unitInfo) return durationStr;

	const unitLabel = unitInfo.label.toLowerCase();
	const pluralSuffix = parsed.value !== 1 ? "s" : "";
	return `${parsed.value} ${unitLabel}${pluralSuffix}`;
}

/**
 * Convert a duration string to seconds for calculations
 */
export function durationToSeconds(durationStr: string): number | null {
	const parsed = parseFlexibleDuration(durationStr);
	if (!parsed) return null;

	const unitInfo = durationUnits.find((u) => u.value === parsed.unit);
	if (!unitInfo) return null;

	return parsed.value * unitInfo.seconds;
}

/**
 * Get default duration strings for common use cases
 */
export function getDefaultDurations(): Record<string, string> {
	return {
		minute: "1m",
		fiveMinutes: "5m",
		fifteenMinutes: "15m",
		thirtyMinutes: "30m",
		hour: "1h",
		sixHours: "6h",
		day: "1d",
		week: "1w",
		month: "1M",
		year: "1y",
	};
}

/**
 * Validate that a duration string is in the correct format
 */
export function isValidDuration(durationStr: string): boolean {
	return parseFlexibleDuration(durationStr) !== null;
}
