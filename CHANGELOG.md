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
