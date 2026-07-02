# Use sampling

The `smpl` package helps you make a consistent sampling decision (sampled vs. not sampled) and store it in a `context.Context`. Other gosoline packages can then use that decision to change behavior (e.g. reduce log volume).

info

This guide focuses only on the sampling packages. If you want to see how sampling interacts with buffering logs on errors, read: [Sampling & fingers-crossed](/docs/how-to/logging/sampling-and-fingers-crossed.md).

## Concepts[​](#concepts "Direct link to Concepts")

* A **strategy** is a small function that can make a sampling decision based on the current context.
* A **decider** applies strategies in order and stores the final decision in the context.
* The decision lives on the context and can be read from anywhere down the call chain.

### Default behavior[​](#default-behavior "Direct link to Default behavior")

If there is no sampling decision on the context, gosoline treats it as **sampled**.

That means `smplctx.IsSampled(ctx)` returns `true` by default.

## Configuration[​](#configuration "Direct link to Configuration")

The sampling decider reads its settings from the `sampling` config key.

```
sampling:

  enabled: true

  strategies:

    - tracing
```

* `sampling.enabled`: if set to `false`, the decider behaves as if everything is sampled.
* `sampling.strategies`: list of strategy names in priority order. The first strategy that applies wins.

info

When sampling is enabled and a new decision is made, gosoline emits a metric `sampling_decision` (count) with dimension `sampled=true|false`.

### Built-in strategies[​](#built-in-strategies "Direct link to Built-in strategies")

Gosoline ships with a few strategies you can reference by name:

* `tracing`: if a trace is present in the context, use its sampling flag.
* `always`: always sample.
* `never`: never sample.
* `probabilistic`: guarantees at least one sampled decision per time window and additionally samples a small percentage of extra traffic.

#### Probabilistic settings[​](#probabilistic-settings "Direct link to Probabilistic settings")

The `probabilistic` strategy reads its settings from `sampling.settings.probabilistic`:

```
sampling:

  enabled: true

  strategies:

    - probabilistic

  settings:

    probabilistic:

      interval: 1s

      fixed_sample_count: 1

      extra_rate_percentage: 5
```

* `interval`: time window for the per-interval guarantee.
* `fixed_sample_count`: number of guaranteed `sampled=true` decisions per interval.
* `extra_rate_percentage`: probability (0–100) for sampling additional calls within the same interval.

## Make a decision[​](#make-a-decision "Direct link to Make a decision")

Typical flow:

1. Create (or provide) a decider.
2. Call `Decide`.
3. Use the returned context for the rest of your work.

The returned context is important because this is where the decision is stored.

main.go

```
package main



import (

	"context"

	"fmt"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/smpl"

	"github.com/justtrackio/gosoline/pkg/smpl/smplctx"

)



type ctxKey string



func init() {

	// Register a custom sampling strategy that can be referenced in config.

	//

	// This is process-global and should be done before application startup, before a decider is created.

	smpl.AddStrategy("force-by-context", func(ctx context.Context, config cfg.Config) (smpl.Strategy, error) {

		return func(ctx context.Context) (applied bool, sampled bool, err error) {

			if v, ok := ctx.Value(ctxKey("force_sample")).(bool); ok {

				return true, v, nil

			}



			return false, false, nil

		}, nil

	})

}



func main() {

	ctx := context.Background()

	config := cfg.New()

	logger := log.NewLogger()



	// Pretend this is coming from an incoming request/message.

	ctx = context.WithValue(ctx, ctxKey("force_sample"), false)



	// Build a decider from config. It reads `sampling.enabled` and `sampling.strategies`.

	// In real applications you typically use smpl.ProvideDecider(ctx, config).

	decider, err := smpl.NewDecider(ctx, config)

	if err != nil {

		logger.Error(ctx, "can not create decider: %w", err)



		return

	}



	// Decide applies the configured strategies.

	ctx, sampled, err := decider.Decide(ctx)

	if err != nil {

		logger.Error(ctx, "can not decide: %w", err)



		return

	}



	fmt.Printf("config decision: sampled=%v (smplctx.IsSampled=%v)\n", sampled, smplctx.IsSampled(ctx))



	// Per-call overrides: additional strategies run before the configured strategies.

	// This allows you to force sampling behaviour for specific code paths.

	alwaysStrategy, err := smpl.DecideByAlways(ctx, config)

	if err != nil {

		logger.Error(ctx, "can not build override strategy: %w", err)



		return

	}



	ctx, sampled, err = decider.Decide(ctx, alwaysStrategy)

	if err != nil {

		logger.Error(ctx, "can not decide with override: %w", err)



		return

	}



	fmt.Printf("override decision: sampled=%v (smplctx.IsSampled=%v)\n", sampled, smplctx.IsSampled(ctx))

}
```

## Override per call with additional strategies[​](#override-per-call-with-additional-strategies "Direct link to Override per call with additional strategies")

`Decide(ctx, additionalStrategies...)` lets you influence the decision for a specific code path.

Additional strategies are applied **before** the configured ones. This is useful when you need more control in a single place, without changing global config.

For example:

* Force sampling for a specific operation:

  ```
  always, _ := smpl.DecideByAlways(ctx, config)

  ctx, sampled, err := decider.Decide(ctx, always)
  ```

* Force not-sampled behavior:

  ```
  never, _ := smpl.DecideByNever(ctx, config)

  ctx, sampled, err := decider.Decide(ctx, never)
  ```

If your additional strategy can not decide, return `applied=false` so the configured strategies can decide.

A common example is HTTP request sampling: you can derive a strategy from the incoming request and pass it as an additional strategy (so it takes precedence over config). Gosoline's HTTP server middleware does this using the `X-Goso-Sampled` header.

## Register a custom strategy for config[​](#register-a-custom-strategy-for-config "Direct link to Register a custom strategy for config")

If you want to make a strategy configurable (usable via `sampling.strategies`), register it with `smpl.AddStrategy(name, strategy)`.

* This is process-global.
* Register strategies during application startup (before building the decider), so config parsing can resolve the strategy name.

Example pattern:

```
smpl.AddStrategy("my-strategy", func(ctx context.Context, config cfg.Config) (smpl.Strategy, error) {

	return func(ctx context.Context) (applied bool, sampled bool, err error) {

		// Decide on certain contexts, otherwise abstain.

		return false, false, nil

	}, nil

})
```

## Troubleshooting[​](#troubleshooting "Direct link to Troubleshooting")

* **"It always says sampled"**: if `sampling.enabled` is `false` or no decision is stored, `smplctx.IsSampled(ctx)` returns `true` by default.
* **"My strategy does nothing"**: some strategies only apply if required data exists on the context (e.g. `tracing` needs a trace in the context).
* **"I called Decide but later code doesn't see it"**: ensure you propagate the returned context (`ctx = newCtx`).
