# Naming Rules

## 1. Core naming principle

Every coined project-wide name must begin with an object name.

A name must not begin with a verb.

Wrong:

```text
LoadSettings
CheckLibrary
PrepareWorkspace
ValidateSource
```

Correct:

```text
LSettingsLoad
LLibraryCheck
LWorkspacePrepare
LSourceValidate
```

The first element of the name must identify what object the name belongs to.

---

## 2. Object prefixes

Every project object must have one of these prefixes:

```text
P = Presentation object
L = Logic object
```

### 2.1 P objects

`P` means the object refers to a real element, region, control, or visible concept shown in the UI.

Examples:

```text
PSource
PPrep
PBuild
PLibrary
PSettings
PWorkspace
PLog
PDropdown
PButton
```

Use `P` only for real UI/presentation objects.

### 2.2 L objects

`L` means the object does not refer to a real UI element.

Use `L` for backend, logic, domain, state, validation, planning, file handling, command execution, metadata, and build behavior.

Examples:

```text
LSource
LArchive
LSignature
LRelease
LVersion
LLibrary
LPreset
LPlanner
LPreparation
LPackage
LBuild
LCommand
LProcess
LLog
```

### 2.3 Prefixes are part of the object name

The prefix is not decorative.

`PSource` and `LSource` are different objects.

```text
PSource = visible Source UI area
LSource = backend source/archive logic
```

They may share the same noun only because their prefixes make them different project objects.

### 2.4 Code spelling and language visibility

The prefix concept is `P` / `L`, but the literal code spelling may use lowercase `p` or `l` when the programming language gives uppercase names special meaning.

In Go, uppercase identifiers are exported. Therefore, a private Go helper must not be made public merely to satisfy the visual `P` / `L` style.

Correct for public/exported names:

```text
LProgram
LPlanFFmpegRequest
PProgramRender
```

Correct for private Go helpers:

```text
lFFmpegBuild
lLogEmit
lWindowGeometrySave
```

Wrong:

```text
LFFmpegBuild  = wrong if this is meant to be a private helper on a Wails-bound type
LLogEmit      = wrong if this is meant to be private backend infrastructure
```

The naming rule concerns object correspondence. It does not require accidentally exporting private code.

---

## 3. Object registry

Every project object must be recorded and maintained in:

```text
/docs/internal/objects.json
```

An object must satisfy all of these rules:

```text
1. It must be unique throughout the project.
2. It must be a single word after the prefix.
3. It must be a noun.
4. Its canonical registry form must have either P or L as its prefix.
5. Its role must be clearly documented.
6. Code may use lowercase p or l only when language visibility rules require it, such as private Go helpers.
```

Examples of valid objects:

```text
PSource
LSource
PBuild
LBuild
PLibrary
LLibrary
LArchive
LSignature
LCommand
```

Examples of invalid objects:

```text
BasicPreparation
CurrentNativeLibrary
BuildManager
LibraryHandler
SourceHelper
```

Reasons:

```text
BasicPreparation      = two-word object
CurrentNativeLibrary = stacked modifiers
BuildManager         = vague role
LibraryHandler       = vague role
SourceHelper         = vague role
```

---

## 4. Normal project-wide name grammar

Normal coined project-wide names must follow this form:

```text
Object + optional single modifier + bare verb
```

Examples:

```text
LSettingsLoad
LLibraryNativeCheck
LArchiveExtract
LSignatureVerify
LWorkspaceClean
LCommandRun
PSourceRender
PWorkspaceBrowse
PLogAppend
PLibraryToggle
```

A name may have one modifying word, and that modifier must follow the object.

Correct:

```text
LLibraryNativeCheck
LVersionCurrentSelect
PButtonPrimaryRender
```

Wrong:

```text
NativeLibraryCheck
CurrentVersionSelect
PrimaryButtonRender
LPreparationBasicFundamentalRun
```

---

## 5. Method naming rule

A method name must end with a bare verb.

Correct:

```text
LArchiveExtract
LSignatureVerify
LLibraryResolve
LPresetNormalize
LCommandRun
PSourceRender
PLogAppend
```

Wrong:

```text
LArchiveExtraction
LSignatureVerified
LLibraryResolving
LPresetNormalized
LCommandRunner
PSourceRendered
```

