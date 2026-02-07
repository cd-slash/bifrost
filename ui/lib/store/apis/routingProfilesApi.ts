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
		getRoutingProfile: builder.query<RoutingProfile, string>({
			query: (id) => ({
				url: `/governance/routing-profiles/${id}`,
				method: "GET",
			}),
			transformResponse: (response: { profile: RoutingProfile }) => response.profile,
			providesTags: (result, error, id) => [{ type: "RoutingProfiles", id }],
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
	useCreateRoutingProfileMutation,
	useUpdateRoutingProfileMutation,
	useDeleteRoutingProfileMutation,
} = routingProfilesApi;
