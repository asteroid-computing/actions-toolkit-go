# Default environment variables

`actioncontext` reads the default environment variables that GitHub Actions sets for every step. `actioncontext.Load()` returns them as an `ActionContext`; this table maps each variable to its field. GitHub's [variables reference](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables) is authoritative for what each variable means and when it is set.

Unset variables leave strings empty and `*bool` fields nil; integer fields are 0 when unset or unparsable. Booleans parse with `strconv.ParseBool`, so `RUNNER_DEBUG=1` is true.

| Variable                     | Field                  | Type     | Notes                                                        |
| ---------------------------- | ---------------------- | -------- | ------------------------------------------------------------ |
| `CI`                         | `GithubCI`             | `*bool`  |                                                              |
| `GITHUB_ACTION`              | `Action`               | `string` | Action name or step id                                       |
| `GITHUB_ACTION_PATH`         | `ActionPath`           | `string` | Composite actions only                                       |
| `GITHUB_ACTION_REPOSITORY`   | `ActionRepository`     | `string` | `owner/repo` of the running action                           |
| `GITHUB_ACTIONS`             | `Actions`              | `*bool`  | True under GitHub Actions; distinguishes CI from local runs  |
| `GITHUB_ACTOR`               | `Actor`                | `string` |                                                              |
| `GITHUB_ACTOR_ID`            | `ActorID`              | `string` |                                                              |
| `GITHUB_API_URL`             | `APIURL`               | `string` | Used by `NewGitHubClient` to target github.com or GHES       |
| `GITHUB_BASE_REF`            | `BaseRef`              | `string` | `pull_request` and `pull_request_target` only                |
| `GITHUB_ENV`                 | `EnvFilePath`          | `string` | File behind `core.SetEnv`                                    |
| `GITHUB_EVENT_NAME`          | `EventName`            | `string` | Selects the `EventPayload` type                              |
| `GITHUB_EVENT_PATH`          | `EventPayloadFilePath` | `string` | Decoded into `EventPayload`; raw bytes via `RawEventPayload` |
| `GITHUB_GRAPHQL_URL`         | `GraphQLURL`           | `string` |                                                              |
| `GITHUB_HEAD_REF`            | `HeadRef`              | `string` | `pull_request` and `pull_request_target` only                |
| `GITHUB_JOB`                 | `JobID`                | `string` |                                                              |
| `GITHUB_OUTPUT`              | `OutputFilePath`       | `string` | File behind `core.SetOutput`                                 |
| `GITHUB_PATH`                | `PathFilePath`         | `string` | File behind `core.AddToPath`                                 |
| `GITHUB_REF`                 | `Ref`                  | `string` | Fully qualified, e.g. `refs/heads/main`                      |
| `GITHUB_REF_NAME`            | `RefName`              | `string` |                                                              |
| `GITHUB_REF_PROTECTED`       | `RefProtected`         | `*bool`  |                                                              |
| `GITHUB_REF_TYPE`            | `RefType`              | `string` | `branch` or `tag`                                            |
| `GITHUB_REPOSITORY`          | `Repository`           | `string` | `owner/repo`                                                 |
| `GITHUB_REPOSITORY_ID`       | `RepositoryID`         | `string` |                                                              |
| `GITHUB_REPOSITORY_OWNER`    | `RepositoryOwner`      | `string` |                                                              |
| `GITHUB_REPOSITORY_OWNER_ID` | `RepositoryOwnerID`    | `string` |                                                              |
| `GITHUB_RETENTION_DAYS`      | `RetentionDays`        | `int`    |                                                              |
| `GITHUB_RUN_ATTEMPT`         | `RunAttempt`           | `int`    |                                                              |
| `GITHUB_RUN_ID`              | `RunID`                | `int`    |                                                              |
| `GITHUB_RUN_NUMBER`          | `RunNumber`            | `int`    |                                                              |
| `GITHUB_SERVER_URL`          | `ServerURL`            | `string` |                                                              |
| `GITHUB_SHA`                 | `Sha`                  | `string` | Depends on the triggering event                              |
| `GITHUB_STEP_SUMMARY`        | `StepSummaryFilePath`  | `string` | File behind `core.StepSummary`                               |
| `GITHUB_TRIGGERING_ACTOR`    | `TriggeringActor`      | `string` | May differ from `Actor` on re-runs                           |
| `GITHUB_WORKFLOW`            | `Workflow`             | `string` |                                                              |
| `GITHUB_WORKFLOW_REF`        | `WorkflowRef`          | `string` |                                                              |
| `GITHUB_WORKFLOW_SHA`        | `WorkflowSha`          | `string` |                                                              |
| `GITHUB_WORKSPACE`           | `WorkspacePath`        | `string` |                                                              |
| `RUNNER_ARCH`                | `RunnerArch`           | `string` |                                                              |
| `RUNNER_DEBUG`               | `RunnerDebug`          | `*bool`  | Set only when debug logging is enabled                       |
| `RUNNER_ENVIRONMENT`         | `RunnerEnvironment`    | `string` | `github-hosted` or `self-hosted`                             |
| `RUNNER_NAME`                | `RunnerName`           | `string` |                                                              |
| `RUNNER_OS`                  | `RunnerOS`             | `string` |                                                              |
| `RUNNER_TEMP`                | `RunnerTempPath`       | `string` |                                                              |
| `RUNNER_TOOL_CACHE`          | `RunnerToolCachePath`  | `string` |                                                              |

`GITHUB_STATE`, the file behind `core.SetState`, is read by `core` directly and has no `ActionContext` field.