The final word must describe the action in bare verb form.

---

## 6. Verb-first names are forbidden

A coined project-wide name must not begin with an action.

Wrong:

```text
LoadSettings
VerifySignature
ExtractArchive
RunCommand
RenderSource
AppendLog
```

Correct:

```text
LSettingsLoad
LSignatureVerify
LArchiveExtract
LCommandRun
PSourceRender
PLogAppend
```

---

## 7. Modifier rule

A project-wide name may contain only one modifying word.

The modifier must follow the object it modifies.

Correct:

```text
LLibraryNativeCheck
LVersionCurrentSelect
PButtonPrimaryRender
```

Wrong:

```text
NativeLibraryCheck
CurrentVersionSelect
PrimaryButtonRender
LLibraryCurrentNativeCheck
LPreparationBasicFundamentalRun
```

The normal shape is:

```text
Object + Modifier + Verb
```

not:

```text
Modifier + Object + Verb
```

and not:

```text
Object + Modifier + Modifier + Verb
```

---

## 8. XInstance

`XInstance` refers to all events and methods that run because of one user action.

Example:

```text
The user opens the program.
Everything that runs from program start until the system becomes idle again belongs to one XInstance.
```

Another example:

```text
The user presses Build.
Everything that runs from that click until the build-start chain becomes idle belongs to one XInstance.
```

---

## 9. Unique subchain

If a method or event is used only inside one XInstance, it is that XInstance's unique subchain.

A unique subchain may receive a longer name that identifies both:

```text
1. the owning XInstance
2. the subchain object/action
```

---

## 10. XInstance unique-subchain naming grammar

Normal names use:

```text
Object + optional Modifier + Verb
```

XInstance unique-subchain names may use two object-verb pairs:

```text
InstanceObject + InstanceVerb + SubchainObject + SubchainVerb
```

A delimiter is recommended for readability:

```text
InstanceObjectInstanceVerb_SubchainObjectSubchainVerb
```

Examples:

```text
LSettingsLoad_LLibraryCheck
PSourceSelect_LArchiveVerify
PBuildClick_LBuildRun
LProgramStart_LWorkspacePrepare
```

This means:

```text
LSettingsLoad_LLibraryCheck
= the LLibraryCheck subchain exists only inside the LSettingsLoad instance
```

XInstance names should not be used casually. They are allowed only when the subchain is truly unique to that one XInstance.

---

## 11. XInstance registry

XInstance chains should be recorded separately from objects.

Recommended file:

```text
/docs/internal/chains.json
```

The document should record:

```text
1. XInstance name
2. Triggering user action
3. Unique subchains
4. Notes about reuse restrictions
```

Example:

```json
{
  "PBuildClick": {
    "trigger": "The user presses the Build button.",
    "uniqueSubchains": [
      "PBuildClick_LBuildRun",
      "PBuildClick_LLogStream"
    ],
    "notes": "These subchains must not be reused by SourceSelect or SettingsLoad."
  }
}
```

---

## 12. Maximum name length

A normal method name should contain:

```text
one object + optional one modifier + one bare verb
```

An XInstance unique-subchain name may contain:

```text
two object-verb pairs
```

Anything longer is suspicious and should usually be split.

Suspicious:

```text
LProgramStart_LSettingsLoad_LLibraryCompatibilityNormalize
```

Better:

```text
LProgramStart_LSettingsLoad
LSettingsLoad_LLibraryNormalize
```

or, if reusable:

```text
LLibraryNormalize
```

---

## 13. Internal-name exception

Names used only inside a single method/function body are internal names.

Internal names are exempt from the project-wide object-prefix rule.

They do not need to:

```text
1. start with P or L;
2. start with a registered project object;
3. be recorded in /docs/internal/objects.json.
```

Allowed internal-name shapes:

```text
noun
noun + one modifier
```

Examples:

```text
path
archive
version
result
message
command

pathSource
archiveCurrent
versionSelected
resultFinal
messageError
commandBuild
```

Even for internal names, noun-first order is preferred.

Preferred:

```text
versionSelected
archiveCurrent
resultFinal
messageError
commandBuild
```

Less preferred:

