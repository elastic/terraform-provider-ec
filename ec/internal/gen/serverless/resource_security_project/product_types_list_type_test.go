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

package resource_security_project

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestProductTypesListValue_ListSemanticEquals(t *testing.T) {
	ctx := context.Background()

	// Helper function to create a ProductTypesValue
	createProductTypesValue := func(line, tier string) ProductTypesValue {
		return ProductTypesValue{
			ProductLine: basetypes.NewStringValue(line),
			ProductTier: basetypes.NewStringValue(tier),
			state:       attr.ValueStateKnown,
		}
	}

	// Helper function to create a ProductTypesListValue from elements
	createListValue := func(elements []ProductTypesValue) ProductTypesListValue {
		attrValues := make([]attr.Value, len(elements))
		for i, elem := range elements {
			attrValues[i] = elem
		}
		return NewProductTypesListValueMust(ProductTypesValue{}.Type(ctx), attrValues)
	}

	tests := []struct {
		name     string
		value    ProductTypesListValue
		other    ProductTypesListValue
		expected bool
	}{
		{
			name: "equal lists with same order",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
			}),
			expected: true,
		},
		{
			name: "equal lists with different order",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("endpoint", "complete"),
				createProductTypesValue("security", "essentials"),
			}),
			expected: true,
		},
		{
			name: "different lists - different product lines",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("endpoint", "essentials"),
			}),
			expected: false,
		},
		{
			name: "different lists - different product tiers",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "complete"),
			}),
			expected: false,
		},
		{
			name: "different lengths",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
			}),
			expected: false,
		},
		{
			name:     "both empty lists",
			value:    createListValue([]ProductTypesValue{}),
			other:    createListValue([]ProductTypesValue{}),
			expected: true,
		},
		{
			name: "three elements in different order",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
				createProductTypesValue("cloud", "standard"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("cloud", "standard"),
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
			}),
			expected: true,
		},
		{
			name: "duplicate entries - same duplicates",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("security", "essentials"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("security", "essentials"),
			}),
			expected: true,
		},
		{
			name: "duplicate entries - different duplicates",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("security", "essentials"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("endpoint", "complete"),
			}),
			expected: false,
		},
		{
			name: "same product line with different tiers",
			value: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
				createProductTypesValue("security", "complete"),
			}),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "complete"),
				createProductTypesValue("security", "essentials"),
			}),
			expected: true,
		},
		{
			name: "null product line in first list falls back to standard equality",
			value: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringNull(),
					ProductTier: basetypes.NewStringValue("essentials"),
					state:       attr.ValueStateKnown,
				},
			}),
			other: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringNull(),
					ProductTier: basetypes.NewStringValue("essentials"),
					state:       attr.ValueStateKnown,
				},
			}),
			expected: true,
		},
		{
			name: "null product tier in first list falls back to standard equality",
			value: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringValue("security"),
					ProductTier: basetypes.NewStringNull(),
					state:       attr.ValueStateKnown,
				},
			}),
			other: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringValue("security"),
					ProductTier: basetypes.NewStringNull(),
					state:       attr.ValueStateKnown,
				},
			}),
			expected: true,
		},
		{
			name: "unknown product line in first list falls back to standard equality",
			value: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringUnknown(),
					ProductTier: basetypes.NewStringValue("essentials"),
					state:       attr.ValueStateKnown,
				},
			}),
			other: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringUnknown(),
					ProductTier: basetypes.NewStringValue("essentials"),
					state:       attr.ValueStateKnown,
				},
			}),
			expected: true,
		},
		{
			name: "unknown product tier in first list falls back to standard equality",
			value: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringValue("security"),
					ProductTier: basetypes.NewStringUnknown(),
					state:       attr.ValueStateKnown,
				},
			}),
			other: createListValue([]ProductTypesValue{
				{
					ProductLine: basetypes.NewStringValue("security"),
					ProductTier: basetypes.NewStringUnknown(),
					state:       attr.ValueStateKnown,
				},
			}),
			expected: true,
		},
		{
			name:  "null list compared with non-null list",
			value: NewProductTypesListValueNull(),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
			}),
			expected: false,
		},
		{
			name:     "both null lists",
			value:    NewProductTypesListValueNull(),
			other:    NewProductTypesListValueNull(),
			expected: true,
		},
		{
			name:  "unknown list compared with non-unknown list",
			value: NewProductTypesListValueUnknown(),
			other: createListValue([]ProductTypesValue{
				createProductTypesValue("security", "essentials"),
			}),
			expected: false,
		},
		{
			name:     "both unknown lists",
			value:    NewProductTypesListValueUnknown(),
			other:    NewProductTypesListValueUnknown(),
			expected: true,
		},
		{
			name:     "null list compared with unknown list",
			value:    NewProductTypesListValueNull(),
			other:    NewProductTypesListValueUnknown(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.value.ListSemanticEquals(ctx, tt.other)

			if diags.HasError() {
				t.Errorf("unexpected diagnostics: %v", diags)
			}

			if result != tt.expected {
				t.Errorf("ListSemanticEquals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestProductTypesListValue_ListSemanticEquals_WrongType(t *testing.T) {
	ctx := context.Background()

	createListValue := func(elements []ProductTypesValue) ProductTypesListValue {
		attrValues := make([]attr.Value, len(elements))
		for i, elem := range elements {
			attrValues[i] = elem
		}
		return NewProductTypesListValueMust(ProductTypesValue{}.Type(ctx), attrValues)
	}

	value := createListValue([]ProductTypesValue{
		{
			ProductLine: basetypes.NewStringValue("security"),
			ProductTier: basetypes.NewStringValue("essentials"),
			state:       attr.ValueStateKnown,
		},
	})

	// Compare with a standard ListValue instead of ProductTypesListValue
	standardList := basetypes.NewListValueMust(
		basetypes.StringType{},
		[]attr.Value{basetypes.NewStringValue("test")},
	)

	result, diags := value.ListSemanticEquals(ctx, standardList)

	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	if result != false {
		t.Errorf("ListSemanticEquals() with wrong type = %v, expected false", result)
	}
}

func TestProductTypesListValue_ListSemanticEquals_NullElementsInSecondList(t *testing.T) {
	ctx := context.Background()

	createListValue := func(elements []ProductTypesValue) ProductTypesListValue {
		attrValues := make([]attr.Value, len(elements))
		for i, elem := range elements {
			attrValues[i] = elem
		}
		return NewProductTypesListValueMust(ProductTypesValue{}.Type(ctx), attrValues)
	}

	tests := []struct {
		name     string
		value    ProductTypesValue
		other    ProductTypesValue
		expected bool
	}{
		{
			name: "null product line in second list - same in both",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringNull(),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringNull(),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			expected: true,
		},
		{
			name: "null product line in second list - different from first",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringNull(),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			expected: false,
		},
		{
			name: "null product tier in second list - same in both",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringNull(),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringNull(),
				state:       attr.ValueStateKnown,
			},
			expected: true,
		},
		{
			name: "null product tier in second list - different from first",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringNull(),
				state:       attr.ValueStateKnown,
			},
			expected: false,
		},
		{
			name: "unknown product line in second list - same in both",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringUnknown(),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringUnknown(),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			expected: true,
		},
		{
			name: "unknown product line in second list - different from first",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringUnknown(),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			expected: false,
		},
		{
			name: "unknown product tier in second list - same in both",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringUnknown(),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringUnknown(),
				state:       attr.ValueStateKnown,
			},
			expected: true,
		},
		{
			name: "unknown product tier in second list - different from first",
			value: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringValue("essentials"),
				state:       attr.ValueStateKnown,
			},
			other: ProductTypesValue{
				ProductLine: basetypes.NewStringValue("security"),
				ProductTier: basetypes.NewStringUnknown(),
				state:       attr.ValueStateKnown,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := createListValue([]ProductTypesValue{tt.value})
			other := createListValue([]ProductTypesValue{tt.other})

			result, diags := value.ListSemanticEquals(ctx, other)

			if diags.HasError() {
				t.Errorf("unexpected diagnostics: %v", diags)
			}

			if result != tt.expected {
				t.Errorf("ListSemanticEquals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
