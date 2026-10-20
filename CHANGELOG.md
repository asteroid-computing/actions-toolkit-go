# Changelog

## [0.2.0](https://github.com/asteroid-computing/actions-toolkit-go/compare/v0.1.1...v0.2.0) (2026-10-20)


### ⚠ BREAKING CHANGES

* **core:** core.FileCommand, core.NewSummary and core.LookupEnvFunc are removed. Use SetEnv, SetOutput, SetState or AddToPath instead of FileCommand, and core.StepSummary instead of NewSummary.
* **actioncontext:** actioncontext.CurrentActionContext and the package-level variables (Actor, Repository, EventName, EventPayload, ...) are removed. Call `ac, err := actioncontext.Load()` and read ac's fields instead; Load returns an error when the event file cannot be read or decoded, and the context is no longer logged automatically (call core.Debug(ac.String()) for the old output).

### Features

* **actioncontext:** load the context explicitly with Load ([#5](https://github.com/asteroid-computing/actions-toolkit-go/issues/5)) ([b8116fa](https://github.com/asteroid-computing/actions-toolkit-go/commit/b8116fa08765cd17e2f907b2a4f60fc662c78e76))
* **core:** add GetIDToken for GitHub OIDC tokens ([#9](https://github.com/asteroid-computing/actions-toolkit-go/issues/9)) ([1255a09](https://github.com/asteroid-computing/actions-toolkit-go/commit/1255a09a785d5913c9c1cc2ae4180d6d3fa92cc5))
* **core:** add GetState ([#8](https://github.com/asteroid-computing/actions-toolkit-go/issues/8)) ([6d00664](https://github.com/asteroid-computing/actions-toolkit-go/commit/6d0066454304d907510aee2f4260959c13ab35f9))
* **core:** add Info and Group ([6d00664](https://github.com/asteroid-computing/actions-toolkit-go/commit/6d0066454304d907510aee2f4260959c13ab35f9))
* **core:** add IsDebug and SetFailed ([6d00664](https://github.com/asteroid-computing/actions-toolkit-go/commit/6d0066454304d907510aee2f4260959c13ab35f9))
* **core:** add path conversion helpers ([#11](https://github.com/asteroid-computing/actions-toolkit-go/issues/11)) ([618e2e8](https://github.com/asteroid-computing/actions-toolkit-go/commit/618e2e8cf94912fd21def4227fdd0bd68d7b81ce))
* **core:** add SetCommandEcho and SetOutputJSON ([618e2e8](https://github.com/asteroid-computing/actions-toolkit-go/commit/618e2e8cf94912fd21def4227fdd0bd68d7b81ce))


### Bug Fixes

* **core:** render workflow command properties in sorted order ([b8116fa](https://github.com/asteroid-computing/actions-toolkit-go/commit/b8116fa08765cd17e2f907b2a4f60fc662c78e76))
* **core:** update the current process in SetEnv and AddToPath ([#7](https://github.com/asteroid-computing/actions-toolkit-go/issues/7)) ([e3c62a3](https://github.com/asteroid-computing/actions-toolkit-go/commit/e3c62a373c18de08f3a0e5276a0befb20289d9a9))
* **core:** write one directory per line to GITHUB_PATH in AddToPath ([e3c62a3](https://github.com/asteroid-computing/actions-toolkit-go/commit/e3c62a373c18de08f3a0e5276a0befb20289d9a9))


### Code Refactoring

* **core:** keep afero out of the exported API ([b8116fa](https://github.com/asteroid-computing/actions-toolkit-go/commit/b8116fa08765cd17e2f907b2a4f60fc662c78e76))

## [0.1.1](https://github.com/asteroid-computing/actions-toolkit-go/compare/v0.1.0...v0.1.1) (2026-10-12)


### Bug Fixes

* **core:** return an error instead of panicking on incomplete lsb_release output ([d9fc995](https://github.com/asteroid-computing/actions-toolkit-go/commit/d9fc9954bb3979aea93d71ad769c944bd079c9b8))

## 0.1.0 (2026-10-12)


### Features

* initial Go actions toolkit with walle tooling ([#1](https://github.com/asteroid-computing/actions-toolkit-go/issues/1)) ([61733ca](https://github.com/asteroid-computing/actions-toolkit-go/commit/61733ca06948bac2c1e9a00cd3541b6886608d74))
