import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import React from "react";
import { RoutingProfileTargetsEditor } from "./routingProfileTargetsEditor";

describe("RoutingProfileTargetsEditor", () => {
	it("adds a new target", async () => {
		const user = userEvent.setup();
		const onChange = vi.fn();

		render(
			<RoutingProfileTargetsEditor
				targets={[{ provider: "anthropic", model: "claude-3-5-haiku-latest", virtual_model: "light", priority: 1, enabled: true }]}
				onChange={onChange}
				idPrefix="test"
			/>
		);

		await user.click(screen.getByRole("button", { name: /add target/i }));

		expect(onChange).toHaveBeenCalledTimes(1);
		const nextTargets = onChange.mock.calls[0][0];
		expect(nextTargets).toHaveLength(2);
		expect(nextTargets[1].provider).toBe("");
		expect(nextTargets[1].enabled).toBe(true);
	});

	it("removes a target", async () => {
		const user = userEvent.setup();
		const onChange = vi.fn();

		render(
			<RoutingProfileTargetsEditor
				targets={[
					{ provider: "anthropic", model: "claude-3-5-haiku-latest", virtual_model: "light", priority: 1, enabled: true },
					{ provider: "cerebras", model: "glm-4.7-flash", priority: 2, enabled: true },
				]}
				onChange={onChange}
				idPrefix="test"
			/>
		);

		const removeButtons = screen.getAllByRole("button", { name: /remove/i });
		await user.click(removeButtons[0]);

		expect(onChange).toHaveBeenCalledTimes(1);
		const nextTargets = onChange.mock.calls[0][0];
		expect(nextTargets).toHaveLength(1);
		expect(nextTargets[0].provider).toBe("cerebras");
	});
});