```text
selectedVersion
currentArchive
finalResult
errorMessage
buildCommand
```

---

## 14. Boundary of internal names

The internal-name exception applies only inside one method/function body.

It does not apply to:

```text
types
interfaces
classes
structs
constants
file-level variables
package-level variables
component state
events
exported names
function parameters that form an API contract
non-serialized fields
non-serialized properties
```

Once a name escapes a single method body, it becomes a project-wide name and must follow the full project naming rules, unless another explicit exception below applies.

---

## 14-1. Data-contract field and property exception

Serialized data-contract fields and properties are exempt from the object-prefix rule when their names must remain stable for JSON, Wails, frontend model, saved-state, manifest, or external data compatibility.

This exception applies to record-shape names such as:

```text
actionName
workspaceDirectory
selectedLibraryIds
ffmpegVersion
displayName
categoryName
```

These names are data fields, not project objects or methods. Their surrounding type, function, constant, and file-level names must still follow the P/L object naming rules.

Non-serialized implementation fields are not covered by this exception and should follow the project naming rules unless they are internal names used only inside one method body.

---

## 14-2. Package and folder-name exception

These naming rules do not apply to package names, folder names, or source-file path names.

Package and folder names may follow language, framework, and repository conventions such as lowercase Go package names. They should still correspond to registered project objects where practical, but they do not need to start with P or L.

---

## 15. Localization-key exception

These naming rules do not apply to localization key names.

Localization keys may follow the structure required by the localization system, translation files, or UI text grouping.

Examples:

```text
source.workspace.location
prep.status.ready
build.button.start
libraries.category.video
settings.language.korean
```

Localization keys are text-resource identifiers, not project object names.

Therefore, localization keys:

```text
1. do not need to start with P or L;
2. do not need to start with a registered object name;
3. do not need to end with a bare verb;
4. do not need to be recorded in /docs/internal/objects.json;
5. are exempt from the one-modifier rule.
```

However, localization keys should still remain consistent, readable, and grouped according to the visible UI or translation domain they belong to.

---

## 16. UI event naming

A user action is not itself a UI object.

Do not create objects such as:

```text
PClick
PSelect
PToggle
```

Instead, use the UI object plus a bare verb:

```text
PBuildClick
PSourceSelect
PLibraryToggle
```

For a chain triggered by such an action:

```text
PBuildClick_LBuildRun
PSourceSelect_LArchiveVerify
PLibraryToggle_LLibraryResolve
```

---

## 17. Avoid vague object names

Avoid vague structural names unless they have a clearly defined and unique project role.

Usually forbidden or suspicious:

```text
Manager
Handler
Helper
Service
Data
Info
Processor
Controller
Utility
Common
Base
Core
```

Better names should identify the actual project object:

```text
LArchive
LSignature
LLibrary
LPlanner
LCommand
LPreparation
LRelease
LVersion
LWorkspace
```

---

## 18. Summary

The project uses two naming levels.

### Project-wide names

Project-wide names must follow strict object-first rules.

```text
P/L Object + optional one modifier + bare verb
```

Examples:

```text
PSourceRender
PLogAppend
LArchiveExtract
LSignatureVerify
LLibraryResolve
LCommandRun
```

### Internal names

Internal names are used only inside one method body.

They may be simple and local.

Examples:

```text
path
archive
versionSelected
resultFinal
messageError
commandBuild
```

### XInstance unique subchains

XInstance unique subchains are used only when a chain belongs uniquely to one user-action instance.

```text
PBuildClick_LBuildRun
PSourceSelect_LArchiveVerify
LSettingsLoad_LLibraryCheck
```

### Localization keys

Localization key names are exempt from these naming rules.


## Go test entry-function exception

Go test entry functions may follow Go's required `TestXxx` naming convention.

This exception applies only to functions that must be discovered by Go's test runner. Test helper functions, test file-level variables, and test-local support types should still follow the project naming rules unless they are internal names used only inside one test function body.


## Generated binding exception

Generated Wails binding/runtime files are not manually authoritative naming sources.

When generated files expose an old name, the source name should be corrected first and the generated bindings should then be regenerated. Manual edits to generated files are allowed only as a temporary synchronization step when the generator cannot be run in the current environment.
