# 0.15.1 (Jun 2, 2023)

BUGFIXES:

* build(goreleaser): fix regression and set version on released artifact (#316)

# 0.15.0 (Jun 2, 2023)

NOTES: this release includes a complete rewrite of the `honeycombio_trigger` resource: migrating it from the Terraform Plugin SDKv2 to the new Plugin Framework.
This was done to fix a number of long-standing bugs related to the `recipient` block.

This migration has resulted in some subtle, but non-breaking side effects:

* after updating, the next "plan" will show all trigger recipients being updated in-place
  * at the core of most all of these bugs was that fact that all of `id`, `type`, and `target` for a recipient were being stored in state. Now only `id` or the `type`+`target` pair will be stored in the state and the plan output should reflect this.
* enforcement of only specifying one of `id` or `type`+`target` is now possible due to the new flexibility gained by migrating to the Plugin Framework. Due to the shape of the recipient blocks in the schema, this validation was not possible with the Plugin SDK.
  * in configurations specifying both `id` and `type`+`target` in recipient blocks, the suggestion is to just use `id` going forward.
* the migration has introduced a new bug (#309) affecting only PagerDuty recipients where the default notification severity of `critical` was being relied upon without specifying a `notification_details` block.
  * we felt that the benefit of these fixes outweighted the impact of this newly introduced bug
  * the bug has a very straight forward work around (just specify the severity!), documented in the issue (#309)

FEATURES:

* *New Datasource*: `honeycombio_column` (#297)
* *New Datasource*: `honeycombio_columns` (#297)

ENHANCEMENTS:

* resource/honeycombio_trigger: add `evaluation_schedule` support (#314)

BUGFIXES:

* client - escape query string when listing burn alerts for an SLO (#301)
* resource/honeycombio_trigger: recipient fixes (#306, #311)

HOUSEKEEPING:

* build(ci): appease the linter gods (#296)
* build(deps): bump codecov/codecov-action from 3.1.2 to 3.1.4 (#298, #307)
* build(deps): bump github.com/stretchr/testify from 1.8.2 to 1.8.4 (#308, #312)
* build(deps): introduce TF Plugin Framework (#305)

# 0.14.0 (Apr 19, 2023)

ENHANCEMENTS:

* resource/honeycombio_board: add overlaid charts `graph_settings` support (#291)
* resource/honeycombio_dataset: add resource import support (#294)

BUGFIXES:

* client: fix marker client test flake (#293) 

HOUSEKEEPING:

* build(deps): Bump actions/setup-go from 3 to 4 (#286)
* build(deps): Bump codecov/codecov-action from 3.1.1 to 3.1.2 (#289)
* build(deps): bump Go to 1.19 (#292)
* build(deps): Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.25.0 to 2.26.1 (#288)

# 0.13.1 (Mar 3, 2023)

BUGFIXES:

* docs: correct dataset in `honeycombio_query` example (#270)
* docs: correct `time_range` in `honeycombio_trigger` example (#272)
* docs: mention SLI's in `__all__` in SLO dataset docs (#269)
* datasource/honeycombio_query_specfication: fix query equivalence for time_range, filter ops, and calculations (#282)

HOUSEKEEPING:

* build(deps): Bump goreleaser/goreleaser-action from 4.1.0 to 4.2.0 (#264, #265)
* build(deps): Bump github.com/joho/godotenv from 1.4.0 to 1.5.1 (#274)
* build(deps): Bump honeycombio/gha-create-asana-task from 1.0.0 to 1.0.1 (#276)
* build(ci): Fix Asana Task creation (#277, #279)
* build(deps): Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.24.1 to 2.25.0 (#281)
* build(deps): Bump golang.org/x/net from 0.6.0 to 0.7.0 (#283)
* build(deps): Bump github.com/stretchr/testify from 1.8.1 to 1.8.2 (#284)

# 0.13.0 (Jan 18, 2023)

NOTES: The `honeycombio_column` resource will now *delete* dataset columns on destroy.
Deletes are a destructive and irreversible operation.
Prior to this release, column destroys were a 'noop' leaving the column untouched in the dataset.

ENHANCEMENTS:

* resource/honeycombio_column: delete column on destroy (#258)

HOUSEKEEPING:

* build(ci) - add repo name and link to Asana task (#257)

# 0.12.0 (Dec 16, 2022)

NOTES:

* `honeycombio_column` resource's argument `key_name` has been deprecated in favor of `name` and will be removed in a future release of the provider.
* `honeycombio_column` no longer silently imports an existing column on create.

ENHANCEMENTS:

* *New Resource*: `honeycombio_dataset_definitions` (#217)
* resource/honeycombio_column: deprecate 'key_name' in favor of 'name' (#242)
* resource/honeycombio_board: add new 'board_url' attribute (#254)

BUGFIXES:

* datasource/honeycombio_recipient: fix bug where only supplying 'type' would error (#240)
* resource/honeycombio_column: no longer silently imports an existing column on create (#242)
* docs: fix 'recipients' misspellings (#246)
* docs: remove deprecated 'dataset' from simple board example (#249)
* resource/honeycombio_board: fix panic on board graph settings parsing (#250)
* docs: fix 'README' markdown rendering (#253)

HOUSEKEEPING:

* build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.24.0 to 2.24.1 (#239)
* add missing OSS Lifecycle badge (#243)
* build(ci): send GitHub issues and PRs to Asana (#244, #251)
* build(deps): Bump goreleaser/goreleaser-action from 3.2.0 to 4.1.0 (#252)

# 0.11.2 (Nov 7, 2022)

BUGFIXES:

* datasource/honeycombio_query_specfication: missing 'calculation' can cause infinite diff (#234)
* resource/honeycombio_column: missing `type` can cause infinite diff (#235) 
* datasource/honeycombio_query_specification: suppress 'equivalent' Query Specification diffs (#236)

HOUSEKEEPING:

* build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.23.0 to 2.24.0 (#228)
* build(deps): bump goreleaser/goreleaser-action from 3.1.0 to 3.2.0 (#231)
* build(deps): bump github.com/stretchr/testify from 1.8.0 to 1.8.1 (#232)

# 0.11.1 (Oct 14, 2022)

BUGFIXES:

* resource/honeycombio_dataset - properly update `description` and `expand_json_depth` attributes (#229)

# 0.11.0 (Oct 5, 2022)

ENHANCEMENTS:

* resource/honeycombio_marker_setting: support for `marker_setting` (#224)

BUGFIXES:

* docs: add clarifying note about Trigger time_range vs frequency (#219)

HOUSEKEEPING:

* build(deps): bump goreleaser/goreleaser-action from 3.0.0 to 3.1.0 (#218)
* build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 (#223)
* build(deps): bump codecov/codecov-action from 3.1.0 to 3.1.1 (#225)
* Update CODEOWNERS (#220)

# 0.10.0 (Aug 16, 2022)

ENHANCEMENTS:

* resource/honeycombio_board: support for `column_layout` (single vs multi) (#210)
* resource/honeycombio_board: board queries' graph settings (Omit Missing Values, Use UTC X-Axis, et cetera) are now configurable via the optional `graph_settings` block (#211)

BUGFIXES:

* resource/honeycombio_pagerduty_recipient: `integration_key` is now marked "sensitive" (#213)

HOUSEKEEPING:

* Go upgraded from 1.17.13 to 1.18.5 (#215)
* terraform-plugin-sdk upgraded from 2.20.0 to 2.21.0 (#214)

# 0.9.0 (Aug 10, 2022)

NOTES:

* `honeycombio_board` no longer requires specifying `dataset` for each `query` object. `dataset` has been marked deprecated.

FEATURES:

* the provider's configuration now respects the more standard `HONEYCOMB_API_KEY` in addition to `HONEYCOMBIO_APIKEY` (#187, #208)

ENHANCEMENTS:

* resource/honeycombio_dataset: support for `description` and `expand_json_depth` arguments (#185)
* resource/honeycombio_dataset: addition of `created_at` and `last_written_at` attributes (#204)
* resource/honeycombio_column: addition of `created_at`, `last_written_at`, and `updated_at` attributes (#198)
* resource/honeycombio_board: environment-wide queries are now supported. `dataset` is no longer required as part of the board `query` definition and has been marked deprecated. (#203)

BUGFIXES:

* docs: typos and corrections (#199, #206)
* resource/honeycombio_trigger: fix recipients ordering and potential infinite diff for pagerduty recipients (#202)

HOUSEKEEPING:

* CI: workflow improvements and scheduled nightly 'smoketest' runs (#195)
* terraform-plugin-sdk upgraded from 2.19.0 to 2.20.0 (#201)
* Go upgraded from 1.17.11 to 1.17.13 (#207)

# 0.8.0 (July 22, 2022)

NOTES:

* client: support for `zenoss` recipient type removed (#190)
  * this was never available at the Terraform resource level
* `honeycombio_recipient` will now fail if your query returns more than one recipient. Before it just picked the first one returned by the API.

FEATURES:

* *New Resource*: `honeycombio_email_recipient` (#186)
* *New Resource*: `honeycombio_pagerduty_recipient` (#188)
* *New Resource*: `honeycombio_slack_recipient` (#188)
* *New Resource*: `honeycombio_webhook_recipient` (#188)
* *New Data Source*: `honeycombio_recipients` (#188)

ENHANCEMENTS:

* client: error details from the API are now displayed in Terraform errors (#184)
* datasource/honeycombio_recipient: - now uses the [Recipients API](https://docs.honeycomb.io/api/recipients/) and can filter recipient types with an optional `detail_filter` (#188)
  * `dataset` is now ignored and marked as a deprecated argument
  * `target` contines to work but is now deprecated
  * `detail_filter` improves the experience of selecting the _correct_ PagerDuty recipient you are looking for.
* resource/honeycombio_trigger and resource/honeycombio_burn_alert - notification severity can now be specified when a Trigger or a Burn Alert fires (#191)

BUGFIXES:

* docs: syntax and correctness updates (#176, #180)
* resource/honeycombio_trigger - correct Trigger query test schema (#177)

HOUSEKEEPING:

* terraform-plugin-sdk upgraded from 2.16.0 to 2.19.0 (#175, #183, #189)
* testify upgraded from 1.7.1 to 1.8.0 (#178, #181, #182)
* CI: goreleaser-action bumped from 2.9.1 to 3.0.0 (#168)

# 0.7.0 (Jun 2, 2022)

NOTES:

* `honeycombio_trigger_recipient` data source has been deprecated in favour of the more generic `honeycombio_recipient`.
The deprecated data source will be removed in a future release.

FEATURES:

* *New Resource*: `honeycombio_slo` (#166)
* *New Resource*: `honeycombio_burn_alert` (#166)
* *New Data Source*: `honeycombio_recipient` (#166)

ENHANCEMENTS:

* resource/honeycombio_trigger: add `alert_type` argument (#159)
* docs: fixes and additional examples (#167, #169)

BREAKING CHANGES:

* `honeycombio_query_result` now takes the Query Specification JSON directly (#165)

HOUSEKEEPING:

* terraform-plugin-sdk upgraded from 2.15.0 to 2.16.0 (#164)

# 0.6.0 (May 9, 2022)

FEATURES:

* *New Data Source*: `honeycombio_query_result` (#151)

HOUSEKEEPING:

* terraform-plugin-sdk upgraded from 2.14.0 to 2.15.0 (#161)

# 0.5.0 (Apr 25, 2022)

BUGFIXES:

* docs: grammar fixes (#153, #152)
* client: ensure Derived Column `alias` is properly URL escaped (#154)
* resource/honeycombio_query_annotation: properly validate length for `name` at 80 characters (#155)

ENHANCEMENTS:

* resource/honeycombio_derived_column: validate length for `alias`, `expression`, and `description` (#154)
* resource/honeycombio_board: validate length for `name`, `description`, and query `caption` (#155)
* resource/honeycombio_column: validate length for `key_name`, and `description` (#155)
* resource/honeycombio_dataset: validate length for `name` (#155)

HOUSEKEEPING:

* terraform-plugin-sdk upgraded from 2.13.0 to 2.14.0 (#149)
* CI: remove unmaintained buildevents action (#150)
* CI: bump Go version to 1.17 (#150)
* CI: hashicorp/setup-terraform action upgraded from 1 to 2 (#157)
* CI: codecov/codecov-action action upgraded from 3.0.0 to 2.1.0 (#156)

# 0.4.0 (Apr 13, 2022)

NOTES:

* A Trigger may need to be destroyed and recreated in order to stabalize the ordering of recipients.

BUGFIXES:

* resource/honeycombio_trigger: fix unstable recipient ordering causing infinite diffs (#142)
* datasource/honeycombio_query_specfication: fix for `filter_combination` 'AND' causing infinite diffs (#144)

ENHANCEMENTS:

* docs: add SLI example (#138)
* validation for Trigger and Board name and description lengths (#143)

HOUSEKEEPING:

* terraform-plugin-sdk upgraded from 2.10.1 to 2.13.0 (#135, #139)
* testify upgraded from 1.7.0 to 1.7.1 (#137)

# 0.3.2 (Mar 9, 2022)

BUGFIXES:

* resource/honeycombio_trigger: workaround for misparsing a recipient's empty 'target' when using dynamic blocks (#132)

ENHANCEMENTS:

* provider can be started in debug mode with support for debuggers like delve (#129)

# 0.3.1 (Mar 4, 2022)

BUGFIXES:

* client: error if creating a derived column with an alias that already exists (#124)

# 0.3.0 (Feb 17, 2022)

NOTES:

* the `value` filter attribute has been *undeprecated* and now properly coerces the input when marshaling JSON to the Honeycomb API.
* the type-specific `value_boolean`, `value_float`, `value_integer` and `value_string` filter values (introduced by #29) have been
deprecated.
The `value_*` filter attributes (introduced by #29) will be removed before the 1.0 release.

ENHANCEMENTS:

* datasource/honeycombio_query_specfication: support for `having` filters (#110)
* datasource/honeycombio_query_specfication: support for `CONCURRENCY` operator (#112)
* docs: handful of fixes and clarifications (#111, #115, #116)

BUGFIXES:

* datasource/honeycombio_query_specfication: filtering by the 'zero value' of a type and properly coerced values now sent to the API. Filter `value` has been undeprecated and the `value_*` have been deprecated (#114)
* datasource/honeycombio_query_specfication: specifiying `ascending` sort order no longer causes constant diffs (#120)

# 0.2.0 (Jan 27, 2022)

NOTES:

* This is the first official release made by Honeycomb!
* This release does contain three breaking changes, see below.

FEATURES:

* resource/honeycombio_board: board queries now support annotations (#100)

ENHANCEMENTS:

* client: API client is no longer a third party dependency (#88)
* client: query specification support for RATE_AVG, RATE_MAX, and RATE_SUM (#92)

BUGFIXES:

* resource/honeycombio_trigger: `disabled` properly marshaled to allow disabling triggers (#91)
* resource/honeycombio_query: Suppress equivalent `query_json` differences (#100)
* documentation fixes (#94, #99)

BREAKING CHANGES:

* datasource/honeycombio_query: renamed to `datasource/honeycombio_query_specification` (#98)
* resource/honeycombio_board: board queries no longer support inline query JSON (#96)
* resource/honeycombio_trigger: triggers no longer support inline query JSON (#96)
