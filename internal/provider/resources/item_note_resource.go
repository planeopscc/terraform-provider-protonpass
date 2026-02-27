// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

var _ resource.Resource = &ItemNoteResource{}
var _ resource.ResourceWithImportState = &ItemNoteResource{}

type ItemNoteResource struct {
	client *passcli.Client
}

type ItemNoteResourceModel struct {
	ItemID     types.String `tfsdk:"item_id"`
	ShareID    types.String `tfsdk:"share_id"`
	Title      types.String `tfsdk:"title"`
	NoteWO     types.String `tfsdk:"note_wo"`
	NoteWOVer  types.Int64  `tfsdk:"note_wo_version"`
	CreateTime types.String `tfsdk:"create_time"`
	ModifyTime types.String `tfsdk:"modify_time"`
}

func NewItemNoteResource() resource.Resource {
	return &ItemNoteResource{}
}

func (r *ItemNoteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_note"
}

func (r *ItemNoteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Proton Pass note item.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the item.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"share_id": schema.StringAttribute{
				MarkdownDescription: "Share ID of the vault.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the note.",
				Required:            true,
			},
			"note_wo": schema.StringAttribute{
				MarkdownDescription: "Note content (write-only, never stored in state).",
				Optional:            true,
				WriteOnly:           true,
			},
			"note_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Increment to trigger note content update.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"create_time": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"modify_time": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ItemNoteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*passcli.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *passcli.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ItemNoteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ItemNoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var noteContent string
	var noteWO types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("note_wo"), &noteWO)...)
	if !noteWO.IsNull() && !noteWO.IsUnknown() {
		noteContent = noteWO.ValueString()
	}

	tflog.Debug(ctx, "creating item_note", map[string]interface{}{"title": data.Title.ValueString()})

	item, err := r.client.CreateItemNote(ctx, data.ShareID.ValueString(), data.Title.ValueString(), noteContent)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create note item", err.Error())
		return
	}

	data.ItemID = types.StringValue(item.ItemID)
	data.ShareID = types.StringValue(item.ShareID)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	if data.NoteWOVer.IsNull() || data.NoteWOVer.IsUnknown() {
		data.NoteWOVer = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemNoteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ItemNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	item, err := r.client.ReadItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		if passcli.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read note item", err.Error())
		return
	}
	data.Title = types.StringValue(item.Title)
	data.CreateTime = types.StringValue(item.CreateTime)
	data.ModifyTime = types.StringValue(item.ModifyTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ItemNoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ItemNoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fields := map[string]string{}
	if plan.Title.ValueString() != state.Title.ValueString() {
		fields["title"] = plan.Title.ValueString()
	}
	if len(fields) > 0 {
		err := r.client.UpdateItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString(), fields)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update note item", err.Error())
			return
		}
	}
	item, err := r.client.ReadItemLogin(ctx, state.ItemID.ValueString(), state.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read note after update", err.Error())
		return
	}
	plan.ItemID = state.ItemID
	plan.ShareID = state.ShareID
	plan.Title = types.StringValue(item.Title)
	plan.CreateTime = types.StringValue(item.CreateTime)
	plan.ModifyTime = types.StringValue(item.ModifyTime)
	if plan.NoteWOVer.IsNull() || plan.NoteWOVer.IsUnknown() {
		plan.NoteWOVer = state.NoteWOVer
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemNoteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ItemNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteItemLogin(ctx, data.ItemID.ValueString(), data.ShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete note item", err.Error())
		return
	}
}

func (r *ItemNoteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: share_id:item_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: share_id:item_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("item_id"), idParts[1])...)
}
