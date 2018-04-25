package subscriptiontest

import (
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"
)

type rootResolver struct {
	*helloSaidResolver
	*helloResolver
}

type helloSaidResolver struct{}

type helloSaidEventResolver struct {
	msg string
}

func (r *helloSaidResolver) HelloSaid() chan *helloSaidEventResolver {
	c := make(chan *helloSaidEventResolver)
	go func() {
		c <- &helloSaidEventResolver{msg: "Hello world!"}
		c <- &helloSaidEventResolver{msg: "Hello again!"}
		close(c)
	}()

	return c
}

func (r *helloSaidEventResolver) Msg() string {
	return r.msg
}

type helloResolver struct{}

func (r *helloResolver) Hello() string {
	return "Hello world!"
}

func TestSchemaSubscribe(t *testing.T) {
	gqltesting.RunSubscribe(t, &gqltesting.TestSubscription{
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
				Data: `
					{
						"helloSaid": {
							"msg": "Hello world!"
						}
					}
				`,
			},
			{
				Data: `
					{
						"helloSaid": {
							"msg": "Hello again!"
						}
					}
				`,
			},
		}})
}

func TestSchemaSubscribe_Errors(t *testing.T) {
	gqltesting.RunSubscribe(t, &gqltesting.TestSubscription{
		Schema: graphql.MustParseSchema(schema, &rootResolver{}),
		Query: `
				query Hello {
					hello
				}
			`,
		ExpectedResults: []gqltesting.TestResponse{
			{
				Errors: []*errors.QueryError{errors.Errorf("%s: %s", "subscription unavailable for operation of type", "QUERY")},
				Data:   `{}`,
			},
		}})
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
