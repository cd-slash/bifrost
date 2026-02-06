import { baseApi } from "./baseApi";
import { GetRoutingProfilesResponse, RoutingProfile } from "@/lib/types/routingProfiles";

export const routingProfilesApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		getRoutingProfiles: builder.query<RoutingProfile[], void>({
			query: () => ({
				url: "/governance/routing-profiles",
				method: "GET",
			}),
			transformResponse: (response: GetRoutingProfilesResponse) => response.profiles || [],
			providesTags: ["RoutingProfiles"],
		}),
	}),
});

export const { useGetRoutingProfilesQuery } = routingProfilesApi;
