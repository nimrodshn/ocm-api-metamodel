# Changes

This document describes the relevant changes between releases of the
API metamodel.

## 0.0.44 Nov 22 2021

- Check loops in locator paths.
- Add `Empty` method to builders.

## 0.0.43 Nov 15 2021

- Add `status` attribute to errors.

## 0.0.42 Oct 14 2021

- Accept iterator as parameter in `helpers.NewIterator`.

## 0.0.41 Sep 24 2021

The only change in this release is that the _GitHub_ action that publishes
releases has been fixed so that it publishes correct binaries. There are no
changes in functionality.

## 0.0.40 Aug 27 2021

This release doesn't contain any functional changes, it only contains changes
in the build process intended to automatically publish the release binaries so
that other projects can use them without having to build the project.

## 0.0.39 Aug 23 2021

- Add `details` attribute to errors.

## 0.0.38 Jul 6 2021

- Explicitly import `github.com/golang/glog` to avoid conflicts with
  `github.com/istio/glog`.

## 0.0.37 Feb 15 2021

- Add metrics support generator, and remove `X-Metric` header.

## 0.0.36 Feb 8 2021

- Use Go 1.15
- Add `documentedSupport` and `namedSupport`
- Add `typedSupport`
- Make reporter streams configurable
- Add presence bitmap

## 0.0.35 Nov 17 2020

- Update to version 4.8 of Antlr
- Wrap errors

## 0.0.34 Oct 5 2020

- names: Support numeric initialisms

## 0.0.33 Sep 30 2020

- json: Support NoContent on POST responses

## 0.0.32 Aug 23 2020

- Add search method

## 0.0.31 Jul 30 2020

- Adding List type to checkUpdate validator

## 0.0.30 Jun 28 2020

- Add Interface type to generator

## 0.0.29 Jun 9 2020

- pr_check: Lock in dependency versions for test pipeline
- Fix setter for Poll request params

## 0.0.28 May 13 2020

- OpenAPI: Fix expected response

## 0.0.27 Apr 7 2020

- Update file header year to 2020

## 0.0.26 Feb 26 2020

- Add `operation_id` attribute to error objects and error messages.

  An error object like this:

  ```json
  {
    "kind": "Error",
    "id": "401",
    "href": "/api/clusters_mgmt/v1/errors/401",
    "code": "CLUSTERS-MGMT-401",
    "reason": "My reason",
    "operation_id": "456"
  }
  ```

  Will result in the following error string (in one single line):

  ```
  identifier is '401', code is 'CLUSTERS-MGMT-401' and
  operation identifier is '456': My reason
  ```

## 0.0.25 Feb 20 2020

- Run the `gofmt` command only once for all generated files instead of running
  it once per each generated file.
- Avoid generating code with constructs that would then be simplified by the
  `-s` flag of the `gofmt` command.

## 0.0.24 Feb 14 2020

- Add `Content-Type` to responses sent by the generated server code.
- Don't require developer to explicitly remove the `/api` when using the
  server code.
- Remove redundant quotes from error responses sent by the generated
  server code.

## 0.0.23 Feb 12 2020

- Fix missing _OpenAPI_ paths due to incorrect use of `append`.
- Move code generators to separate packages: one per language.

## 0.0.22 Jan 9 2020

- Fix generation of _OpenAPI_ paths so that all the characters are lower case.

## 0.0.21 Jan 8 2020

- Use JSON iterator instead of the default JSON Go package.

## 0.0.20 Dec 18 2019

- Fix conversion of errors to JSON so that the `kind` attribute is generated
  correctly.

## 0.0.19 Dec 12 2019

- Don't fail on wrong kind.

## 0.0.18 Nov 25 2019

- Add stage URL and `securitySchemes` to the generated _OpenAPI_
  specifications.

## 0.0.17 Nov 23 2019

- Add semantic checks.
- Add support for default values.
- Check default values of paging parameters.

## 0.0.16 Nov 19 2019

- Add simple conversion from AsciiDoc to Markdown.

## 0.0.15 Nov 19 2019

- Add support for the version metadata resource.

## 0.0.14 Nov 17 2019

- Add `Poll` method to clients that have a `Get` method.

## 0.0.13 Nov 14 2019

- Fix imports of `helpers` and `errors` packages.

## 0.0.12 Nov 4 2019

- Add _OpenAPI_ specification generator.

## 0.0.11 Oct 27 2019

- Improve parsing of initialisms.
- Fix the method not allowed code.
- Send not found when server returns `nil` target.
- Generate service and version servers.
- Don't generate files with execution permission.

## 0.0.10 Oct 25 2019

- Make HTTP server adapters stateless.

## 0.0.9 Oct 15 2019

- Generate shorter adapter names.
- Use constants from the `http` package.
- Shorter _read_ and _write_ names.
- Rename `SetStatusCode` to `Status`.
- Improve naming of variables.
- Set default status.
- Move errors and helpers generators to separate files.

## 0.0.8 Oct 12 2019

- Use a private model for tests.
- Improve support for maps of objects.

## 0.0.7 Sep 13 2019

- Keep concepts sorted by name.
- Don't generate empty `const` block for errors.
- Add `Copy` method to builders.

## 0.0.6 Sep 12 2019

- Explicitly enable Go modules so that the build works correctly when the
  project is located inside the Go path.

## 0.0.5 Sep 10 2019

- Fix generation of field names for query parameters.
- Remove `query` and `path` fields from request objects.
- Remove unused imports.

## 0.0.4 Sep 03 2019

- Generated servers parse request query arguments.

## 0.0.3 Aug 27 2019

- Don't install binaries.

## 0.0.2 Aug 27 2019

- Added new `check` command that loads and checks the model but doesn't
  generate any code.

## 0.0.1 Aug 23 2019

- Initial release.