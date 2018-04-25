package gqltesting

import (
	"bytes"
	"context"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
)

// TestResponse ...
type TestResponse struct {
	Data   string
	Errors []*errors.QueryError
}

// TestSubscription is a GraphQL test case to be used with RunSubscribe.
type TestSubscription struct {
	Context         context.Context
	Schema          *graphql.Schema
	Query           string
	OperationName   string
	Variables       map[string]interface{}
	ExpectedResults []TestResponse
}

// RunSubscribe runs a single GraphQL subscription test case.
func RunSubscribe(t *testing.T, test *TestSubscription) {
	if test.Context == nil {
		test.Context = context.Background()
	}
	c := test.Schema.Subscribe(test.Context, test.Query, test.OperationName, test.Variables)

	var results []*graphql.Response
	for res := range c {
		results = append(results, res)
	}

	for i, expected := range test.ExpectedResults {
		res := results[i]

		checkErrors(t, expected.Errors, res.Errors)

		got := formatJSON(t, res.Data)
		want := formatJSON(t, []byte(expected.Data))

		if !bytes.Equal(got, want) {
			t.Logf("got:  %s", got)
			t.Logf("want: %s", want)
			t.Fail()
		}
	}
}
