import { baseApi } from "./baseApi";
import { GetRoutingProfilesResponse, RoutingProfile } from "@/lib/types/routingProfiles";

export const routingProfilesApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		getRoutingProfiles: builder.query<RoutingProfile[], { virtualProvider?: string } | void>({
			query: (params) => {
				const searchParams = new URLSearchParams();
				if (params?.virtualProvider) {
					searchParams.append("virtual_provider", params.virtualProvider);
				}
				return {
					url: `/governance/routing-profiles${searchParams.toString() ? `?${searchParams.toString()}` : ""}`,
				method: "GET",
				};
			},
			transformResponse: (response: GetRoutingProfilesResponse) => response.profiles || [],
			providesTags: (result, error, params) => ["RoutingProfiles", { type: "RoutingProfiles", id: params?.virtualProvider || "all" }],
		}),
		getRoutingProfile: builder.query<RoutingProfile, string>({
			query: (id) => ({
				url: `/governance/routing-profiles/${id}`,
				method: "GET",
			}),
			transformResponse: (response: { profile: RoutingProfile }) => response.profile,
			providesTags: (result, error, id) => [{ type: "RoutingProfiles", id }],
		}),
		exportRoutingProfiles: builder.query<{ plugin: unknown }, void>({
			query: () => ({
				url: "/governance/routing-profiles/export",
				method: "GET",
			}),
			providesTags: ["RoutingProfiles"],
		}),
		createRoutingProfile: builder.mutation<RoutingProfile, Partial<RoutingProfile>>({
			query: (body) => ({
				url: "/governance/routing-profiles",
				method: "POST",
				body,
			}),
			transformResponse: (response: { profile: RoutingProfile }) => response.profile,
			invalidatesTags: ["RoutingProfiles"],
		}),
		updateRoutingProfile: builder.mutation<void, { id: string; data: Partial<RoutingProfile> }>({
			query: ({ id, data }) => ({
				url: `/governance/routing-profiles/${id}`,
				method: "PUT",
				body: data,
			}),
			invalidatesTags: ["RoutingProfiles"],
		}),
		deleteRoutingProfile: builder.mutation<void, string>({
			query: (id) => ({
				url: `/governance/routing-profiles/${id}`,
				method: "DELETE",
			}),
			invalidatesTags: ["RoutingProfiles"],
		}),
	}),
});

export const {
	useGetRoutingProfilesQuery,
	useGetRoutingProfileQuery,
	useExportRoutingProfilesQuery,
	useCreateRoutingProfileMutation,
	useUpdateRoutingProfileMutation,
	useDeleteRoutingProfileMutation,
} = routingProfilesApi;
