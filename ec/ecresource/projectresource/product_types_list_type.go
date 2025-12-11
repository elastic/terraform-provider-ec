// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package projectresource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/resource_security_project"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ProductTypesListType is a custom list type that implements semantic equality
// for product_types, ignoring order differences when the content is the same.
// This allows Terraform to recognize that two lists with the same elements in
// different orders are semantically equivalent.
type ProductTypesListType struct {
	basetypes.ListType
}

func (t ProductTypesListType) Equal(o attr.Type) bool {
	other, ok := o.(ProductTypesListType)
	if !ok {
		return false
	}
	return t.ListType.Equal(other.ListType)
}

func (t ProductTypesListType) String() string {
	return "ProductTypesListType"
}

func (t ProductTypesListType) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	value := ProductTypesListValue{
		ListValue: in,
	}
	return value, nil
}

func (t ProductTypesListType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	listValue, ok := attrValue.(basetypes.ListValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	listValuable, diags := t.ValueFromList(ctx, listValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ListValue to ListValuable: %v", diags)
	}

	return listValuable, nil
}

func (t ProductTypesListType) ValueType(ctx context.Context) attr.Value {
	return ProductTypesListValue{}
}

// ProductTypesListValue is a custom list value that implements semantic equality
type ProductTypesListValue struct {
	basetypes.ListValue
}

func NewProductTypesListValueNull() ProductTypesListValue {
	return ProductTypesListValue{
		ListValue: basetypes.NewListNull(resource_security_project.ProductTypesValue{}.Type(context.Background())),
	}
}

func NewProductTypesListValueUnknown() ProductTypesListValue {
	return ProductTypesListValue{
		ListValue: basetypes.NewListUnknown(resource_security_project.ProductTypesValue{}.Type(context.Background())),
	}
}

func NewProductTypesListValueMust(elementType attr.Type, elements []attr.Value) ProductTypesListValue {
	return ProductTypesListValue{
		ListValue: basetypes.NewListValueMust(elementType, elements),
	}
}

func (v ProductTypesListValue) Equal(o attr.Value) bool {
	other, ok := o.(ProductTypesListValue)
	if !ok {
		return false
	}
	return v.ListValue.Equal(other.ListValue)
}

func (v ProductTypesListValue) Type(ctx context.Context) attr.Type {
	return ProductTypesListType{
		ListType: v.ListValue.Type(ctx).(basetypes.ListType),
	}
}

// ListSemanticEquals implements semantic equality that ignores order differences.
// Two product_types lists are considered equal if they contain the same set of
// product_line/product_tier combinations, regardless of order.
func (v ProductTypesListValue) ListSemanticEquals(ctx context.Context, otherV basetypes.ListValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	other, ok := otherV.(ProductTypesListValue)
	if !ok {
		return false, diags
	}

	// If either is null or unknown, use standard equality
	if v.IsNull() || v.IsUnknown() || other.IsNull() || other.IsUnknown() {
		return v.Equal(other), diags
	}

	// Get both lists
	var items, otherItems []resource_security_project.ProductTypesValue
	diags.Append(v.ElementsAs(ctx, &items, false)...)
	diags.Append(other.ElementsAs(ctx, &otherItems, false)...)

	if diags.HasError() {
		return false, diags
	}

	// If different lengths, they're different
	if len(items) != len(otherItems) {
		return false, diags
	}

	// Create maps of product_line -> product_tier for comparison
	itemsMap := make(map[string]string)
	for _, item := range items {
		itemsMap[item.ProductLine.ValueString()] = item.ProductTier.ValueString()
	}

	otherMap := make(map[string]string)
	for _, item := range otherItems {
		otherMap[item.ProductLine.ValueString()] = item.ProductTier.ValueString()
	}

	// Check if maps are equal (ignoring order)
	if len(itemsMap) != len(otherMap) {
		return false, diags
	}

	for k, v := range itemsMap {
		if otherMap[k] != v {
			return false, diags
		}
	}

	return true, diags
}
