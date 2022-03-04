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
