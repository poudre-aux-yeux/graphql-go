package subscriptiontest

import (
	"encoding/json"
	stdErrors "errors"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"
)

type rootResolver struct {
	*helloSaidResolver
	*helloResolver
}

type rootResolverWithErrors struct {
	*helloSaidWithErrorsResolver
	*helloResolver
}

type helloSaidWithErrorsResolver struct{}

var resolverError = stdErrors.New("resolver error")

func (r *helloSaidWithErrorsResolver) HelloSaid() (chan *helloSaidEventResolver, chan<- struct{}, error) {
	return nil, nil, resolverError
}

type helloSaidResolver struct{}

type helloSaidEventResolver struct {
	msg string
}

func (r *helloSaidResolver) HelloSaid() (chan *helloSaidEventResolver, chan<- struct{}) {
	c := make(chan *helloSaidEventResolver)
	go func() {
		c <- &helloSaidEventResolver{msg: "Hello world!"}
		c <- &helloSaidEventResolver{msg: "Hello again!"}
		close(c)
	}()

	return c, make(chan<- struct{})
}

func (r *helloSaidEventResolver) Msg() string {
	return r.msg
}

type helloResolver struct{}

func (r *helloResolver) Hello() string {
	return "Hello world!"
}

func TestSchemaSubscribe(t *testing.T) {
	gqltesting.RunSubscribes(t, []*gqltesting.TestSubscription{
		{
			Name:   "subscribe_works",
			Schema: graphql.MustParseSchema(schema, &rootResolver{}),
			Query: `
				subscription onHelloSaid {
					helloSaid {
            msg
          }
				}
			`,
			ExpectedResults: []gqltesting.TestResponse{
				{
					Data: json.RawMessage(`
					{
						"helloSaid": {
							"msg": "Hello world!"
						}
					}
				`),
				},
				{
					Data: json.RawMessage(`
					{
						"helloSaid": {
							"msg": "Hello again!"
						}
					}
				`),
				},
			},
		},
	})
}

func TestSchemaSubscribe_Errors(t *testing.T) {
	gqltesting.RunSubscribes(t, []*gqltesting.TestSubscription{
		{
			Name:   "subscribe_to_query",
			Schema: graphql.MustParseSchema(schema, &rootResolver{}),
			Query: `
				query Hello {
					hello
				}
			`,
			ExpectedResults: []gqltesting.TestResponse{
				{
					Errors: []*errors.QueryError{errors.Errorf("%s: %s", "subscription unavailable for operation of type", "QUERY")},
				},
			},
		},
		{
			Name:   "resolver_can_error",
			Schema: graphql.MustParseSchema(schema, &rootResolverWithErrors{}),
			Query: `
				subscription onHelloSaid {
					helloSaid {
		        msg
		      }
				}
			`,
			ExpectedResults: []gqltesting.TestResponse{
				{
					Errors: []*errors.QueryError{errors.Errorf("%s", resolverError)},
				},
			},
		},
	})
}

const schema = `
  schema {
    subscription: Subscription,
		query: Query
  }

  type Subscription {
    helloSaid: HelloSaidEvent!
  }

  type HelloSaidEvent {
    msg: String!
  }

	type Query {
		hello: String!
	}
`
