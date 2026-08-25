# [3.9.0](https://github.com/martynvdijke/sandwitches/compare/v3.8.5...v3.9.0) (2026-08-25)


### Bug Fixes

* **ci:** build frontend bundle before python UI tests ([41dda6f](https://github.com/martynvdijke/sandwitches/commit/41dda6fde6ca2e166423263924b56926a9440058))


### Features

* add photo zoom lightbox for recipe photos ([4e085a0](https://github.com/martynvdijke/sandwitches/commit/4e085a07e4b0b737cc9500754fdbf6eb9aae5b55))

## [3.8.5](https://github.com/martynvdijke/sandwitches/compare/v3.8.4...v3.8.5) (2026-08-24)

## [3.8.4](https://github.com/martynvdijke/sandwitches/compare/v3.8.3...v3.8.4) (2026-08-20)

## [3.8.3](https://github.com/martynvdijke/sandwitches/compare/v3.8.2...v3.8.3) (2026-08-18)


### Bug Fixes

* use latest ([0662b62](https://github.com/martynvdijke/sandwitches/commit/0662b62f8717e03a2603710634ac2392049c7a49))

## [3.8.2](https://github.com/martynvdijke/sandwitches/compare/v3.8.1...v3.8.2) (2026-08-18)

## [3.8.1](https://github.com/martynvdijke/sandwitches/compare/v3.8.0...v3.8.1) (2026-08-18)


### Bug Fixes

* **deps:** update module github.com/mattn/go-sqlite3 to v1.14.50 ([#149](https://github.com/martynvdijke/sandwitches/issues/149)) ([aa91db1](https://github.com/martynvdijke/sandwitches/commit/aa91db165903a15addc1f93248f92f0414124d61))

# [3.8.0](https://github.com/martynvdijke/sandwitches/compare/v3.7.2...v3.8.0) (2026-08-16)


### Features

* **auth:** add password reset flow with emailed reset links ([e2c0bd1](https://github.com/martynvdijke/sandwitches/commit/e2c0bd110ddec649cb82bcb858afc1328a753044))

## [3.7.2](https://github.com/martynvdijke/sandwitches/compare/v3.7.1...v3.7.2) (2026-08-16)

## [3.7.1](https://github.com/martynvdijke/sandwitches/compare/v3.7.0...v3.7.1) (2026-08-15)


### Bug Fixes

* **trmnl:** fall back to instance URL when plugin custom field is unset ([96b9a24](https://github.com/martynvdijke/sandwitches/commit/96b9a24746b82c36f51571d4fe1f66b7acb8f99c))

# [3.7.0](https://github.com/martynvdijke/sandwitches/compare/v3.6.1...v3.7.0) (2026-08-15)


### Bug Fixes

* **trmnl:** reference current icon asset in plugin layouts ([139ad92](https://github.com/martynvdijke/sandwitches/commit/139ad9278e9fcabc78570cf1d08a4b247d933779))
* **trmnl:** remove leftover instance URL from half_horizontal title bar ([3453d9b](https://github.com/martynvdijke/sandwitches/commit/3453d9bf48dd6941a1348d59a99e8379f1ed7699))


### Features

* **trmnl:** add uploader and fan list rendering to quadrant layout ([3e78f74](https://github.com/martynvdijke/sandwitches/commit/3e78f747cceac6204967346613ea2b32c4cf16da))

## [3.6.1](https://github.com/martynvdijke/sandwitches/compare/v3.6.0...v3.6.1) (2026-08-15)


### Bug Fixes

* **trmnl:** restructure plugin into trmnlp layout and point workflow at trmnl/ ([0d81988](https://github.com/martynvdijke/sandwitches/commit/0d819881bb5ea99171df4dc94730fe5d2a70c712))

# [3.6.0](https://github.com/martynvdijke/sandwitches/compare/v3.5.0...v3.6.0) (2026-08-15)


### Features

* **ui:** link API menu item to OpenAPI docs instead of ping ([8d444e8](https://github.com/martynvdijke/sandwitches/commit/8d444e81e042713c594b96f12e7d5785d586bd78))

# [3.5.0](https://github.com/martynvdijke/sandwitches/compare/v3.4.0...v3.5.0) (2026-08-15)


### Features

* **api:** add pagination, filtering, error envelope, rate limiting, CORS, ETag caching ([49d41bf](https://github.com/martynvdijke/sandwitches/commit/49d41bf2ea4389c9d38a254ce6cce96aceb00529))

# [3.4.0](https://github.com/martynvdijke/sandwitches/compare/v3.3.0...v3.4.0) (2026-08-15)


### Features

* **api:** restore v2.x API parity with DTOs, auth 401s, OpenAPI docs ([7e5060b](https://github.com/martynvdijke/sandwitches/commit/7e5060b0e1a672ae7b6f7ae5e37c87f4deed244c))
* **trmnl:** restore TRMNL plugin payloads and template URLs ([9396bc8](https://github.com/martynvdijke/sandwitches/commit/9396bc8c86e6f004df9ed288f370e14b021c33d0))

# [3.3.0](https://github.com/martynvdijke/sandwitches/compare/v3.2.0...v3.3.0) (2026-08-15)


### Features

* rotate legacy log file and write Go logs to sandwitches.log ([408355f](https://github.com/martynvdijke/sandwitches/commit/408355f1f4520fdc8b34b69ae0222228fe238397))

# [3.2.0](https://github.com/martynvdijke/sandwitches/compare/v3.1.3...v3.2.0) (2026-08-15)


### Features

* run Go binary directly and drop legacy Django tables ([efe460c](https://github.com/martynvdijke/sandwitches/commit/efe460ce7a2c0cd3e5b138d5729feaa0ac3f8a25))

## [3.1.3](https://github.com/martynvdijke/sandwitches/compare/v3.1.2...v3.1.3) (2026-08-15)


### Bug Fixes

* make navbar full width ([96d07ca](https://github.com/martynvdijke/sandwitches/commit/96d07ca64a5f13c7f79fcffe81bd06300e303a10))
* repair admin charts and user deletion ([2f411fa](https://github.com/martynvdijke/sandwitches/commit/2f411fa42d0061f45a1e363d666bb29ef1b8ae03))

## [3.1.2](https://github.com/martynvdijke/sandwitches/compare/v3.1.1...v3.1.2) (2026-08-14)


### Bug Fixes

* **deps:** update all non-major dependencies ([#145](https://github.com/martynvdijke/sandwitches/issues/145)) ([d80f0d0](https://github.com/martynvdijke/sandwitches/commit/d80f0d0c721f77c9b15f2ea4ee3787dcdae00f17))

## [3.1.1](https://github.com/martynvdijke/sandwitches/compare/v3.1.0...v3.1.1) (2026-08-14)


### Bug Fixes

* use proper beercss drawer layout in sidebar menu ([21703f6](https://github.com/martynvdijke/sandwitches/commit/21703f68c4703d8198159691eaa62df7ce1e0a64))

# [3.1.0](https://github.com/martynvdijke/sandwitches/compare/v3.0.12...v3.1.0) (2026-08-13)


### Features

* resize and compress images at upload time ([a10bf1a](https://github.com/martynvdijke/sandwitches/commit/a10bf1a8ee2701381cfdfa5f4596b6f0313c369b))

## [3.0.12](https://github.com/martynvdijke/sandwitches/compare/v3.0.11...v3.0.12) (2026-08-13)


### Bug Fixes

* serve resized thumbnails for recipe cards and fix RSS base URL ([06e6e5c](https://github.com/martynvdijke/sandwitches/commit/06e6e5ca273dbde320b49acb674689dca0ca1bc7))

## [3.0.11](https://github.com/martynvdijke/sandwitches/compare/v3.0.10...v3.0.11) (2026-08-13)


### Bug Fixes

* restore homepage viewport and show release version in footer ([a2958f6](https://github.com/martynvdijke/sandwitches/commit/a2958f6df7c50dd98d362f07b952a3dc2ac2f353))

## [3.0.10](https://github.com/martynvdijke/sandwitches/compare/v3.0.9...v3.0.10) (2026-08-13)

## [3.0.9](https://github.com/martynvdijke/sandwitches/compare/v3.0.8...v3.0.9) (2026-08-12)


### Bug Fixes

* preserve Django logins and repair homepage filters ([4852032](https://github.com/martynvdijke/sandwitches/commit/4852032fc6711e723d179158f76dcff5d7c60bb2))

## [3.0.8](https://github.com/martynvdijke/sandwitches/compare/v3.0.7...v3.0.8) (2026-08-12)

## [3.0.7](https://github.com/martynvdijke/sandwitches/compare/v3.0.6...v3.0.7) (2026-08-12)


### Bug Fixes

* **deps:** update module golang.org/x/crypto to v0.55.0 ([#143](https://github.com/martynvdijke/sandwitches/issues/143)) ([2bb3ee0](https://github.com/martynvdijke/sandwitches/commit/2bb3ee0d2929108ecf3552f64d372256d832bc69))

## [3.0.6](https://github.com/martynvdijke/sandwitches/compare/v3.0.5...v3.0.6) (2026-08-11)


### Bug Fixes

* repair anonymous homepage truncation and flatten Go app to repo root ([c4e996a](https://github.com/martynvdijke/sandwitches/commit/c4e996a949a84fad3ecfa049332829bb2db2b060))

## [3.0.5](https://github.com/martynvdijke/sandwitches/compare/v3.0.4...v3.0.5) (2026-08-11)

## [3.0.4](https://github.com/martynvdijke/sandwitches/compare/v3.0.3...v3.0.4) (2026-08-11)

## [3.0.3](https://github.com/martynvdijke/sandwitches/compare/v3.0.2...v3.0.3) (2026-08-11)


### Bug Fixes

* **deps:** update all non-major dependencies ([#131](https://github.com/martynvdijke/sandwitches/issues/131)) ([4ba3eab](https://github.com/martynvdijke/sandwitches/commit/4ba3eab8fe564420ff25e37f15ef8401a71ce8e2))

## [3.0.2](https://github.com/martynvdijke/sandwitches/compare/v3.0.1...v3.0.2) (2026-08-10)


### Bug Fixes

* **deps:** update module golang.org/x/crypto to v0.52.0 [security] ([#130](https://github.com/martynvdijke/sandwitches/issues/130)) ([669bbb4](https://github.com/martynvdijke/sandwitches/commit/669bbb4ff3986d7804d855281a7aa41ab341d8ef))

# [3.0.1](https://github.com/martynvdijke/sandwitches/compare/v2.13.9...v3.0.1) (2026-08-09)


* feat!: release Go port replacing Django backend ([4f76dfb](https://github.com/martynvdijke/sandwitches/commit/4f76dfba39726cef103f09a93039de558e7949c6))


### Bug Fixes

* complete Go port with working templates, hot reload, and missing UI components ([8d3b157](https://github.com/martynvdijke/sandwitches/commit/8d3b157a9c7157c4567742605a4a91a5277c26a6))
* **docker:** remove COPY of empty go-app/locale dir ([8c463c3](https://github.com/martynvdijke/sandwitches/commit/8c463c3f903cd54b3df3c0a5f13a674c28b14275))
* ensure Gotify notification always fires on release workflow ([79513cd](https://github.com/martynvdijke/sandwitches/commit/79513cd2e89940cb4c0c36c215d14bf6ff3260c8))
* **go:** add CSRF token and nil checks to admin recipe form, add Playwright tests for Go port ([9bf26b4](https://github.com/martynvdijke/sandwitches/commit/9bf26b4bfc1e6732ade7356f34929e0d109934ca))
* invalid timezone UTC+1, use Europe/Amsterdam instead ([87f7c37](https://github.com/martynvdijke/sandwitches/commit/87f7c37ae033c4ab98cb28c0d3069471ef7cf86b))
* **release:** allow npm version to keep same version in prepareCmd ([6ebed66](https://github.com/martynvdijke/sandwitches/commit/6ebed66151df3a46c3301a28da87474d533c997a))
* **release:** install uv before semantic-release prepare step ([5c56960](https://github.com/martynvdijke/sandwitches/commit/5c56960731c4a386b5bd3689437e7a0481b0fc77))
* remove stalePr from renovate.json (no longer valid in Renovate v37) ([4828872](https://github.com/martynvdijke/sandwitches/commit/48288720b895d40ffed5b6c7d5e9b4f2250df224))
* remove stalePrAge from renovate.json (removed in Renovate v37) ([f780c4d](https://github.com/martynvdijke/sandwitches/commit/f780c4d151819cdc75ae66558a017654a7ae5eef))
* resolve CI failures - SECRET_KEY, Go vet, and action versions ([3db5379](https://github.com/martynvdijke/sandwitches/commit/3db5379cfa2582474864350a3c14a9489f935e24))
* **tests:** drain Go server output pipes to prevent 64KB pipe deadlock ([be07bde](https://github.com/martynvdijke/sandwitches/commit/be07bde3e8c242fa6bf8da90770bc78e3c844cc6))
* **ui:** add autocomplete attributes to auth forms (go-app + django) ([8e1ecd2](https://github.com/martynvdijke/sandwitches/commit/8e1ecd2faab1ca602a5551ac26c96d3d9623a551))
* use githubToken instead of otelToken for otel-cicd-action@v4 ([e8ee09a](https://github.com/martynvdijke/sandwitches/commit/e8ee09a1d2bd625d0c8b98c03d599f5e7ad0212f))
* visual parity with Django, CSRF HTML escaping, migration password conversion, GORM log silence, missing features ([8a479a2](https://github.com/martynvdijke/sandwitches/commit/8a479a251bb158b22d914c121ea6bf5afa8fb58f))


### Features

* add opentelemetry django instrumentation and CI test workflow ([39a6362](https://github.com/martynvdijke/sandwitches/commit/39a6362c889a105e3bf564e289ab2755716a05ff))
* add OTel endpoint admin configuration with DB-backed settings ([c53979c](https://github.com/martynvdijke/sandwitches/commit/c53979cf6f78484bf8b73cf39299da4763b0ba1b))
* add otlpAuthorization input for Bearer auth ([53311e0](https://github.com/martynvdijke/sandwitches/commit/53311e0b05008a2b4d257cf1652cce0dcbe51bc7))
* add sandwitches-go binary ([ffd4c75](https://github.com/martynvdijke/sandwitches/commit/ffd4c75c014bf464db72d6aa80583859236bda4d))
* automatic Django-to-GORM database migration ([dc92aea](https://github.com/martynvdijke/sandwitches/commit/dc92aeac8dc14cfcabe6c5e7abedd31dc65a2d8d))
* complete Go port with pagination, CSRF, flash messages, image upload, and missing features ([d90516f](https://github.com/martynvdijke/sandwitches/commit/d90516f62569e82544559c836e8269414d3422b6))
* feature parity with Django - quick order, filter UI, HTMX partials, cart validation, self-delete prevention ([556d4d1](https://github.com/martynvdijke/sandwitches/commit/556d4d1a4d141bef5850f2d9489f56471e082e6e))
* Go port - API, tasks, i18n, utils, tests, CI ([cb13cac](https://github.com/martynvdijke/sandwitches/commit/cb13cac13571ec3368b378e4a7e4184e9a34c89c))
* initial Go/Gin port - models, auth, handlers, templates, admin ([28a0f30](https://github.com/martynvdijke/sandwitches/commit/28a0f3025470a559cd287afbe2810e2f794bfeb7))
* port e-ink mode, cooking mode, and TRMNL templates to Go ([8eae437](https://github.com/martynvdijke/sandwitches/commit/8eae43773a1f44e6020dcb6ccdded5ac0f602dbd))
* rewrite admin templates with proper BeerCSS layout, Chart.js, HTMX ([7cce515](https://github.com/martynvdijke/sandwitches/commit/7cce515d3d584ab8b9c3a7d9113d4f9539b6d0b6))


### Reverts

* undo premature 3.0.0 version bump ([aa17e34](https://github.com/martynvdijke/sandwitches/commit/aa17e3457e25532aafeb048e3dc893479bd8251b))


### BREAKING CHANGES

* the Django backend (src/) is fully replaced by the Go
application (go-app/). The Docker image now runs the Go server on port
6270 with a /api/v1/ping healthcheck, Python/PyPI publishing is dropped,
and the e-ink/cooking modes plus TRMNL templates are served from the Go
stack. This marks the migration as a major release (3.0.1).

## [2.13.9](https://github.com/martynvdijke/sandwitches/compare/v2.13.8...v2.13.9) (2026-08-06)

## [2.13.8](https://github.com/martynvdijke/sandwitches/compare/v2.13.7...v2.13.8) (2026-08-04)


### Bug Fixes

* **deps:** update dependency beercss to v5 ([be011f3](https://github.com/martynvdijke/sandwitches/commit/be011f3722579196e4dbb5f2b8afcc00caa85d86))

## [2.13.7](https://github.com/martynvdijke/sandwitches/compare/v2.13.6...v2.13.7) (2026-07-26)

## [2.13.6](https://github.com/martynvdijke/sandwitches/compare/v2.13.5...v2.13.6) (2026-07-26)

## [2.13.5](https://github.com/martynvdijke/sandwitches/compare/v2.13.4...v2.13.5) (2026-07-25)

## [2.13.4](https://github.com/martynvdijke/sandwitches/compare/v2.13.3...v2.13.4) (2026-07-23)

## [2.13.3](https://github.com/martynvdijke/sandwitches/compare/v2.13.2...v2.13.3) (2026-07-21)

## [2.13.2](https://github.com/martynvdijke/sandwitches/compare/v2.13.1...v2.13.2) (2026-07-16)

## [2.13.1](https://github.com/martynvdijke/sandwitches/compare/v2.13.0...v2.13.1) (2026-07-14)

# [2.13.0](https://github.com/martynvdijke/sandwitches/compare/v2.12.26...v2.13.0) (2026-07-12)


### Bug Fixes

* **deps:** update all non-major dependencies ([f2e07b2](https://github.com/martynvdijke/sandwitches/commit/f2e07b299b9fe5f1c4e1182af956ee9539692cea))
* **deps:** update babel monorepo to v8 ([8afd9eb](https://github.com/martynvdijke/sandwitches/commit/8afd9eb172d6ed998f58657b0c7e25cc72df4240))
* remove duplicate Umami script tag in admin base template ([98b093b](https://github.com/martynvdijke/sandwitches/commit/98b093ba94a7eb53ac5ed02541cbc933449d89a0))


### Features

* add e-ink mode with high-contrast CSS, cooking mode, and toggle ([2b3c49b](https://github.com/martynvdijke/sandwitches/commit/2b3c49b58feb2d505d33af735cfd71e96596e24f)), closes [hi#contrast](https://github.com/hi/issues/contrast) [hi#contrast](https://github.com/hi/issues/contrast)
* remove experimental Flutter mobile app and its CI job ([a3f5c8d](https://github.com/martynvdijke/sandwitches/commit/a3f5c8d69843c12b6a1d7309df10904fcf3cdb90))

## [2.12.26](https://github.com/martynvdijke/sandwitches/compare/v2.12.25...v2.12.26) (2026-06-09)

## [2.12.25](https://github.com/martynvdijke/sandwitches/compare/v2.12.24...v2.12.25) (2026-06-04)

## [2.12.24](https://github.com/martynvdijke/sandwitches/compare/v2.12.23...v2.12.24) (2026-05-29)

## [2.12.23](https://github.com/martynvdijke/sandwitches/compare/v2.12.22...v2.12.23) (2026-05-27)

## [2.12.22](https://github.com/martynvdijke/sandwitches/compare/v2.12.21...v2.12.22) (2026-05-23)

## [2.12.21](https://github.com/martynvdijke/sandwitches/compare/v2.12.20...v2.12.21) (2026-05-11)

## [2.12.20](https://github.com/martynvdijke/sandwitches/compare/v2.12.19...v2.12.20) (2026-05-09)

## [2.12.19](https://github.com/martynvdijke/sandwitches/compare/v2.12.18...v2.12.19) (2026-05-06)

## [2.12.18](https://github.com/martynvdijke/sandwitches/compare/v2.12.17...v2.12.18) (2026-05-05)

## [2.12.17](https://github.com/martynvdijke/sandwitches/compare/v2.12.16...v2.12.17) (2026-05-05)

## [2.12.16](https://github.com/martynvdijke/sandwitches/compare/v2.12.15...v2.12.16) (2026-05-03)


### Bug Fixes

* invoke ci linting, typecheck, tests all passing ([ddc388f](https://github.com/martynvdijke/sandwitches/commit/ddc388ff2e20a971194f44e4c52a24a2c0be62be))
* lock file ([077465f](https://github.com/martynvdijke/sandwitches/commit/077465fe29da080b47e851a41aad9921da96f68a))
* playwright UI tests check profile link instead of img ([b1e7726](https://github.com/martynvdijke/sandwitches/commit/b1e7726ef9667b816b10d6687cb40524cb83ebb3))
* remove Instagram integration, improve UI with beercss, optimize Docker image ([86f555f](https://github.com/martynvdijke/sandwitches/commit/86f555fda7ff2856d6ec2ac934f7b850fa0e9a54))
* revert Dockerfile and .dockerignore to pre-Instagram-removal state ([5a536ba](https://github.com/martynvdijke/sandwitches/commit/5a536ba4897a0d5173d451298fb46282fcf1d2f3))

## [2.12.15](https://github.com/martynvdijke/sandwitches/compare/v2.12.14...v2.12.15) (2026-05-02)

## [2.12.14](https://github.com/martynvdijke/sandwitches/compare/v2.12.13...v2.12.14) (2026-04-29)

## [2.12.13](https://github.com/martynvdijke/sandwitches/compare/v2.12.12...v2.12.13) (2026-04-27)

## [2.12.12](https://github.com/martynvdijke/sandwitches/compare/v2.12.11...v2.12.12) (2026-04-22)

## [2.12.11](https://github.com/martynvdijke/sandwitches/compare/v2.12.10...v2.12.11) (2026-04-21)

## [2.12.10](https://github.com/martynvdijke/sandwitches/compare/v2.12.9...v2.12.10) (2026-04-20)

## [2.12.9](https://github.com/martynvdijke/sandwitches/compare/v2.12.8...v2.12.9) (2026-04-17)

## [2.12.8](https://github.com/martynvdijke/sandwitches/compare/v2.12.7...v2.12.8) (2026-04-16)

## [2.12.7](https://github.com/martynvdijke/sandwitches/compare/v2.12.6...v2.12.7) (2026-04-13)

## [2.12.6](https://github.com/martynvdijke/sandwitches/compare/v2.12.5...v2.12.6) (2026-04-11)

## [2.12.5](https://github.com/martynvdijke/sandwitches/compare/v2.12.4...v2.12.5) (2026-04-10)

## [2.12.4](https://github.com/martynvdijke/sandwitches/compare/v2.12.3...v2.12.4) (2026-04-10)


### Bug Fixes

* **chore:** formatting & testing issues ([bfb99e7](https://github.com/martynvdijke/sandwitches/commit/bfb99e7543a29bfa99ad3e0022b5e3c832dbda03))
* **chore:** formatting issues ([cd44921](https://github.com/martynvdijke/sandwitches/commit/cd44921cb0e9654f0c39493d0fd0b05f9bf71f91))
* instgram sync and logging ([21888c8](https://github.com/martynvdijke/sandwitches/commit/21888c8a26dd320788bb26a2be5887f4ae227e4c))

## [2.12.3](https://github.com/martynvdijke/sandwitches/compare/v2.12.2...v2.12.3) (2026-04-06)

## [2.12.2](https://github.com/martynvdijke/sandwitches/compare/v2.12.1...v2.12.2) (2026-04-05)


### Bug Fixes

* **ci:** mobile build ([5bb93df](https://github.com/martynvdijke/sandwitches/commit/5bb93df1d759079699a1cccce03c685854f5cd81))

## [2.12.1](https://github.com/martynvdijke/sandwitches/compare/v2.12.0...v2.12.1) (2026-04-05)


### Bug Fixes

* dart build ([715bc5d](https://github.com/martynvdijke/sandwitches/commit/715bc5db6e9dbcb85cf87e58674813a3ee8dc0cb))
* **mobile:** fix RefreshIndicator error and format code ([a282003](https://github.com/martynvdijke/sandwitches/commit/a2820033add3cd1d0ee56013b3fb83f925dbe648))

# [2.12.0](https://github.com/martynvdijke/sandwitches/compare/v2.11.9...v2.12.0) (2026-04-03)


### Bug Fixes

* **ui:** bugs in admin interface ([8134e06](https://github.com/martynvdijke/sandwitches/commit/8134e061cca5b736a456f7bb907ccab42e0d22f4))


### Features

* add throttling to instagram upload and update success message ([cd4a8de](https://github.com/martynvdijke/sandwitches/commit/cd4a8de21cb3422783293fe1509b76ad4f556b49))

## [2.11.9](https://github.com/martynvdijke/sandwitches/compare/v2.11.8...v2.11.9) (2026-04-03)

## [2.11.8](https://github.com/martynvdijke/sandwitches/compare/v2.11.7...v2.11.8) (2026-03-30)

## [2.11.7](https://github.com/martynvdijke/sandwitches/compare/v2.11.6...v2.11.7) (2026-03-28)

## [2.11.6](https://github.com/martynvdijke/sandwitches/compare/v2.11.5...v2.11.6) (2026-03-24)

## [2.11.5](https://github.com/martynvdijke/sandwitches/compare/v2.11.4...v2.11.5) (2026-03-23)

## [2.11.4](https://github.com/martynvdijke/sandwitches/compare/v2.11.3...v2.11.4) (2026-03-22)


### Bug Fixes

* instagram + mobile fixes ([0fb3c80](https://github.com/martynvdijke/sandwitches/commit/0fb3c8002a6fc657aeb694a1fecc6ceafba519d7))

## [2.11.3](https://github.com/martynvdijke/sandwitches/compare/v2.11.2...v2.11.3) (2026-03-22)


### Bug Fixes

* **deps:** update dependency intl to v0.20.2 ([368f684](https://github.com/martynvdijke/sandwitches/commit/368f684bcdbcdc60a1cc9c8fc6dc462b6fb9c020))

## [2.11.2](https://github.com/martynvdijke/sandwitches/compare/v2.11.1...v2.11.2) (2026-03-19)


### Bug Fixes

* add in sync instagram button ([f084d04](https://github.com/martynvdijke/sandwitches/commit/f084d0419cb61ebca1c1d88451ca1dd687a72608))

## [2.11.1](https://github.com/martynvdijke/sandwitches/compare/v2.11.0...v2.11.1) (2026-03-18)

# [2.11.0](https://github.com/martynvdijke/sandwitches/compare/v2.10.2...v2.11.0) (2026-03-18)


### Bug Fixes

* dashboard and add more logging ([f1195cd](https://github.com/martynvdijke/sandwitches/commit/f1195cd94f8185be713e805692b64a6036fb5e12))
* instagram save ([6659a6a](https://github.com/martynvdijke/sandwitches/commit/6659a6ac56b890031a33989afbad02eb78623d60))
* **ty:** fix ty for tests ([b900a52](https://github.com/martynvdijke/sandwitches/commit/b900a5211b25185494e5ae2065540368d889d458))


### Features

* add in recipe schema ([0ebef8b](https://github.com/martynvdijke/sandwitches/commit/0ebef8bb9afbea15a4fbeafa0d63d741df6e83a0))

## [2.10.2](https://github.com/martynvdijke/sandwitches/compare/v2.10.1...v2.10.2) (2026-03-18)

## [2.10.1](https://github.com/martynvdijke/sandwitches/compare/v2.10.0...v2.10.1) (2026-03-13)

# [2.10.0](https://github.com/martynvdijke/sandwitches/compare/v2.9.10...v2.10.0) (2026-03-11)


### Bug Fixes

* **formatting:** fix formatting ([b75ea0b](https://github.com/martynvdijke/sandwitches/commit/b75ea0b46b3f19bce5322c8fb5521ab639cdae55))


### Features

* add in instagram to sync recipes images with instagram ([0226561](https://github.com/martynvdijke/sandwitches/commit/022656128c6c275a41e334346f0071666d1f8b38))

## [2.9.10](https://github.com/martynvdijke/sandwitches/compare/v2.9.9...v2.9.10) (2026-03-08)

## [2.9.9](https://github.com/martynvdijke/sandwitches/compare/v2.9.8...v2.9.9) (2026-03-08)

## [2.9.8](https://github.com/martynvdijke/sandwitches/compare/v2.9.7...v2.9.8) (2026-03-06)

## [2.9.7](https://github.com/martynvdijke/sandwitches/compare/v2.9.6...v2.9.7) (2026-03-04)

## [2.9.6](https://github.com/martynvdijke/sandwitches/compare/v2.9.5...v2.9.6) (2026-03-03)

## [2.9.5](https://github.com/martynvdijke/sandwitches/compare/v2.9.4...v2.9.5) (2026-02-28)

## [2.9.4](https://github.com/martynvdijke/sandwitches/compare/v2.9.3...v2.9.4) (2026-02-25)

## [2.9.3](https://github.com/martynvdijke/sandwitches/compare/v2.9.2...v2.9.3) (2026-02-23)

## [2.9.2](https://github.com/martynvdijke/sandwitches/compare/v2.9.1...v2.9.2) (2026-02-22)


### Bug Fixes

* **ci:** fix package write ([180a9ff](https://github.com/martynvdijke/sandwitches/commit/180a9ffb42afbbf7a29d5886b93672ee9a449918))

## [2.9.1](https://github.com/martynvdijke/sandwitches/compare/v2.9.0...v2.9.1) (2026-02-21)


### Bug Fixes

* **ci:** correct repo name ([f6f0e56](https://github.com/martynvdijke/sandwitches/commit/f6f0e569b3101f8e813e985e2c36de76e5c66d40))

# [2.9.0](https://github.com/martynvdijke/sandwitches/compare/v2.8.2...v2.9.0) (2026-02-21)


### Bug Fixes

* **ci:** add in write packages v2 ([9e60b63](https://github.com/martynvdijke/sandwitches/commit/9e60b63f75a892fbe40947fa721230f8682959bb))


### Features

* add umami anaytlics tracking ([2d263fb](https://github.com/martynvdijke/sandwitches/commit/2d263fbc010b2040af0fa8e5e67c495752961385))

## [2.8.2](https://github.com/martynvdijke/sandwitches/compare/v2.8.1...v2.8.2) (2026-02-21)


### Bug Fixes

* **ci:** add in write packages ([4ea6057](https://github.com/martynvdijke/sandwitches/commit/4ea6057c1acda65ddc0fb20050a74583fc995017))

## [2.8.1](https://github.com/martynvdijke/sandwitches/compare/v2.8.0...v2.8.1) (2026-02-21)


### Bug Fixes

* **ci:** also push to ghcr ([371b042](https://github.com/martynvdijke/sandwitches/commit/371b0428a1f8e1b122bd2e0ad514c16d14a0d2fb))
* **ci:** remove codeql ([eb38375](https://github.com/martynvdijke/sandwitches/commit/eb383752fcdbc881a20fcd2d66400731d58215fd))

# [2.8.0](https://github.com/martynvdijke/sandwitches/compare/v2.7.5...v2.8.0) (2026-02-20)


### Bug Fixes

* enable all ouput in api for get recipes ([1c1263d](https://github.com/martynvdijke/sandwitches/commit/1c1263d2a740dcbe8483c634921d7913115bc23a))
* remove login auth ([ff2272a](https://github.com/martynvdijke/sandwitches/commit/ff2272ac690799bb30fdae78682d00f8b9a1ebe0))
* remove login auth ([9ee1b33](https://github.com/martynvdijke/sandwitches/commit/9ee1b33d1bdb494c5f1e401150d45b720fedb5ae))


### Features

* **ci:** Refactor CodeQL workflow by cleaning up comments ([8d02c63](https://github.com/martynvdijke/sandwitches/commit/8d02c63aa3e43ea7d5bd205448c5cb14a1489c94))

## [2.7.6](https://github.com/martynvdijke/sandwitches/compare/v2.7.5...v2.7.6) (2026-02-20)


### Bug Fixes

* enable all ouput in api for get recipes ([1c1263d](https://github.com/martynvdijke/sandwitches/commit/1c1263d2a740dcbe8483c634921d7913115bc23a))
* remove login auth ([ff2272a](https://github.com/martynvdijke/sandwitches/commit/ff2272ac690799bb30fdae78682d00f8b9a1ebe0))
* remove login auth ([9ee1b33](https://github.com/martynvdijke/sandwitches/commit/9ee1b33d1bdb494c5f1e401150d45b720fedb5ae))

## [2.7.5](https://github.com/martynvdijke/sandwitches/compare/v2.7.4...v2.7.5) (2026-02-20)

## [2.7.4](https://github.com/martynvdijke/sandwitches/compare/v2.7.3...v2.7.4) (2026-02-19)

## [2.7.3](https://github.com/martynvdijke/sandwitches/compare/v2.7.2...v2.7.3) (2026-02-11)


### Bug Fixes

* fix cropperv2 ([2f041f3](https://github.com/martynvdijke/sandwitches/commit/2f041f3503a7651a0d9da9cbdb28b614377d3277))

## [2.7.2](https://github.com/martynvdijke/sandwitches/compare/v2.7.1...v2.7.2) (2026-02-11)

## [2.7.1](https://github.com/martynvdijke/sandwitches/compare/v2.7.0...v2.7.1) (2026-02-11)

# [2.7.0](https://github.com/martynvdijke/sandwitches/compare/v2.6.5...v2.7.0) (2026-02-11)


### Features

* update api to be more feature complete add easymde: ([e522315](https://github.com/martynvdijke/sandwitches/commit/e52231567df377de19304b8f38c90ce44e99435d))

## [2.6.5](https://github.com/martynvdijke/sandwitches/compare/v2.6.4...v2.6.5) (2026-02-10)

## [2.6.4](https://github.com/martynvdijke/sandwitches/compare/v2.6.3...v2.6.4) (2026-02-09)

## [2.6.3](https://github.com/martynvdijke/sandwitches/compare/v2.6.2...v2.6.3) (2026-02-07)

## [2.6.2](https://github.com/martynvdijke/sandwitches/compare/v2.6.1...v2.6.2) (2026-02-06)


### Bug Fixes

* resolve migration conflict by populating unique tracking tokens for orders ([d83cdb7](https://github.com/martynvdijke/sandwitches/commit/d83cdb76a5a125022ed73c23372a2900476750c4))

## [2.6.1](https://github.com/martynvdijke/sandwitches/compare/v2.6.0...v2.6.1) (2026-02-05)

# [2.6.0](https://github.com/martynvdijke/sandwitches/compare/v2.5.6...v2.6.0) (2026-02-05)


### Features

* complete personal order tracker with email confirmation and status tracking ([4c3df37](https://github.com/martynvdijke/sandwitches/commit/4c3df376a1ed46a8067f41586fb222af72e0fafb))

## [2.5.6](https://github.com/martynvdijke/sandwitches/compare/v2.5.5...v2.5.6) (2026-02-04)

## [2.5.5](https://github.com/martynvdijke/sandwitches/compare/v2.5.4...v2.5.5) (2026-02-03)


### Bug Fixes

* **ci:** clean up renovate config ([04795a1](https://github.com/martynvdijke/sandwitches/commit/04795a1c4bf32325531b4cc2327a21f50de77cbe))
* shopping cart supports multiple sandwithces ([1b77c1c](https://github.com/martynvdijke/sandwitches/commit/1b77c1ca05fd5bed6bbbdabf11ccd4aff241c0f7))

## [2.5.4](https://github.com/martynvdijke/sandwitches/compare/v2.5.3...v2.5.4) (2026-01-30)


### Bug Fixes

* **ci:** push docker image on release ([7e4c2dd](https://github.com/martynvdijke/sandwitches/commit/7e4c2dd3f5b0d23ef75d878ec4d1ac7ee6974574))

## [2.5.3](https://github.com/martynvdijke/sandwitches/compare/v2.5.2...v2.5.3) (2026-01-29)


### Bug Fixes

* **ci:** disable docker container ([e0d7eb3](https://github.com/martynvdijke/sandwitches/commit/e0d7eb3e681a7a22257068e427078120eeeb63ee))
* **ui:** add tabeed interface ([eeb346f](https://github.com/martynvdijke/sandwitches/commit/eeb346faaa91c15fbc35b0bddd7e2e7b71c4c911))

## [2.5.2](https://github.com/martynvdijke/sandwitches/compare/v2.5.1...v2.5.2) (2026-01-29)


### Bug Fixes

* **ci:** add back in ci tests and update renovate ([852f987](https://github.com/martynvdijke/sandwitches/commit/852f98719565cd8e9a0108d338555167a6d41a94))
* **ci:** update ci checks to run on all pr's ([4b2ac36](https://github.com/martynvdijke/sandwitches/commit/4b2ac36f07021a8a1f75aeefcfbd024c83ce8e2d))
* **ci:** update renovate with correct config ([82ab090](https://github.com/martynvdijke/sandwitches/commit/82ab0909dc3eb70ba9aa0fcc4e612c0f7c428b35))
* cropper js import ([b05d41f](https://github.com/martynvdijke/sandwitches/commit/b05d41f491783e15f095b6fb0fb4e76c77e1b4bd))
* **deps:** update dependency beercss to v4 ([3088299](https://github.com/martynvdijke/sandwitches/commit/3088299b2756965cab6c7e61583b49a0b3aa8013))
* **deps:** update dependency cropperjs to v2 ([15a35f3](https://github.com/martynvdijke/sandwitches/commit/15a35f3e3e193902ee98ff742794d5f192c17cb2))
* release workflow and settings fix ([38f58ce](https://github.com/martynvdijke/sandwitches/commit/38f58ceb287f8f6b0a01e2c447611985c2134e24))
* release workflow and settings fix ([6009b39](https://github.com/martynvdijke/sandwitches/commit/6009b3976e6332a9c6984db0642e9ee688ce0645))
* tests ([f20962f](https://github.com/martynvdijke/sandwitches/commit/f20962f4fc0bb31f8af30bc73ff5ffefc2ca99f1))
* tests & add request depedency ([1d3f451](https://github.com/martynvdijke/sandwitches/commit/1d3f451db9ba5d83474293a959d29191c9d95fb8))
* **ui:** make admin tables bigger and more readable ([15948b6](https://github.com/martynvdijke/sandwitches/commit/15948b65872027d850e6999f7ed1275f50988859))

## [2.5.1](https://github.com/martynvdijke/sandwitches/compare/v2.5.0...v2.5.1) (2026-01-28)


### Bug Fixes

* renovate config ([463f015](https://github.com/martynvdijke/sandwitches/commit/463f0153c8e8bc5998d5a5b4d7fe8a9d540d761b))
* renovate config v2 ([8afcf61](https://github.com/martynvdijke/sandwitches/commit/8afcf615582438c159ade2518372932779bc2231))
* renovate config v3 ([9470a87](https://github.com/martynvdijke/sandwitches/commit/9470a87a9b9b2181bfa251e938f91eb59054c3f8))
* renovate config v4 ([242b0cd](https://github.com/martynvdijke/sandwitches/commit/242b0cded29b3e6ccaae8f435af67cf6dac13969))

# [2.5.0](https://github.com/martynvdijke/sandwitches/compare/v2.4.1...v2.5.0) (2026-01-28)


### Bug Fixes

* add user config panel to clean up ui ([501a014](https://github.com/martynvdijke/sandwitches/commit/501a014b7161ae36bb8abbe7203208453ef19cfe))


### Features

* add gotify notifications ([8965e20](https://github.com/martynvdijke/sandwitches/commit/8965e20778d590afe9b6858d408e85b0659c8201))
* **ci:** add noreavate support ([efacab0](https://github.com/martynvdijke/sandwitches/commit/efacab07c3ca126368be9847b6813c98c33f9e30))

## [2.4.1](https://github.com/martynvdijke/sandwitches/compare/v2.4.0...v2.4.1) (2026-01-27)


### Bug Fixes

* **ui:** add in photo cropper support ([131908b](https://github.com/martynvdijke/sandwitches/commit/131908b67a803fc97d752628595c74b1b33baa24))
* **web:** add orders history and have order history and order status ([66202d0](https://github.com/martynvdijke/sandwitches/commit/66202d01456c70cacd73730aea7091d15ad4ba60))
* **web:** serve dist and css ourselves ([1e72961](https://github.com/martynvdijke/sandwitches/commit/1e729615df399a626cf7361675bd7668633d395d))

# [2.4.0](https://github.com/martynvdijke/sandwitches/compare/v2.3.3...v2.4.0) (2026-01-25)


### Bug Fixes

* **tests:** skip flaky test ([b502a19](https://github.com/martynvdijke/sandwitches/commit/b502a190bfa07df02c1e77c2c28cbcf0bae4fb94))


### Features

* add full order flow along with shopping cart and backend commnnity recipes optimalizations ([99e0708](https://github.com/martynvdijke/sandwitches/commit/99e07089555e5c7ff35acf63ad08a47238064928))

## [2.3.3](https://github.com/martynvdijke/sandwitches/compare/v2.3.2...v2.3.3) (2026-01-24)


### Bug Fixes

* remove static main css in admin dashboard ([fd20fec](https://github.com/martynvdijke/sandwitches/commit/fd20fecebef9eb0ed665bfb8cb204856aebd3ee1))

## [2.3.2](https://github.com/martynvdijke/sandwitches/compare/v2.3.1...v2.3.2) (2026-01-24)


### Bug Fixes

* **ui:** dashboard charts not rendering by loading Chart.js in head ([7b8a73a](https://github.com/martynvdijke/sandwitches/commit/7b8a73a234a762380f18ebce3ba89912dcc5b726))
* **ui:** fix ui and tests ([4b8c93c](https://github.com/martynvdijke/sandwitches/commit/4b8c93cdfe6e0990009f2dfee73f75c62dcd4c63))
* **ui:** move recipe submission to a dedicated 'Community' page ([46992d6](https://github.com/martynvdijke/sandwitches/commit/46992d6146702735cb4bd57dc8a764cbc1378330))
* **ui:** separate normal and community recipes on index and community pages ([c287cdd](https://github.com/martynvdijke/sandwitches/commit/c287cdd8f85da5e25f880a12809f66e75bdebc7c))

## [2.3.1](https://github.com/martynvdijke/sandwitches/compare/v2.3.0...v2.3.1) (2026-01-22)


### Bug Fixes

* **admin:** admin dashboard fixes ([62fe19a](https://github.com/martynvdijke/sandwitches/commit/62fe19aa6851bbb2948271b674da1d5f3c9c5084))
* **backend:** add in bassis for recipe ordering ([7c2a7df](https://github.com/martynvdijke/sandwitches/commit/7c2a7dffbefb855e25c7fe7eaf45d8072b46c6f4))
* orders backend and notification ([513dd0d](https://github.com/martynvdijke/sandwitches/commit/513dd0d4fcd372f64378b88e6e0425ffef41723a))
* **web:** ui base for login ([4a41807](https://github.com/martynvdijke/sandwitches/commit/4a418074beca5666aafa161aa96aecbec9300753))

# [2.3.0](https://github.com/martynvdijke/sandwitches/compare/v2.2.0...v2.3.0) (2026-01-20)


### Bug Fixes

* **web:** add in profile page settings ([b2abed2](https://github.com/martynvdijke/sandwitches/commit/b2abed284e3d6da2b1c7f658bce3edd964816cd2))


### Features

* add orders to recipes ([32afdb2](https://github.com/martynvdijke/sandwitches/commit/32afdb2bb2485e29c9935a4de5f31a1d09b48a3c))

# [2.2.0](https://github.com/martynvdijke/sandwitches/compare/v2.1.2...v2.2.0) (2026-01-17)


### Bug Fixes

* prevent carousel items overlap by forcing flex-shrink 0 and explicit width ([edf6d0c](https://github.com/martynvdijke/sandwitches/commit/edf6d0c81d581758fdde64198a9c603e1989bdc0))


### Features

* implement highlighted sandwiches carousel on homepage ([e8e12ae](https://github.com/martynvdijke/sandwitches/commit/e8e12aec2be8b51644ffef2d5974ffa85052175c))
* improve highlighted recipe carousel with auto-scroll and navigation ([27fc4fa](https://github.com/martynvdijke/sandwitches/commit/27fc4fae6e76ad176c21ed6e63bd32cd3973a54b))
* **ui:** implement recipe highlight and carasol ([f288b9e](https://github.com/martynvdijke/sandwitches/commit/f288b9e902a578b9c5e6ece0d98ec99c96cdd611))

## [2.1.2](https://github.com/martynvdijke/sandwitches/compare/v2.1.1...v2.1.2) (2026-01-15)


### Bug Fixes

* **admin:** invalide photo rotation cache ([4ee8d69](https://github.com/martynvdijke/sandwitches/commit/4ee8d6953fc78ef7ae7c49cb1774e8c9b6e572ba))
* **web:** keep recipe card the same height ([b13cae3](https://github.com/martynvdijke/sandwitches/commit/b13cae31ec77927e3f2efd0b7bd38525270b05b6))
* **web:** use own login instead of django default login ([c307da6](https://github.com/martynvdijke/sandwitches/commit/c307da65c45fa68d43c9209e7f5dabb063fafab8))

## [2.1.1](https://github.com/martynvdijke/sandwitches/compare/v2.1.0...v2.1.1) (2026-01-15)


### Bug Fixes

* **ci:** add in pre commit hooks ([2bec5be](https://github.com/martynvdijke/sandwitches/commit/2bec5be828cef8c0dd2c2e962bc840af0cb8ca21))
* **ci:** update pre-commit to include all files ([5e5cc94](https://github.com/martynvdijke/sandwitches/commit/5e5cc94eb1b1885a33fac841803e72fb98df3979))

# [2.1.0](https://github.com/martynvdijke/sandwitches/compare/v2.0.0...v2.1.0) (2026-01-15)


### Bug Fixes

* **ci:** set django allowed async in task env ([c944e38](https://github.com/martynvdijke/sandwitches/commit/c944e38e4738e9ed4a22eee2785e0521469e9912))
* **web:** use our login instead of django default login ([1de3e74](https://github.com/martynvdijke/sandwitches/commit/1de3e744b8daf019c32e585f3a27f217095a681f))


### Features

* **tests:** add basiss playwright tests ([4525eef](https://github.com/martynvdijke/sandwitches/commit/4525eefed2dfaa7a7814b517160743014cf79297))

# [2.0.0](https://github.com/martynvdijke/sandwitches/compare/v1.4.2...v2.0.0) (2026-01-14)


### Bug Fixes

* add build web to task ([0c56c39](https://github.com/martynvdijke/sandwitches/commit/0c56c39d2074a34270caff13dc6734c29966ec2a))
* **admin:** css ([472dbdc](https://github.com/martynvdijke/sandwitches/commit/472dbdc85f7ffe658ec6ec3491589d443cab45aa))
* go abck to time before npm ([a8913d1](https://github.com/martynvdijke/sandwitches/commit/a8913d1afcfff93a0b1f64bc812563cb2563ec22))
* search ui fix ([e291e67](https://github.com/martynvdijke/sandwitches/commit/e291e67d5142b352d169102444fcc38472323561))
* **tests:** fix formatting and tests ([6b719cf](https://github.com/martynvdijke/sandwitches/commit/6b719cff93b71766073085f3b21b9743132af441))
* **ui:** numerous bug fixes ([b2abf2a](https://github.com/martynvdijke/sandwitches/commit/b2abf2a0fa4632fb2a8ebffd79a849ed9ca271b3))


### Features

* add favourites panel and fix search dropdown ([de559ac](https://github.com/martynvdijke/sandwitches/commit/de559aca1841d7a99924ca400b00c4a475a8694e))
* add in click on person and click on tag view ([f90214f](https://github.com/martynvdijke/sandwitches/commit/f90214f1973d2bcd7e77b6fda20be68e2f743174))
* add in recipe serving options ([a7f2f10](https://github.com/martynvdijke/sandwitches/commit/a7f2f101bf952f1a7d16c9a1ca31cc164365aac5))
* add in settings as singleton to db ([4a13099](https://github.com/martynvdijke/sandwitches/commit/4a13099205cca0fb13bb3dc618d3135b7138abae))
* Add singleton settings model and webpack for static assets ([5e00729](https://github.com/martynvdijke/sandwitches/commit/5e007294390e9daf163b0a146e65121a08028ca2))
* add super user hotkeys in ([a8a6079](https://github.com/martynvdijke/sandwitches/commit/a8a60797e2f541c7cac7b560a5bd3b51944e3025))
* atom/rss feed ([460dfb3](https://github.com/martynvdijke/sandwitches/commit/460dfb3a3249ba802976037840b9cceec633ef27))
* Display likers and allow comments on ratings ([e6f5973](https://github.com/martynvdijke/sandwitches/commit/e6f59735445ec8ebb6fadd9e7417c677a61b4fa9))
* move over to beercss revamp ui ([3d6bf9f](https://github.com/martynvdijke/sandwitches/commit/3d6bf9f0d1d8bf342caf4f45b8106b208dc2da3d))
* revamp search now with htmx to search more and better in a dynamic fashion ([3971a3f](https://github.com/martynvdijke/sandwitches/commit/3971a3f0d6103c817430fb37d50f239f2d25c50f))
* **tests:** add in more tests ([300e488](https://github.com/martynvdijke/sandwitches/commit/300e48846e48e4d7db476fa482226f9eeb16b129))
* **ui:** add in animation loader ([ebe2ae3](https://github.com/martynvdijke/sandwitches/commit/ebe2ae333e772b435f11c93e1bbcb808c0dadce1))
* **ui:** Improve UI consistency and admin page functionality ([ee521be](https://github.com/martynvdijke/sandwitches/commit/ee521be35fbd3a3a6826a7588cd0bd818654d3fd))
* use npm for css and js versioning ([44f816b](https://github.com/martynvdijke/sandwitches/commit/44f816b40b515a03513dac241e9614a9b00c9402))
* **web:** add in new dashboard overview graph ([146bcd8](https://github.com/martynvdijke/sandwitches/commit/146bcd8a0945dcf36f4b598f0bfdde7f85f582cd))


### Performance Improvements

* revamp ui with beercss and move users into new db structure ([32ea240](https://github.com/martynvdijke/sandwitches/commit/32ea240dcd2d08838245833e3cad30a29b78d4bf))


### BREAKING CHANGES

* revamp ui


## [1.4.2](https://github.com/martynvdijke/sandwitches/compare/v1.4.1...v1.4.2) (2026-01-03)


### Bug Fixes

* add supervised ini ([1044d2a](https://github.com/martynvdijke/sandwitches/commit/1044d2abd7bc03fb04ca3822653b1586d02d2724))
* correct user ([6350e52](https://github.com/martynvdijke/sandwitches/commit/6350e528f4a8ea1fd9a7b5494bd10846b1f688c0))
* **email:** correct email handling instead of incorrect bcc ([36b4b4c](https://github.com/martynvdijke/sandwitches/commit/36b4b4cc4c123e8338151570dd4d4bd32aa902ba))
* **formatting:** profile delete ([c88cd88](https://github.com/martynvdijke/sandwitches/commit/c88cd886247797844b373bca6d47fcc90539387a))
* make profile delition migration ([16ded4e](https://github.com/martynvdijke/sandwitches/commit/16ded4e8354eb8fa76835b70745cee8436e05a1e))

## [1.4.1](https://github.com/martynvdijke/sandwitches/compare/v1.4.0...v1.4.1) (2026-01-03)


### Bug Fixes

* **image:** fix starup procedure for worker node ([1c59e58](https://github.com/martynvdijke/sandwitches/commit/1c59e5840a70db237ea76e8ac635b079019cceec))

# [1.4.0](https://github.com/martynvdijke/sandwitches/compare/v1.3.1...v1.4.0) (2026-01-01)


### Features

* add django import export ([954e66a](https://github.com/martynvdijke/sandwitches/commit/954e66ab0eccefd1c31896e18a5155b12882c690))

## [1.3.1](https://github.com/martynvdijke/sandwitches/compare/v1.3.0...v1.3.1) (2025-12-31)


### Bug Fixes

* **ci:** remove old poetry.lock file and old vscode settings ([b580a2a](https://github.com/martynvdijke/sandwitches/commit/b580a2ac795dff6ce1c8968f81f7e52228c761c0))
* formatting of translations ([b1a96dc](https://github.com/martynvdijke/sandwitches/commit/b1a96dc313ff5e91e118a503dc523bd421c71a92))
* translations also translate email ([7f21c27](https://github.com/martynvdijke/sandwitches/commit/7f21c272f32d47da399c954d091c7cb35530e370))

# [1.3.0](https://github.com/martynvdijke/sandwitches/compare/v1.2.0...v1.3.0) (2025-12-29)


### Bug Fixes

* collect static using no input and clear always ([d74acca](https://github.com/martynvdijke/sandwitches/commit/d74accac8013320608bad3ce4f930f89ff385098))


### Features

* app recipe rating to the api ([20d70e4](https://github.com/martynvdijke/sandwitches/commit/20d70e4d9a983f034ad03b6d4e9e25e9f1c1d166))

# [1.2.0](https://github.com/martynvdijke/sandwitches/compare/v1.1.0...v1.2.0) (2025-12-29)


### Bug Fixes

* **ci:** setup ci settings ([4c11a82](https://github.com/martynvdijke/sandwitches/commit/4c11a823c6213dd13e6a9fd5a1e9182f2eeac829))
* collect static for docker image ([3a94c16](https://github.com/martynvdijke/sandwitches/commit/3a94c16c525165cb9cf74af47fbf6dbf35d2202e))
* **deps:** add in svg icons and banners ([ac2d7db](https://github.com/martynvdijke/sandwitches/commit/ac2d7dbfc842a25984319b80ff35c607da4fd1f2))
* fix correct model for tags schema ([217b4b4](https://github.com/martynvdijke/sandwitches/commit/217b4b453b38ab3ff41e39c7279827ee98ed92d9))
* tags api use ([87c7c16](https://github.com/martynvdijke/sandwitches/commit/87c7c164583f19d85f9bed4ae8f40163d386729e))
* Update tasks.py with collect static ([898370b](https://github.com/martynvdijke/sandwitches/commit/898370bac71b3ec917f087fd4b5d2090bde98a02))


### Features

* add image resizing and profile model ([09adec9](https://github.com/martynvdijke/sandwitches/commit/09adec93191ab40db082386892b151021f56a051))
* add in tags and users to api ([d9e35de](https://github.com/martynvdijke/sandwitches/commit/d9e35deaa6097680012e9fbe28dfeb22021af0b5))
* add initial internalization support ([c7d2b9e](https://github.com/martynvdijke/sandwitches/commit/c7d2b9e2c718ce2829e24ca3c6613f8427a9d163))
* initial email support ([acbe6b8](https://github.com/martynvdijke/sandwitches/commit/acbe6b80262c50326fb4e4477973914e9cd9e765))
* move to uvicorn & gunicorn for http serving ([55dbc4a](https://github.com/martynvdijke/sandwitches/commit/55dbc4a19f2e9d9d78d15fd1193eb0821655aeb5))
* serve files staticlly ([e93b19f](https://github.com/martynvdijke/sandwitches/commit/e93b19ff7aabedc3d30de7c71e84a639a86ea6eb))
* try to implement django task backend for email notifications ([d7f107f](https://github.com/martynvdijke/sandwitches/commit/d7f107f2aebf3808cc8498dacaea7b707bacfeb9))
* update ty checks and make code ty compatible ([325a7ec](https://github.com/martynvdijke/sandwitches/commit/325a7ec69f6c14ba14f72661bb49ab92dd61baf2))

# [1.1.0](https://github.com/martynvdijke/sandwitches/compare/v1.0.1...v1.1.0) (2025-12-23)


### Features

* add in basic api ([3eb0772](https://github.com/martynvdijke/sandwitches/commit/3eb07720310d6d3fa23bef566cc29e197e94646e))
* add in recipe rating ([2a62edd](https://github.com/martynvdijke/sandwitches/commit/2a62eddcaa277224cf4a06a942d58a38bdeb6d73))
* add in user signup ([6ddae0e](https://github.com/martynvdijke/sandwitches/commit/6ddae0e236b0288ddf731d0d0f085af8e14cd7ee))

## [1.0.1](https://github.com/martynvdijke/sandwitches/compare/v1.0.0...v1.0.1) (2025-11-12)


### Bug Fixes

* add uv lock for release ([b30ffcc](https://github.com/martynvdijke/sandwitches/commit/b30ffcc587dda35c9c2b753d31842e763cc42e1e))
* awk quotes ([1eef1b6](https://github.com/martynvdijke/sandwitches/commit/1eef1b67eafbe3382366adb29c9febaf1b950813))
* fix final push of docker container ([659d7c1](https://github.com/martynvdijke/sandwitches/commit/659d7c18d0bb9b0371c6e1c62f3b94ff544c92ed))
* not escaping quotes ([6ee2bd9](https://github.com/martynvdijke/sandwitches/commit/6ee2bd9363168fe333e9478cf3062960ccf0f8d2))
* quotes charachters ([93ca515](https://github.com/martynvdijke/sandwitches/commit/93ca515234570cb638f168a6282e66ffbfb84064))
* rerun uv lock ([1728c7e](https://github.com/martynvdijke/sandwitches/commit/1728c7e58bb0317356ae62bd97c2dc16968c753a))

# 1.0.0 (2025-11-12)


### Bug Fixes

* docker command entry ([d7610f9](https://github.com/martynvdijke/sandwitches/commit/d7610f94d5087205714d02456bf2963169b829e6))
* enable writing of github actions ([f9e63df](https://github.com/martynvdijke/sandwitches/commit/f9e63df342e8bf6ca405dca87ace8aef49f8b7c4))
* redirect to setup on initial page load ([943b3ac](https://github.com/martynvdijke/sandwitches/commit/943b3ac4fd869d102137b809c6bc8911aa6c3672))
* run uv dev group ([acd3725](https://github.com/martynvdijke/sandwitches/commit/acd372541b96019dadffa0b79117b51fd93e72d5))
* run uv dev group ([def5438](https://github.com/martynvdijke/sandwitches/commit/def5438fe8202b397e65fc980d04cbb038eac268))
* run uv dev group ([7687ade](https://github.com/martynvdijke/sandwitches/commit/7687ade48927fd56620d16446281db3f8baab2b8))
* run uv dev group ([6c8a975](https://github.com/martynvdijke/sandwitches/commit/6c8a975214c1a826e51f99b79e1aad23dff18ba6))


### Features

* add docker file and build ([82dd067](https://github.com/martynvdijke/sandwitches/commit/82dd067d2813c58669c45ac6661cf65ca7111d59))
* add in github workflow ([4f448c2](https://github.com/martynvdijke/sandwitches/commit/4f448c2d1bab12d23aecd98032e8c4922b39b7a9))
* add markdown rendering add first time setup ([4c57a91](https://github.com/martynvdijke/sandwitches/commit/4c57a910cd23a84cb9932618963dc350e52b647e))
* added static tests recipies ([719ab7c](https://github.com/martynvdijke/sandwitches/commit/719ab7cb5e40a8e1bfbbb4168aa053dc5608554f))
* first session with reflex basic working website ([c474f23](https://github.com/martynvdijke/sandwitches/commit/c474f23302440f3fee55322ad5c431359614b02a))
* first try add db connection, storing getting from database ([d2c0aed](https://github.com/martynvdijke/sandwitches/commit/d2c0aed41bbe65fc1b141bb08cf552ce4c17e624))
* intial alpha version of django ([4b9fa20](https://github.com/martynvdijke/sandwitches/commit/4b9fa20c5a8c6c6efd4321d36350f3d70552eb1a))
* move to django ([e8a52d2](https://github.com/martynvdijke/sandwitches/commit/e8a52d20568b2c6716895ec96a9f4fd6e5754ee1))
