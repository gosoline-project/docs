# Test your consumer

In this tutorial, you'll create an integration test for a message queue consumer that reads from an input and writes to an output.

## Before you begin[​](#before-you-begin "Direct link to Before you begin")

This tutorial tests a minimal application with the following files:

config.dist.yml

```
app:

  env: dev

  name: consumer

  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"

  tags:

    project: gosoline

    family: how-to

    group: grp



stream:

  input:

    consumer:

      type: sqs

      target_queue_id: events



  output:

    todos:

      type:

        type: sqs

        queue_id: todos
```

consumer.go

consumer.go

```
package consumertest



import (

	"context"

	"fmt"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/stream"

)



type Todo struct {

	Id     int

	Text   string

	Status string

}



type Consumer struct {

	producer stream.Producer

}



func NewConsumer(ctx context.Context, config cfg.Config, logger log.Logger) (stream.ConsumerCallback[Todo], error) {

	var err error

	var producer stream.Producer



	if producer, err = stream.NewProducer(ctx, config, logger, "todos"); err != nil {

		return nil, fmt.Errorf("can not create the todos producer: %w", err)

	}



	consumer := &Consumer{

		producer: producer,

	}



	return consumer, nil

}



func (c Consumer) Consume(ctx context.Context, todo Todo, attributes map[string]string) (bool, error) {

	todo.Status = "pending"



	if err := c.producer.WriteOne(ctx, todo); err != nil {

		return false, fmt.Errorf("can not write todo with id %d: %w", todo.Id, err)

	}



	return true, nil

}
```

## Write your integration test[​](#write-your-integration-test "Direct link to Write your integration test")

Here is a preview of all the code you'll cover in this tutorial:

consumer\_test.go

```
package consumertest



import (

	"testing"



	"github.com/justtrackio/gosoline/pkg/stream"

	"github.com/justtrackio/gosoline/pkg/test/suite"

)



type ConsumerTestSuite struct {

	suite.Suite

}



func TestConsumerTestSuite(t *testing.T) {

	suite.Run(t, new(ConsumerTestSuite))

}



func (s *ConsumerTestSuite) SetupSuite() []suite.Option {

	return []suite.Option{

		suite.WithConfigFile("config.dist.yml"),

		suite.WithConsumer(NewConsumer),

	}

}



func (s *ConsumerTestSuite) TestSuccess() *suite.StreamTestCase {

	return &suite.StreamTestCase{

		Input: map[string][]suite.StreamTestCaseInput{

			"consumer": {

				{

					Body: &Todo{

						Id:   1,

						Text: "do it",

					},

				},

			},

		},

		Output: map[string][]suite.StreamTestCaseOutput{

			"todos": {

				{

					Model: &Todo{},

					ExpectedBody: &Todo{

						Id:     1,

						Text:   "do it",

						Status: "pending",

					},

					ExpectedAttributes: map[string]string{

						stream.AttributeEncoding: stream.EncodingJson.String(),

					},

				},

			},

		},

		Assert: func() error {

			msgCount := s.Env().StreamOutput("todos").Len()

			s.Equal(1, msgCount)



			return nil

		},

	}

}
```

### Import your dependencies[​](#import-your-dependencies "Direct link to Import your dependencies")

First, import your dependencies:

```
import (

	"testing"



	"github.com/justtrackio/gosoline/pkg/test/suite"

)
```

Here, you've imported the standard `testing` package as well as gosoline's `test/suite`.

### Create a test suite[​](#create-a-test-suite "Direct link to Create a test suite")

Next, create a test suite that embeds the type `suite.Suite`:

```
type ConsumerTestSuite struct {

	suite.Suite

}
```

### Create a runner[​](#create-a-runner "Direct link to Create a runner")

Then, create the main test runner which will run the suite:

```
func TestConsumerTestSuite(t *testing.T) {

	suite.Run(t, new(ConsumerTestSuite))

}
```

### Set up your test suite[​](#set-up-your-test-suite "Direct link to Set up your test suite")

