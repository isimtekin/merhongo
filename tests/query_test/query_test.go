package query_test

import (
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/query"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestQueryBuilder_GetFilter(t *testing.T) {
	// Simple filter
	builder := query.New().Where("name", "john")
	filter, err := builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value, exists := filter["name"]; !exists || value != "john" {
		t.Errorf("expected filter[name] = john, got %v", filter)
	}

	// Multiple conditions
	builder = query.New().Where("age", 30).Where("active", true)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value, exists := filter["age"]; !exists || value != 30 {
		t.Errorf("expected filter[age] = 30, got %v", filter)
	}
	if value, exists := filter["active"]; !exists || value != true {
		t.Errorf("expected filter[active] = true, got %v", filter)
	}
}

func TestQueryBuilder_Operators(t *testing.T) {
	// Test $eq operator
	builder := query.New().Equals("age", 30)
	filter, err := builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ageFilter, exists := filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok := ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$eq"] != 30 {
		t.Errorf("expected filter[age][$eq] = 30, got %v", ageMap["$eq"])
	}

	// Test $ne operator
	builder = query.New().NotEquals("status", "inactive")
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	statusFilter, exists := filter["status"]
	if !exists {
		t.Fatalf("expected filter to have status key")
	}
	statusMap, ok := statusFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[status] to be bson.M, got %T", statusFilter)
	}
	if statusMap["$ne"] != "inactive" {
		t.Errorf("expected filter[status][$ne] = inactive, got %v", statusMap["$ne"])
	}

	// Test $gt operator
	builder = query.New().GreaterThan("age", 25)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ageFilter, exists = filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok = ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$gt"] != 25 {
		t.Errorf("expected filter[age][$gt] = 25, got %v", ageMap["$gt"])
	}

	// Test $lt operator
	builder = query.New().LessThan("age", 50)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ageFilter, exists = filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok = ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$lt"] != 50 {
		t.Errorf("expected filter[age][$lt] = 50, got %v", ageMap["$lt"])
	}

	// Test $in operator
	values := []string{"active", "pending"}
	builder = query.New().In("status", values)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	statusFilter, exists = filter["status"]
	if !exists {
		t.Fatalf("expected filter to have status key")
	}
	statusMap, ok = statusFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[status] to be bson.M, got %T", statusFilter)
	}
	inValues, ok := statusMap["$in"].([]string)
	if !ok {
		t.Fatalf("expected filter[status][$in] to be []string, got %T", statusMap["$in"])
	}
	if len(inValues) != 2 || inValues[0] != "active" || inValues[1] != "pending" {
		t.Errorf("expected filter[status][$in] = [active, pending], got %v", inValues)
	}
}

func TestQueryBuilder_SortByLimitSkip(t *testing.T) {
	builder := query.New().
		Where("active", true).
		SortBy("name", true).
		SortBy("age", false).
		Limit(10).
		Skip(20)

	// Test filter
	filter, err := builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value, exists := filter["active"]; !exists || value != true {
		t.Errorf("expected filter[active] = true, got %v", filter)
	}

	// Test options
	options, err := builder.GetOptions()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if options.Limit == nil || *options.Limit != 10 {
		t.Errorf("expected options.Limit = 10, got %v", options.Limit)
	}
	if options.Skip == nil || *options.Skip != 20 {
		t.Errorf("expected options.Skip = 20, got %v", options.Skip)
	}
}

func TestQueryBuilder_Build(t *testing.T) {
	builder := query.New().
		Where("active", true).
		GreaterThan("age", 25).
		SortBy("name", true).
		Limit(10)

	filter, options, err := builder.Build()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check filter
	if value, exists := filter["active"]; !exists || value != true {
		t.Errorf("expected filter[active] = true, got %v", filter)
	}

	ageFilter, exists := filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok := ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$gt"] != 25 {
		t.Errorf("expected filter[age][$gt] = 25, got %v", ageMap["$gt"])
	}

	// Check options
	if options.Limit == nil || *options.Limit != 10 {
		t.Errorf("expected options.Limit = 10, got %v", options.Limit)
	}
}

func TestQueryBuilder_ValidationErrors(t *testing.T) {
	// Test empty key
	builder := query.New().Where("", "value")
	_, err := builder.GetFilter()
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
	if !errors.IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}

	// Test negative limit
	builder = query.New().Limit(-10)
	_, err = builder.GetOptions()
	if err == nil {
		t.Error("expected error for negative limit, got nil")
	}
	if !errors.IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}

	// Test negative skip
	builder = query.New().Skip(-5)
	_, err = builder.GetOptions()
	if err == nil {
		t.Error("expected error for negative skip, got nil")
	}
	if !errors.IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}

	// Test error propagation
	builder = query.New().Where("", "invalid") // This creates an error
	// The following operations should not clear the error
	builder.Where("name", "john").Limit(10)

	_, _, err = builder.Build()
	if err == nil {
		t.Error("expected error to be propagated, got nil")
	}
}

