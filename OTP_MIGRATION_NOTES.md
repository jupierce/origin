# OpenShift Tests Private (OTP) to Origin Migration Notes

This document provides comprehensive guidance for migrating openshift-tests-private (OTP) from using its internal `github.com/openshift/openshift-tests-private/test/extended/util` package to the external `github.com/openshift/origin/test/extended/util` package or a package the migration will create called `github.com/openshift/origin/test/extended/compat_otp`. The migration must preserve all OTP behavior while consolidating utility functions. Keep notes about your progress in the origin project under /migration-progress.log so that you can reference to pick back up on the process later. Record both that you started and that you completed a step so that partially completed steps can be detected.

Steps in this migration:
1. Update origin to use go1.24.0 and ensure it can `make clean build` before proceeding. 
2. Add an import for github.com/openshift/origin to otp's go.mod. For the duration of the migration, a replace statement should replace this module with the local ../origin. This will allow us to make changes in the local repos during the migration and allow otp to compile with the locally adjusted origin.
3. Copy openshift-tests-private/test/extended/util/* to a new package created for the migration origin/test/extended/util/compat_otp/. These are the otp implementations whose behavior we need to preserve while still eliminating duplication. By copying all files and child packages to compat_otp, we have a starting point for the migration.
4. At this point, work to get origin to a state where "make clean build" succeeds. This will involve pulling in modules that otp used and increasing origin's go.mod to use go1.24.0. Keep origin's module versions unless an increase is necessary. When otp and origin go.mod versions conflict, choose the greater of the two. 
5. Once origin can be builds successfully, create a new comment and branch (named compat_otp_5) to lock in this progress.
6. Now compare the CLI interfaces in test/extended/util/cli.go and test/extended/util/compat_otp/cli.go (include other go files in this comparison if they add functions to the interface). The compat_otp CLI will possess functions that are not present in util.CLI. Create a new file in origin: test/extended/util/util_otp.go. This file will be used exclusively for functions that MUST go into the origin util package. Most noteably, functions that extend the CLI interface. Functions in this file should be minimal. Do not be tempted to reexport functions from compat_otp to limit changes in other source files. Instead, those source files should be changed to refer to compat_otp functions and packages directly. Again, only place content in util_otp.go if it needs special access to the internals of origin's util package.
7. origin's util.CLI should now offer a drop in replacement for compat_otp.CLI. For all files in compat_otp package and subpackages that use the compat_otp.CLI, replace it with util.CLI.
8. Without undoing any progress, get origin back to a buildable state using only util.CLI. Create a new commit and branch (named compat_otp_8).
9. After the commit, eliminate the use of functions from compat_otp files and child packages that already exist in origin's util/ package. For example util/db and compat_otp/db are likely identical. In this case, files under compat_otp should be updated to use these files from github.com/openshift/origin/test/extended/util/db instead of compat_otp/db. This will leave functions that are no longer being used in compat_otp. That can be addressed later.
10. Perform this adjustment package by package, returning origin to a buildable state whenever possible before proceeding to the next step. When complete, create a new origin commit and branch (named compat_otp_10) to lock in the progress
11. Next, rename compat_util.CLI to compat_util.CLI_do_not_use across. This rename should not propagate to many files in compat_otp, because they should already be using util.CLI. From this point on, it is there is no reason for anything to use compat_util.CLI_do_not_use.
12. Once origin is compileable again, create a new commit and branch (named compat_otp_12) to lock in the progress.
13. It is now time to focus on github.com/openshift/openshift-tests-private (otp). otp's test/extended/util should be moved to the root of the otp project to a directly named otp_util_orig. This will serve as a reference for the state of the otp utilities prior to the migration, but should no longer be used for compilation. Firstly, replace all references to github.com/openshift/openshift-tests-private/test/extended/util to be github.com/openshift/origin/test/extended/util. Since origin's CLI has been updated to include any missing functions otp might expect, origin's util.CLI should be a drop-in replacement. Secondly, update other otp source files that require functions and packages under origin's new compat_otp package to import and use those packages directly from origin. Iterate on building otp, fixing failures according to the principles set forth in this document, until it builds successfully.

**Critical Rules**: 
- When a function is very similar between otp and origin, attempt to use origin's implementation and make any small adjustments necessary to otp's invocation of the function. For example, if origin's signature requires a Context, provide one. If origin's signature returns (string, error), while otp only returns (string), alter otp to handle the error and reuse origin's implementation. 
- Currently, otp imports origin, but does a local replaces in its go.mod. This should continue so that the migration can update both directory structures and test immediately.
- Don't enabling vendor/ mode in otp. 
- It should not be necessary to add new modules to origin's go.mod that don't also exists in otp's go.mod. Versions can change, of course, but there should be no need to add new SDKs, for example. The implemention copied from otp is what we want to get working with minimal changes.


### Background
- OTP originally copied origin's util package and both have evolved independently
- This has led to duplicate code and maintenance burden
- OTP has added many specialized functions not present in origin

### Goal
Make origin provide a complete replacement for OTP's util packages without changing any OTP test behavior. While behavior must not change, this is not a drop in replacement, as files in OTP will be updated in hundreds of locations to refer to github.com/openshift/origin/test/extended/util/compat_otp .

### Constraints
1. OTP behavior must not change
2. All implementations in origin must be real (no stubs)
3. The migration should be performed in careful incremental steps. Between significant changes, ensure that "make clean build" still works in otp and origin.
4. Both migrated projects should use go 1.24.0.

**Migration Patterns Established:**
1. Use `*OTP` suffix for functions with irreconcilably different signatures in util_otp.go
3. Use compat_otp package for complex implementations