Implement `SetupSuite`:

```
func (s *ConsumerTestSuite) SetupSuite() []suite.Option {

	return []suite.Option{

		// 1

		suite.WithConfigFile("config.dist.yml"),



		// 2

		suite.WithConsumer(NewConsumer),

	}

}
```

This is run at the beginning of every suite. It's required, and it provides all the options to set up the suite.

Here, you:

1. Load settings from a specified config file.
2. Test the specified consumer.

### Create a test case[​](#create-a-test-case "Direct link to Create a test case")

Create a test case. This must match the pattern `TestCASE() *suite.StreamTestCase`, where `TestCASE` can be any name that starts with `Test`:

```
func (s *ConsumerTestSuite) TestSuccess() *suite.StreamTestCase {

	// 1

	return &suite.StreamTestCase{

		// 2

		Input: map[string][]suite.StreamTestCaseInput{

            // 3

			"consumer": {

				{					

					Body: &Todo{

						Id:   1,

						Text: "do it",

					},

				},

			},

		},

		// 4

		Output: map[string][]suite.StreamTestCaseOutput{

			"todos": {

				{					

					Model: &Todo{},				

					ExpectedBody: &Todo{

						Id:     1,

						Text:   "do it",

						Status: "pending",

					},					

					ExpectedAttributes: map[string]string{},

				},

			},

		},

        // 5

		Assert: func() error {

			msgCount := s.Env().StreamOutput("todos").Len()

			s.Equal(1, msgCount)



			return nil

		},

	}

}
```

In your integration test case, you:

1. Return the specification of the test case

2. Define the input for the consumer to use

3. Define your `Input` test data.

   <!-- -->

   * The top-level key is the input name. This must match the name from the configuration.
   * Under the input name, you define the test input stream. In this example, we're using a `Todo` struct that's defined in our example code.

4. Define the expected output just like you did with `Input`

   * The top-level key is the output name. This must match the name from the configuration.
   * Model is used to unmarshal the written bytes back into a struct.
   * `ExpectedBody` is compared with the written one. It has to be a perfect match.
   * If the messages also have attributes, they will be checked against `ExpectedAttributes`.

5. You can also provide an `Assert` function which is executed at the end of the test case in which everything else could be checked. Here, you can access the test environment. In this example, we get the output stream and check if there was exactly one message written.

:::info Technical Detail

In the sample configuration file we provided, the IO type was `sqs`. Even so, during the test, the type will automatically be set to `inMemory`. This is to bypass the need for external dependencies.

:::

## Run your test[​](#run-your-test "Direct link to Run your test")

Now that you've implemented your integration test, it's time to run it:

```
go mod tidy

go test
```

If your test succeeded, you will see the following response at the end of your log stream:

```
PASS
```

If your test failed, you will see a response like this:

```
--- FAIL: TestConsumerTestSuite (0.00s)

    --- FAIL: TestConsumerTestSuite/TestSuccess (0.00s)

        testcase_stream.go:78: 

                Error Trace:    testcase_stream.go:78

                                                        testcase_application.go:95

                                                        testcase_stream.go:57

                                                        run.go:149

                Error:          Not equal: 

                                expected: &main.Todo{Id:1, Text:"do it", Status:"pending"}

                                actual  : &main.Todo{Id:1, Text:"do it!", Status:"pending"}

                            

                                Diff:

                                --- Expected

                                +++ Actual

                                @@ -2,3 +2,3 @@

                                  Id: (int) 1,

                                - Text: (string) (len=5) "do it",

                                + Text: (string) (len=6) "do it!",

                                  Status: (string) (len=7) "pending"

                Test:           TestConsumerTestSuite/TestSuccess

                Messages:       body does not match

FAIL
```

This provides a detailed explanation of the test failure.

## Conclusion[​](#conclusion "Direct link to Conclusion")

You're done! You created your first integration test with gosoline.

Check out these resources to learn more about testing and creating consumers with gosoline:

* [Create a consumer](/docs/pr-1/getting-started/create-a-consumer.md)
