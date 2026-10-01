## [Clearer Stream Processing, Safer Shutdown](/docs/pr-4/blog/gosoline-v0-66-0-stream-processing/.md)

September 30, 2026 ·

<!-- -->

7 min read

[![Jan Kamieth](https://avatars.githubusercontent.com/u/783502?s=400\&v=4)](https://github.com/j4k4)

[Jan Kamieth](https://github.com/j4k4)

Stream consumers spend most of their time doing something straightforward: receive a message, run application logic, and record that the message was handled. The difficult part is making that flow behave predictably when callbacks run concurrently, a partition changes owners, or a deployment interrupts processing.

Gosoline **[v0.66.0](https://github.com/justtrackio/gosoline/releases/tag/v0.66.0)** brings those concerns together in a simpler stream lifecycle. Inputs call processing callbacks directly, concurrency and ordering become explicit transport decisions, and processing drain is separated from transport cleanup. The result is a clearer model for building and operating streaming services.

**Tags:**

* [gosoline](/docs/pr-4/blog/tags/gosoline/.md)
* [streaming](/docs/pr-4/blog/tags/streaming/.md)
* [kafka](/docs/pr-4/blog/tags/kafka/.md)
* [lifecycle](/docs/pr-4/blog/tags/lifecycle/.md)

[**Read more**](/docs/pr-4/blog/gosoline-v0-66-0-stream-processing/.md)
