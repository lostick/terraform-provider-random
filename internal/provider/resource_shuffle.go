package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

var _ tfsdk.ResourceType = (*shuffleResourceType)(nil)

type shuffleResourceType struct{}

func (r *shuffleResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "The resource `random_shuffle` generates a random permutation of a list of strings " +
			"given as an argument.",
		Attributes: map[string]tfsdk.Attribute{
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"seed": {
				Description: "Arbitrary string with which to seed the random number generator, in order to " +
					"produce less-volatile permutations of the list.\n" +
					"\n" +
					"**Important:** Even with an identical seed, it is not guaranteed that the same permutation " +
					"will be produced across different versions of Terraform. This argument causes the " +
					"result to be *less volatile*, but not fixed for all time.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"input": {
				Description: "The list of strings to shuffle.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"result_count": {
				Description: "The number of results to return. Defaults to the number of items in the " +
					"`input` list. If fewer items are requested, some elements will be excluded from the " +
					"result. If more items are requested, items will be repeated in the result but not more " +
					"frequently than the number of items in the input list.",
				Type:     types.Int64Type,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"result": {
				Description: "Random permutation of the list of strings given in `input`.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Computed: true,
			},
			"id": {
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (r *shuffleResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return &shuffleResource{}, nil
}

var _ tfsdk.Resource = (*shuffleResource)(nil)

type shuffleResource struct{}

func (r *shuffleResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan shuffleModelV0
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := plan.Input
	seed := plan.Seed.Value
	resultCount := plan.ResultCount.Value

	if resultCount == 0 {
		resultCount = int64(len(input.Elems))
	}

	result := make([]attr.Value, 0, resultCount)

	if len(input.Elems) > 0 {
		rand := random.NewRand(seed)

		// Keep producing permutations until we fill our result
	Batches:
		for {
			perm := rand.Perm(len(input.Elems))

			for _, i := range perm {
				result = append(result, input.Elems[i])

				if int64(len(result)) >= resultCount {
					break Batches
				}
			}
		}
	}

	s := shuffleModelV0{
		ID:      types.String{Value: "-"},
		Keepers: plan.Keepers,
		Input:   plan.Input,
		Result: types.List{
			Unknown:  false,
			Null:     false,
			Elems:    result,
			ElemType: types.StringType,
		},
	}

	if plan.Seed.Null {
		s.Seed.Null = true
	} else {
		s.Seed.Value = seed
	}

	if plan.ResultCount.Null {
		s.ResultCount.Null = true
	} else {
		s.ResultCount.Value = resultCount
	}

	diags = resp.State.Set(ctx, s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *shuffleResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update is intentionally left blank as all required and optional attributes force replacement of the resource
// through the RequiresReplace AttributePlanModifier.
func (r *shuffleResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *shuffleResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

type shuffleModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Keepers     types.Map    `tfsdk:"keepers"`
	Seed        types.String `tfsdk:"seed"`
	Input       types.List   `tfsdk:"input"`
	ResultCount types.Int64  `tfsdk:"result_count"`
	Result      types.List   `tfsdk:"result"`
}
