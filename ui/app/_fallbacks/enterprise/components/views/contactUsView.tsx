"use client";

import { cn } from "@/lib/utils";
import { useMemo } from "react";

interface Props {
	className?: string;
	icon: React.ReactNode;
	title: string;
	description: string;
	readmeLink: string;
	align?: "middle" | "top";
}

export default function ContactUsView({ icon, title, description, className, readmeLink, align = "middle" }: Props) {
	const normalizedTitle = useMemo(() => {
		if (title.toLowerCase().startsWith("unlock ")) {
			return `Feature unavailable: ${title.substring(7)}`;
		}
		return title;
	}, [title]);

	void readmeLink;

	const normalizedDescription = description.toLowerCase().includes("enterprise license")
		? "This section is not enabled in this build."
		: description;

	return (
		<div className={cn("flex flex-col items-center gap-4 text-center", align === "middle" ? "justify-center" : "justify-start", className)}>
			<div className="text-muted-foreground">{icon}</div>
			<div className="flex flex-col gap-1">
				<h1 className="text-muted-foreground text-xl font-medium">{normalizedTitle}</h1>
				<div className="text-muted-foreground mt-2 max-w-[600px] text-sm font-normal">{normalizedDescription}</div>
			</div>
		</div>
	);
}