// Test the new features
func TestQueryBuilder_NewFeatures(t *testing.T) {
	// Test WithError
	expectedErr := errors.WithDetails(errors.ErrValidation, "test error")
	builder := query.WithError(expectedErr)
	if builder.Error() != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, builder.Error())
	}

	// Test WhereOperator
	builder = query.New().WhereOperator("status", "$custom", "value")
	filter, err := builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	statusFilter, exists := filter["status"]
	if !exists {
		t.Fatalf("expected filter to have status key")
	}
	statusMap, ok := statusFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[status] to be bson.M, got %T", statusFilter)
	}
	if statusMap["$custom"] != "value" {
		t.Errorf("expected filter[status][$custom] = value, got %v", statusMap["$custom"])
	}

	// Test GreaterThanOrEqual
	builder = query.New().GreaterThanOrEqual("age", 18)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ageFilter, exists := filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok := ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$gte"] != 18 {
		t.Errorf("expected filter[age][$gte] = 18, got %v", ageMap["$gte"])
	}

	// Test LessThanOrEqual
	builder = query.New().LessThanOrEqual("age", 65)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ageFilter, exists = filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok = ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$lte"] != 65 {
		t.Errorf("expected filter[age][$lte] = 65, got %v", ageMap["$lte"])
	}

	// Test NotIn
	values := []string{"deleted", "banned"}
	builder = query.New().NotIn("status", values)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	statusFilter, exists = filter["status"]
	if !exists {
		t.Fatalf("expected filter to have status key")
	}
	statusMap, ok = statusFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[status] to be bson.M, got %T", statusFilter)
	}
	ninValues, ok := statusMap["$nin"].([]string)
	if !ok {
		t.Fatalf("expected filter[status][$nin] to be []string, got %T", statusMap["$nin"])
	}
	if len(ninValues) != 2 || ninValues[0] != "deleted" || ninValues[1] != "banned" {
		t.Errorf("expected filter[status][$nin] = [deleted, banned], got %v", ninValues)
	}

	// Test Exists
	builder = query.New().Exists("email", true)
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	emailFilter, exists := filter["email"]
	if !exists {
		t.Fatalf("expected filter to have email key")
	}
	emailMap, ok := emailFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[email] to be bson.M, got %T", emailFilter)
	}
	if emailMap["$exists"] != true {
		t.Errorf("expected filter[email][$exists] = true, got %v", emailMap["$exists"])
	}

	// Test Regex
	builder = query.New().Regex("name", "^john", "i")
	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	nameFilter, exists := filter["name"]
	if !exists {
		t.Fatalf("expected filter to have name key")
	}
	nameMap, ok := nameFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[name] to be bson.M, got %T", nameFilter)
	}
	if nameMap["$regex"] != "^john" {
		t.Errorf("expected filter[name][$regex] = ^john, got %v", nameMap["$regex"])
	}
	if nameMap["$options"] != "i" {
		t.Errorf("expected filter[name][$options] = i, got %v", nameMap["$options"])
	}

	// Test MergeFilter
	builder = query.New().Where("active", true)
	additionalFilter := bson.M{"age": bson.M{"$gt": 18}}
	builder.MergeFilter(additionalFilter)

	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if value, exists := filter["active"]; !exists || value != true {
		t.Errorf("expected filter[active] = true, got %v", filter)
	}

	ageFilter, exists = filter["age"]
	if !exists {
		t.Fatalf("expected filter to have age key")
	}
	ageMap, ok = ageFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[age] to be bson.M, got %T", ageFilter)
	}
	if ageMap["$gt"] != 18 {
		t.Errorf("expected filter[age][$gt] = 18, got %v", ageMap["$gt"])
	}

	// Test merging complex filters
	builder = query.New().Where("status", bson.M{"$in": []string{"active", "pending"}})
	additionalFilter = bson.M{"status": bson.M{"$ne": "deleted"}}
	builder.MergeFilter(additionalFilter)

	filter, err = builder.GetFilter()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	statusFilter, exists = filter["status"]
	if !exists {
		t.Fatalf("expected filter to have status key")
	}
	statusMap, ok = statusFilter.(bson.M)
	if !ok {
		t.Fatalf("expected filter[status] to be bson.M, got %T", statusFilter)
	}

	// Should have both $in and $ne
	if _, exists := statusMap["$in"]; !exists {
		t.Errorf("expected filter[status] to have $in key")
	}
	if _, exists := statusMap["$ne"]; !exists {
		t.Errorf("expected filter[status] to have $ne key")
	}
}
